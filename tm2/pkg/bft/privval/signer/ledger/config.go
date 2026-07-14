package ledger

import "errors"

// DefaultHDPath is the standard Cosmos/Tendermint validator BIP32
// derivation path: 44'/118'/0'/0/0.
var DefaultHDPath = []uint32{44, 118, 0, 0, 0}

// Config defines the configuration options for a Signer backed by a Ledger
// hardware wallet running the Tendermint Validator app.
type Config struct {
	// Enabled toggles the Ledger signer. It is a separate flag (rather than
	// gating on HDPath being empty, as the other external signers do on
	// their secret identifier) because a BIP32 path has no natural "unset"
	// value that isn't also a legitimate path.
	Enabled bool `json:"enabled" toml:"enabled" comment:"Enable signing with a Ledger hardware wallet running the Tendermint Validator app. If true, the local signer is disabled"`

	// HDPath is the BIP32 derivation path used to select the validator key
	// on the device. Defaults to DefaultHDPath if empty.
	HDPath []uint32 `json:"hd_path" toml:"hd_path" comment:"BIP32 derivation path for the validator key on the Ledger device. Defaults to [44, 118, 0, 0, 0] if empty"`
}

// DefaultConfig returns a default, disabled configuration for the Ledger signer.
func DefaultConfig() *Config {
	return &Config{
		Enabled: false,
		HDPath:  DefaultHDPath,
	}
}

// TestConfig returns a configuration for testing the Ledger signer.
func TestConfig() *Config {
	return DefaultConfig()
}

// IsEnabled reports whether the Ledger signer is configured for use.
func (cfg *Config) IsEnabled() bool {
	return cfg != nil && cfg.Enabled
}

// Config validation errors.
var errEmptyHDPath = errors.New("ledger signer: hd_path cannot be empty when enabled")

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *Config) ValidateBasic() error {
	if cfg != nil && cfg.Enabled && len(cfg.HDPath) == 0 {
		return errEmptyHDPath
	}

	return nil
}

// hdPath returns the configured HD path, defaulting to DefaultHDPath if unset.
func (cfg *Config) hdPath() []uint32 {
	if len(cfg.HDPath) == 0 {
		return DefaultHDPath
	}

	return cfg.HDPath
}
