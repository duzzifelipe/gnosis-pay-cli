package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultAPIURL = "https://api.gnosispay.com"
	StateFileName = ".gnosis-pay-state.json"
	GnosisChainID = 100

	// Environment variable names
	EnvPrivateKey = "GNOSIS_PAY_PRIVATE_KEY"
	EnvDomain     = "GNOSIS_PAY_DOMAIN"
	EnvURI        = "GNOSIS_PAY_URI"
)

// State persists between CLI invocations.
type State struct {
	JWT         string `json:"jwt,omitempty"`
	UserJWT     string `json:"userJwt,omitempty"`
	Address     string `json:"address,omitempty"`
	UserID      string `json:"userId,omitempty"`
	SafeAddress string `json:"safeAddress,omitempty"`
	Email       string `json:"email,omitempty"`
}

func statePath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	return filepath.Join(dir, StateFileName), nil
}

// LoadState reads persisted state from the working directory.
func LoadState() (*State, error) {
	p, err := statePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{}, nil
		}
		return nil, fmt.Errorf("read state file: %w", err)
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse state file: %w", err)
	}
	return &s, nil
}

type jwtClaims struct {
	UserID string `json:"userId"`
	Exp    int64  `json:"exp"`
}

func parseJWTClaims(token string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode JWT payload: %w", err)
	}
	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("parse JWT claims: %w", err)
	}
	return &claims, nil
}

func GetAuthToken() (string, *State, error) {
	state, err := LoadState()
	if err != nil {
		return "", nil, fmt.Errorf("load state: %w", err)
	}

	if state.JWT != "" {
		claims, err := parseJWTClaims(state.JWT)
		if err == nil && claims.UserID != "" {
			return state.JWT, state, nil
		}
	}

	if state.UserJWT != "" {
		claims, err := parseJWTClaims(state.UserJWT)
		if err != nil {
			return "", nil, fmt.Errorf("parse userJwt: %w", err)
		}
		if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
			return "", nil, fmt.Errorf("userJwt is expired")
		}
		return state.UserJWT, state, nil
	}

	return "", nil, fmt.Errorf("not authenticated: run 'signup' or 'login' first")
}

// Save persists the state to disk.

// Save writes the state to disk.
func (s *State) Save() error {
	p, err := statePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := os.WriteFile(p, data, 0600); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}
	return nil
}

// PrivateKey reads the hex-encoded private key from the environment.
func PrivateKey() (string, error) {
	key := os.Getenv(EnvPrivateKey)
	if key == "" {
		return "", fmt.Errorf("environment variable %s is not set", EnvPrivateKey)
	}
	// Strip optional 0x prefix
	if len(key) >= 2 && key[:2] == "0x" {
		key = key[2:]
	}
	return key, nil
}

// Domain returns the SIWE domain from environment or a default.
// Must be a real domain, not localhost/127.0.0.1.
func Domain() string {
	domain := os.Getenv(EnvDomain)
	if domain == "" {
		// Default: use a generic placeholder domain
		// Users should set GNOSIS_PAY_DOMAIN to their actual domain
		return "localhost"
	}
	return domain
}

// URI returns the SIWE URI from environment or a default.
// Must match the domain and use https://
func URI() string {
	uri := os.Getenv(EnvURI)
	if uri == "" {
		domain := Domain()
		if domain == "localhost" || domain == "127.0.0.1" {
			return "http://" + domain
		}
		return "https://" + domain
	}
	return uri
}
