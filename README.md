# anu

The Tigris-backed Docker Registry. The name comes from the Sumerian god [Anu (𒀭𒀀𒉡)](https://en.wikipedia.org/wiki/Anu), the god of the sky (or clouds, among other things).

## Deploying Anu

The easiest deployment target for Anu is on [fly.io](https://fly.io), but in theory it should work on any cloud platform. Here's what you need:

- An account on [Tigris](https://console.tigris.dev)
- A [Tigris bucket](https://storage.new) (such as `mybucket`), this is where all your docker images will be stored. This will be called the registry bucket.
- A keypair with Editor permissions on the registry bucket.
- A machine with the following packages installed (names are what you can find in [Homebrew](https://formulae.brew.sh/)):
  - `openssl`
  - `flyctl`
  - `skopeo`
  - The [Docker desktop app](https://www.docker.com/products/docker-desktop/) or a locally installed Docker daemon (on Linux)

Here are the steps:

1. Create two fly.io apps, one for the authentication endpoint and another for the registry.
2. Generate an RSA keypair for signing authentication tokens.
3. Configuring the authentication endpoint app.
4. Configuring the registry app.
5. Deploy the authentication endpoint and registry.
6. Test the registry by loading images into it and running them.

### 1. Create two fly.io apps

Anu needs two apps to work: one for the authentication endpoint and the other for an unmodified Docker registry. Create them like so:

```sh
fly launch --no-deploy
(cd fly/registry && fly launch --no-deploy)
```

Write down the app names for the authentication endpoint and the registry in your notes. This guide will refer to your apps as `anu-auth-endpoint` and `anu-registry` respectively.

### 2. Generating a keypair

Anu signs tokens using an RSA keypair. Generate the keypair in the `certs` directory using the `openssl` command:

```sh
cd certs
openssl req -x509 -nodes -new -sha256 -days 36500 -newkey rsa:4096 -keyout anu.key -out anu.pem -subj "/C=US/CN=Registry Auth CA"
cd ..
```

### 3. Configuring the authentication endpoint app

Set the key and certificate you just generated as base64-encoded bytes:

```sh
fly secrets set -a anu-auth-endpoint \
  JWT_CERT_B64="$(cat certs/anu.pem | base64 -w0)" \
  JWT_KEY_B64="$(cat certs/anu.key | base64 -w0)"
```

Create a secret in the authentication endpoint app pointing to the registry bucket:

```sh
fly secrets set -a anu-auth-endpoint BUCKET_NAME=mybucket
```

### 4. Configuring the registry app

Load the certificate into a fly secret:

```sh
fly secrets set -a anu-registry JWT_CERT_B64="$(cat certs/anu.pem | base64 -w0)
```

Put the auth endpoint URL, bucket name, access key ID, and secret access key into your registry app's secrets:

```sh
fly secrets set -a anu-registry \
  REGISTRY_AUTH_TOKEN_REALM="https://anu-auth-endpoint.fly.dev" \
  REGISTRY_STORAGE_S3_BUCKET=mybucket \
  REGISTRY_STORAGE_S3_ACCESSKEY=${AWS_ACCESS_KEY_ID} \
  REGISTRY_STORAGE_S3_SECRETKEY=${AWS_SECRET_ACCESS_KEY}
```

If you change the URL of the authentication endpoint (such as by adding a [custom domain name](https://fly.io/docs/flyctl/certs-add/)), you will need to change the `REGISTRY_AUTH_TOKEN_REALM` secret to point to the new URL.

### 5. Deploy the authentication endpoint and registry

Deploy both apps:

```sh
fly deploy
(cd fly/registry && fly deploy)
```

If all goes well, you will have two apps online:

- `https://anu-auth-endpoint.fly.dev`
- `https://anu-registry.fly.dev`

### 6. Test the registry

Create a new keypair in the Tigris Dash. Do not give it any permissions to any buckets. This keypair (or any keypair in your account) will be what you use to authenticate to your registry.

```sh
docker login anu-registry.fly.dev -u <access key ID>
```

Then paste your secret access key and hit enter.

Repeat this for `skopeo`:

```sh
skopeo login anu-registry.fly.dev -u <access key ID>
```

Copy the [`hello-world` image](https://hub.docker.com/_/hello-world/) from the Docker Hub to your registry:

```sh
skopeo copy --all docker://hello-world docker://anu-registry.fly.dev/hello-world
```

Wait a moment for everything to be copied over and then try to run it on your local machine:

```sh
docker run --rm -it anu-registry.fly.dev/hello-world
```

This will download the image from your registry and run it, giving you a hello world message. You can repeat this process to authenticate to your other private repositories and copy over your existing images.
