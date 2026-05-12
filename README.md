# Gnosis Pay CLI

A CLI tool implementing the full Gnosis Pay permissionless integration flow, enabling users to:
generate wallets, authenticate via SIWE, complete KYC, deploy Safes, create cards, and more.

## Prerequisites

- [Go](https://go.dev/dl/) 1.25+
- [mise](https://mise.jdx.dev/) (recommended for toolchain management)

## Setup

```bash
# Clone the repository
git clone https://github.com/duzzifelipe/gnosis-pay.git
cd gnosis-pay

# Install Go via mise (optional)
mise install

# Build the binary
mise run build
# or: go build -o gnosis-pay .

# Verify
./gnosis-pay --help
```

## Configuration

Copy the example environment file and fill in your values:

```bash
cp .env.example .env
```

### Required Environment Variables

| Variable | Description |
|---|---|
| `GNOSIS_PAY_PRIVATE_KEY` | Hex-encoded Ethereum private key (generated with `gnosis-pay wallet generate`) |

### Optional Environment Variables

| Variable | Description | Default |
|---|---|---|
| `GNOSIS_PAY_DOMAIN` | Domain for SIWE message | `localhost` |
| `GNOSIS_PAY_URI` | URI for SIWE message | `http://localhost` |

> **Note:** For production SIWE flows, `GNOSIS_PAY_DOMAIN` must be a real domain (not `localhost`).

## Usage

The complete integration flow consists of the following steps:

```bash
# 1. Generate a wallet
gnosis-pay wallet generate

# 2. Authenticate with SIWE
gnosis-pay auth
```

State is persisted to `.gnosis-pay-state.json` in the working directory.


-----

# Testing

Below there is my first try authenticating into the API using SIWE message.

The private key is leaked since it was generated just for this test.

```bash
 gnosis-pay-cli-clone git:(main) mise build
[build] $ go build -o gnosis-pay .
➜  gnosis-pay-cli-clone git:(main) ./gnosis-pay wallet generate
Address:     0x9b3d9D097141344d487DED59870EaDc959e1F407
Private key: 78e9b95143f19c981135a8fed40174985256e5e032f31938549ba63c8774c8d0

Export and retry auth:
  export GNOSIS_PAY_PRIVATE_KEY=78e9b95143f19c981135a8fed40174985256e5e032f31938549ba63c8774c8d0
➜  gnosis-pay-cli-clone git:(main) export GNOSIS_PAY_PRIVATE_KEY=78e9b95143f19c981135a8fed40174985256e5e032f31938549ba63c8774c8d0
➜  gnosis-pay-cli-clone git:(main) ./gnosis-pay auth
Using SIWE domain: localhost
Using SIWE URI: http://localhost
Wallet address: 0x9b3d9D097141344d487DED59870EaDc959e1F407
Requesting nonce...
Nonce: f1ba288110d73bed6bfa8d67eafbc414509090d3e5a4a194cc20158c914c7f837b089cea52758135bb5fe858cf1fffb4
SIWE message:
---
localhost wants you to sign in with your Ethereum account:
0x9b3d9D097141344d487DED59870EaDc959e1F407

Sign in to Gnosis Pay

URI: http://localhost
Version: 1
Chain ID: 100
Nonce: f1ba288110d73bed6bfa8d67eafbc414509090d3e5a4a194cc20158c914c7f837b089cea52758135bb5fe858cf1fffb4
Issued At: 2026-05-12T11:58:38Z
---
Signing SIWE message...
Submitting authentication challenge...
Challenge response: {"code":"WAFForbidden","message":"forbidden"}
 -> Date: Tue, 12 May 2026 11:58:38 GMT
 -> Content-Length: 45
 -> Content-Type: application/json
 -> Server: awselb/2.0
Error: auth challenge failed (HTTP 403): {"code":"WAFForbidden","message":"forbidden"}
Usage:
  gnosis-pay auth [flags]

Flags:
  -h, --help      help for auth
      --ttl int   JWT time-to-live in seconds (min 3600, max 86400) (default 36000)

auth challenge failed (HTTP 403): {"code":"WAFForbidden","message":"forbidden"}
➜  gnosis-pay-cli-clone git:(main) 
```