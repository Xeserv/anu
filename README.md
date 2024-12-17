# anu

The Tigris-backed Docker Registry. The name comes from the Sumerian god [Anu (𒀭𒀀𒉡)](https://en.wikipedia.org/wiki/Anu), the god of the sky (or cloud).

## Generating a keypair

```text
openssl req -x509 -nodes -new -sha256 -days 36500 -newkey rsa:4096 -keyout anu.key -out anu.pem -subj "/C=US/CN=Registry Auth CA"
```

### Loading the anu volume with a keypair

- Add this to `fly.toml`:

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
- Copy the contents of `anu.key` (private key) to your clipboard.
- Run `vi /data/anu.key` in that fly machine.
- Paste the private key.
- Save and quit (:wq).
- Exit from the machine.
- Comment out the `cmd` line in fly.toml:

  ```toml
  [experimental]
  #cmd = ["sleep", "inf"]
  ```

- Run `fly deploy`

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
