package main

import (
	"context"
	"flag"
	"fmt"
	"log"
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
	certFile   = flag.String("cert-fname", "/mnt/certs/RootCA.pem", "certificate file to validate JSON web tokens with")
	keyFile    = flag.String("key-fname", "/mnt/certs/RootCA.key", "key file to sign JSON web tokens with")
)

func main() {
	flagenv.Parse()
	flag.Parse()

	if *bucketName == "" {
		log.Fatal("BUCKET_NAME is not set")
	}

	opt := &registry.Option{
		Certfile:        *certFile,
		Keyfile:         *keyFile,
		TokenExpiration: time.Now().Add(24 * time.Hour).Unix(),
		TokenIssuer:     "Tigris Anu",
		Authenticator:   &httpAuthenticator{},
	}
	srv, err := registry.NewAuthServer(opt)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/auth", srv)
	log.Println("Server running at ", *bind)
	if err := http.ListenAndServe(*bind, nil); err != nil {
		log.Fatal(err)
	}
}

type httpAuthenticator struct{}

func (h *httpAuthenticator) Authenticate(ctx context.Context, username, password string) error {
	cli, err := tigris.Client(ctx, username, password)
	if err != nil {
		return fmt.Errorf("can't auth: %w", err)
	}

	// HACK(Xe): This really should be in the authz step, but this is a HACK and MUST be fixed before shipping to prod for real
	bucketList, err := cli.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("can't list buckets for auth: %w", err)
	}

	var found bool

	for _, bkt := range bucketList.Buckets {
		if *bkt.Name == *bucketName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("user does not have access to bucket %s", *bucketName)
	}

	return nil
}
