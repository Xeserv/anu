package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/facebookgo/flagenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/tigrisdata/anu/internal"
	registry "github.com/tigrisdata/anu/registryauth"
	"github.com/tigrisdata/anu/tigris"
)

var (
	bind       = flag.String("bind", ":5007", "TCP host:port to bind to")
	bucketName = flag.String("bucket-name", "", "bucket to check for access to")
	certFile   = flag.String("cert-fname", "/mnt/certs/RootCA.pem", "certificate file to validate JSON web tokens with")
	keyFile    = flag.String("key-fname", "/mnt/certs/RootCA.key", "key file to sign JSON web tokens with")
	jwtCert    = flag.String("jwt-cert-b64", "", "cert to sign JWTs against (base64 bytes)")
	jwtKey     = flag.String("jwt-key-b64", "", "key to sign JWTs against (base64 bytes)")
	slogLevel  = flag.String("slog-level", "DEBUG", "log level")
)

func main() {
	flagenv.Parse()
	flag.Parse()

	internal.InitSlog(*slogLevel)

	if *bucketName == "" {
		log.Fatal("BUCKET_NAME is not set")
	}

	if *jwtCert != "" {
		certPath, err := writeFile(*jwtCert)
		if err != nil {
			log.Fatalf("can't write certificate file: %v", err)
		}
		defer os.Remove(certPath)
		*certFile = certPath
	}

	if *jwtKey != "" {
		keyPath, err := writeFile(*jwtKey)
		if err != nil {
			log.Fatalf("can't write certificate file: %v", err)
		}
		defer os.Remove(keyPath)
		*keyFile = keyPath
	}

	opt := &registry.Option{
		Certfile:        *certFile,
		Keyfile:         *keyFile,
		TokenExpiration: int64((24 * time.Hour).Seconds()),
		TokenIssuer:     "Tigris Anu",
		Authenticator:   &httpAuthenticator{},
	}

	srv, err := registry.NewAuthServer(opt)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/auth", srv)
	slog.Info("listening", "bind", *bind)
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

func writeFile(body string) (string, error) {
	fout, err := os.CreateTemp("", "anu-*")
	if err != nil {
		return "", err
	}
	defer fout.Close()

	data, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return "", err
	}

	if _, err := fout.Write(data); err != nil {
		return "", err
	}

	return fout.Name(), nil
}
