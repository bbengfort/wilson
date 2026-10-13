# Wilson

**Reads emails to determine if they have transactions that need to be added to a ledger.**

## Purpose

This package is intended as a microservice that uses horizon to orchestrate an agent on behalf of the Bengfort Ledger. Wilson runs an inbound SMTP server that receives emails (e.g. receipts, bank alerts, and purchase confirmations) sent to a dedicated domain, and determines whether they contain financial transactions that should be recorded in the Ledger. Alongside the SMTP listener, Wilson exposes a small HTTP server with Kubernetes-style health probes so it can be deployed as a long-running service.

## Why "Wilson"?

The package is named after [Frank J. Wilson](https://en.wikipedia.org/wiki/Frank_J._Wilson), the famous forensic accountant and IRS agent who analyzed over two million financial records to help convict mobster Al Capone. Like its namesake, Wilson's job is to sift through a large volume of mundane paperwork and find the transactions that matter.

## Quickstart

### Prerequisites

- Go 1.26 or later
- (Optional) Docker, if you prefer to run the container image

### Run from source

Clone the repository and build the binary:

```bash
git clone https://github.com/bbengfort/wilson.git
go run ./cmd/wilson serve
```

Starts the server with the defaults (SMTP on `:25`, probes on `:80`):

Both default ports are privileged on most systems; for local development, bind the SMTP server to an unprivileged port instead.

### Verify the service

Send a test message through the SMTP server using openssl with TLS certs:

```bash
openssl s_client -connect ledger.bengfort.dev:25 -starttls smtp -quiet -crlf
```

Enter the following commands into the terminal:

```bash
EHLO bengfort.com
```

```bash
MAIL FROM:<benjamin@bengfort.com>
```

```bash
RCPT TO:<server@ledger.bengfort.dev>
```

```bash
DATA
Subject: Test Email from the Command Line

This is a test email!

.

```

### Configuration

Wilson is configured entirely through environment variables prefixed with `WILSON_`. A `.env` file in the working directory is loaded automatically on startup. Print the full configuration guide with:

```bash
wilson config
```

| Variable                   | Default      | Description                                                              |
|----------------------------|--------------|--------------------------------------------------------------------------|
| `WILSON_LOG_LEVEL`         | `info`       | Log verbosity: `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic` |
| `WILSON_CONSOLE_LOG`       | `false`      | If `true`, log human-readable text instead of JSON                       |
| `WILSON_BIND_ADDR`         | `:25`        | Address to bind the SMTP server to (e.g. `:25`, `:587`, `:465`)          |
| `WILSON_DOMAIN`            | `localhost`  | Domain name of the server and recipient email addresses                  |
| `WILSON_READ_TIMEOUT`      | `4s`         | Read timeout for the SMTP server                                         |
| `WILSON_WRITE_TIMEOUT`     | `16s`        | Write timeout for the SMTP server                                        |
| `WILSON_MAX_MESSAGE_SIZE`  | `10485760`   | Maximum message size to accept, in bytes                                 |
| `WILSON_MAX_RECIPIENTS`    | `32`         | Maximum number of recipients to accept                                   |
| `WILSON_USE_TLS`           | `false`      | If `true`, the SMTP server requires TLS                                  |
| `WILSON_TLS_CERTS`         |              | Path to the TLS certificate file (required when `WILSON_USE_TLS=true`)   |
| `WILSON_TLS_KEY`           |              | Path to the TLS key file (required when `WILSON_USE_TLS=true`)           |

An example `.env` for a TLS-enabled deployment:

```bash
WILSON_LOG_LEVEL=debug
WILSON_CONSOLE_LOG=true
WILSON_DOMAIN=ledger.example.com
WILSON_USE_TLS=true
WILSON_TLS_CERTS=.secret/tls.crt
WILSON_TLS_KEY=.secret/tls.key
WILSON_READ_TIMEOUT=10m
WILSON_WRITE_TIMEOUT=10m
```

## Development

Run the tests:

```bash
go test ./...
```

