# Secrets Directory

This directory is used to store JWT signing keys for the PDV backend.

## Generating Keys

Use the `auth-keygen` command to generate an Ed25519 key pair:

```bash
go run ./cmd/auth-keygen --kid "2026-07-primary" --dir ./secrets
```

This creates:
- `secrets/2026-07-primary.priv.pem` — Private key (keep secure, never commit)
- `secrets/2026-07-primary.pem` — Public key (safe to commit if needed)

## Configuration

Set the following environment variables:

```bash
JWT_ACTIVE_KEY_ID=2026-07-primary
JWT_PRIVATE_KEY_PATH=./secrets/2026-07-primary.priv.pem
JWT_PUBLIC_KEYS_DIR=./secrets
```

For development with ephemeral keys:

```bash
AUTH_ALLOW_EPHEMERAL_DEV_KEY=true
```

## Security

- Private keys must NEVER be committed to the repository.
- The `.gitignore` excludes `*.pem` files in this directory.
- In production, use a secure key management process.
- Rotate keys following the documented procedure:
  1. Add the new public key alongside the existing one.
  2. Update `JWT_ACTIVE_KEY_ID` to point to the new key.
  3. Wait for the longest access token TTL + clock skew to pass.
  4. Remove the old public key.
