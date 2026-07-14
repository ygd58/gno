package ledger

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// Signer implements types.Signer by delegating public key retrieval and
// signing to a Ledger hardware wallet running the Tendermint Validator app.
// Unlike the file/AWS/GCP/Vault-backed signers, the private key never
// leaves the device: PubKey is fetched once at construction time and
// cached, and Sign sends the raw sign-bytes to the device over its
// USB/HID connection for on-device signing.
type Signer struct {
	device ledgerAPI
	hdPath []uint32
	pubKey ed25519.PubKeyEd25519
}

// Signer type implements types.Signer.
var _ types.Signer = (*Signer)(nil)

// PubKey implements types.Signer.
func (s *Signer) PubKey() crypto.PubKey {
	return s.pubKey
}

// Sign implements types.Signer. The message is sent to the Ledger device,
// which displays it for user confirmation before signing.
func (s *Signer) Sign(signBytes []byte) ([]byte, error) {
	sig, err := s.device.SignED25519(s.hdPath, signBytes)
	if err != nil {
		return nil, fmt.Errorf("ledger signer: sign failed: %w", err)
	}

	return sig, nil
}

// Close implements types.Signer. It closes the underlying USB/HID connection.
func (s *Signer) Close() error {
	return s.device.Close()
}

// Signer type implements fmt.Stringer.
var _ fmt.Stringer = (*Signer)(nil)

// String implements fmt.Stringer.
func (s *Signer) String() string {
	return fmt.Sprintf("{Type: LedgerSigner, Addr: %s}", s.pubKey.Address())
}

// Config validation errors.
var (
	errDisabled         = errors.New("ledger signer: not enabled")
	errInvalidPubKeyLen = errors.New("ledger signer: device returned an unexpected public key length")
)

// NewSignerFromConfig connects to a Ledger device running the Tendermint
// Validator app and returns a ready-to-use Signer bound to cfg's HD path.
func NewSignerFromConfig(cfg *Config) (*Signer, error) {
	if !cfg.IsEnabled() {
		return nil, errDisabled
	}

	device, err := newDevice()
	if err != nil {
		return nil, err
	}

	return newSigner(device, cfg.hdPath())
}

// newSigner contains the constructor logic decoupled from the concrete
// Ledger connection, so it can be exercised in tests against a mock
// ledgerAPI without requiring a physical device.
func newSigner(device ledgerAPI, hdPath []uint32) (*Signer, error) {
	raw, err := device.GetPublicKeyED25519(hdPath)
	if err != nil {
		return nil, fmt.Errorf("ledger signer: unable to retrieve public key: %w", err)
	}

	if len(raw) != ed25519.PubKeyEd25519Size {
		return nil, errInvalidPubKeyLen
	}

	var pubKey ed25519.PubKeyEd25519
	copy(pubKey[:], raw)

	return &Signer{device: device, hdPath: hdPath, pubKey: pubKey}, nil
}
