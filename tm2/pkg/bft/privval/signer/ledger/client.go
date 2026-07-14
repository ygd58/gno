package ledger

import (
	"fmt"

	ledgercosmos "github.com/cosmos/ledger-cosmos-go"
)

// ledgerAPI is the subset of the Ledger Tendermint Validator app client used
// by the signer. It is defined as an interface so tests can substitute a
// mock implementation without requiring a physical Ledger device.
type ledgerAPI interface {
	GetPublicKeyED25519(bip32Path []uint32) ([]byte, error)
	SignED25519(bip32Path []uint32, message []byte) ([]byte, error)
	Close() error
}

// ledgerAPI is implemented by *ledgercosmos.LedgerTendermintValidator.
var _ ledgerAPI = (*ledgercosmos.LedgerTendermintValidator)(nil)

// newDevice connects to a Ledger device running the Tendermint Validator app
// over USB/HID.
func newDevice() (ledgerAPI, error) {
	dev, err := ledgercosmos.FindLedgerCosmosValidatorApp()
	if err != nil {
		return nil, fmt.Errorf("unable to find Ledger Tendermint Validator app: %w", err)
	}

	return dev, nil
}
