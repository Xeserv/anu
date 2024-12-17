# anu

The Tigris-backed Docker Registry. The name comes from the Sumerian god [Anu (𒀭𒀀𒉡)](https://en.wikipedia.org/wiki/Anu), the god of the sky (or cloud).

## Generating a keypair

```text
openssl req -x509 -nodes -new -sha256 -days 36500 -newkey rsa:4096 -keyout anu.key -out anu.pem -subj "/C=US/CN=Registry Auth CA"
```

### Loading the anu app with a keypair

Set the key and certificate as base64-encoded bytes:

```sh
fly secrets set -a anu JWT_CERT_B64="$(cat certs/anu.pem | base64 -w0)" JWT_KEY_B64="$(cat certs/anu.key | base64 -w0)"
```

Run this before you deploy Anu for the first time.

### Loading the registry volume with a keypair

- Add this to `fly/registry/fly.toml`:

  ```toml
  [experimental]
  cmd = ["sleep", "inf"]
  ```

- Run `fly deploy`.
- Open another terminal and run `fly ssh console -a anu`.
- Copy the contents of `anu.pem` (certificate / public key) to your clipboard.
- Run `vi /data/anu.pem` in that fly machine.
- Paste the certificate.
- Save and quit (:wq).
- Exit from the machine.
- Comment out the `cmd` line in fly.toml:

  ```toml
  [experimental]
  #cmd = ["sleep", "inf"]
  ```

- Run `fly deploy`
