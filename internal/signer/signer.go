package signer

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// Signer wraps an ECDSA private key for Ethereum signing operations.
type Signer struct {
	key *ecdsa.PrivateKey
}

// New creates a Signer from a hex-encoded private key (without 0x prefix).
func New(hexKey string) (*Signer, error) {
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return &Signer{key: key}, nil
}

// Address returns the checksummed Ethereum address derived from the private key.
func (s *Signer) Address() common.Address {
	return crypto.PubkeyToAddress(s.key.PublicKey)
}

// BuildSIWEMessage constructs a SIWE (EIP-4361) message string.
func BuildSIWEMessage(domain, uri, address, nonce string, chainID int) string {
	now := time.Now().UTC().Format(time.RFC3339)

	var b strings.Builder
	fmt.Fprintf(&b, "%s wants you to sign in with your Ethereum account:\n", domain)
	fmt.Fprintf(&b, "%s\n\n", address)
	fmt.Fprintf(&b, "Sign in with Ethereum to the app.\n\n")
	fmt.Fprintf(&b, "URI: %s\n", uri)
	fmt.Fprintf(&b, "Version: 1\n")
	fmt.Fprintf(&b, "Chain ID: %d\n", chainID)
	fmt.Fprintf(&b, "Nonce: %s\n", nonce)
	fmt.Fprintf(&b, "Issued At: %s", now)
	return b.String()
}

// SignMessage signs an arbitrary message using the Ethereum personal_sign scheme:
//
//	keccak256("\x19Ethereum Signed Message:\n" + len(message) + message)
func (s *Signer) SignMessage(message string) (string, error) {
	hash := signHash([]byte(message))
	sig, err := crypto.Sign(hash, s.key)
	if err != nil {
		return "", fmt.Errorf("sign message: %w", err)
	}

	// go-ethereum returns [R || S || V] with V in {0, 1}.
	// Ethereum JSON-RPC expects V in {27, 28}.
	sig[64] += 27

	return hexutil.Encode(sig), nil
}

// signHash computes the Ethereum signed message hash.
func signHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}
