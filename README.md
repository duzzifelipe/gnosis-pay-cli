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
