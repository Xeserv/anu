package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/facebookgo/flagenv"
	registry "github.com/tigrisdata/anu/registryauth"
	"github.com/tigrisdata/anu/tigris"
)

var (
	bind       = flag.String("bind", ":5007", "TCP host:port to bind to")
	bucketName = flag.String("bucket-name", "", "bucket to check for access to")
	certFile   = flag.String("cert-fname", "/mnt/certs/RootCA.crt", "certificate file to validate JSON web tokens with")
	keyFile    = flag.String("key-fname", "/mnt/certs/RootCA.key", "key file to sign JSON web tokens with")
)

func main() {
	flagenv.Parse()
	flag.Parse()

	crt, key := "/mnt/certs/RootCA.crt", "/mnt/certs/RootCA.key"

	opt := &registry.Option{
		Certfile:        "/mnt/certs/RootCA.crt",
		Keyfile:         "/mnt/certs/RootCA.key",
		TokenExpiration: time.Now().Add(24 * time.Hour).Unix(),
		TokenIssuer:     "Authz",
		Authenticator:   &httpAuthenticator{},
	}
	srv, err := registry.NewAuthServer(opt)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/auth", srv)
	log.Println("Server running at ", *bind)
	if err := http.ListenAndServeTLS(*bind, crt, key, nil); err != nil {
		log.Fatal(err)
	}
}

type httpAuthenticator struct{}

func (h *httpAuthenticator) Authenticate(ctx context.Context, username, password string) error {
	cli, err := tigris.Client(ctx, username, password)
	if err != nil {
		return fmt.Errorf("can't auth: %w", err)
	}

	bucketList, err := cli.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		slog.Error("can't list buckets", "accessKeyID", username, "err", err)
		return fmt.Errorf("can't auth: %w", err)
	}

	var found bool

	for _, bkt := range bucketList.Buckets {
		if bkt.Name == *&bucketName {
			found = true
			break
		}
	}

	if !found {
		slog.Error("can't find matching", "accessKeyID", username, "wantBucket", *bucketName)
		return fmt.Errorf("user does not have access to bucket %s", *bucketName)
	}

	return nil
}
