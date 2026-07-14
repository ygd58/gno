package ledger

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDevice is a fake implementing ledgerAPI, used so tests never require
// a physical Ledger device.
type mockDevice struct {
	pubKey []byte
	sig    []byte

	getPubKeyErr error
	signErr      error
	closed       bool

	// lastSignPath and lastSignMessage record the arguments of the most
	// recent SignED25519 call, so tests can assert the correct HD path and
	// bytes were sent to the device.
	lastSignPath    []uint32
	lastSignMessage []byte
}

func (m *mockDevice) GetPublicKeyED25519(_ []uint32) ([]byte, error) {
	if m.getPubKeyErr != nil {
		return nil, m.getPubKeyErr
	}

	return m.pubKey, nil
}

func (m *mockDevice) SignED25519(bip32Path []uint32, message []byte) ([]byte, error) {
	m.lastSignPath = bip32Path
	m.lastSignMessage = message

	if m.signErr != nil {
		return nil, m.signErr
	}

	return m.sig, nil
}

func (m *mockDevice) Close() error {
	m.closed = true

	return nil
}

func TestNewSigner_ValidDevice(t *testing.T) {
	t.Parallel()

	device := &mockDevice{
		pubKey: make([]byte, 32),
		sig:    make([]byte, 64),
	}
	device.pubKey[0] = 0xAB
	device.sig[0] = 0xCD

	hdPath := []uint32{44, 118, 0, 0, 0}

	signer, err := newSigner(device, hdPath)
	require.NoError(t, err)
	require.NotNil(t, signer)

	assert.Equal(t, device.pubKey, signer.pubKey[:])
	assert.Contains(t, signer.String(), "LedgerSigner")

	sig, err := signer.Sign([]byte("sign-bytes"))
	require.NoError(t, err)
	assert.Equal(t, device.sig, sig)
	assert.Equal(t, hdPath, device.lastSignPath)
	assert.Equal(t, []byte("sign-bytes"), device.lastSignMessage)

	require.NoError(t, signer.Close())
	assert.True(t, device.closed)
}

func TestNewSigner_GetPubKeyError(t *testing.T) {
	t.Parallel()

	device := &mockDevice{getPubKeyErr: errors.New("device not connected")}

	signer, err := newSigner(device, DefaultHDPath)
	require.Error(t, err)
	assert.Nil(t, signer)
}

func TestNewSigner_InvalidPubKeyLength(t *testing.T) {
	t.Parallel()

	device := &mockDevice{pubKey: make([]byte, 16)} // wrong length

	signer, err := newSigner(device, DefaultHDPath)
	require.ErrorIs(t, err, errInvalidPubKeyLen)
	assert.Nil(t, signer)
}

func TestNewSigner_SignError(t *testing.T) {
	t.Parallel()

	device := &mockDevice{
		pubKey:  make([]byte, 32),
		signErr: errors.New("user rejected on device"),
	}

	signer, err := newSigner(device, DefaultHDPath)
	require.NoError(t, err)

	sig, err := signer.Sign([]byte("sign-bytes"))
	require.Error(t, err)
	assert.Nil(t, sig)
}

func TestConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	assert.False(t, (&Config{}).IsEnabled())
	assert.False(t, (*Config)(nil).IsEnabled())
	assert.True(t, (&Config{Enabled: true}).IsEnabled())
}

func TestConfig_ValidateBasic(t *testing.T) {
	t.Parallel()

	assert.NoError(t, (&Config{}).ValidateBasic())
	assert.NoError(t, (&Config{Enabled: true, HDPath: DefaultHDPath}).ValidateBasic())
	assert.ErrorIs(t, (&Config{Enabled: true}).ValidateBasic(), errEmptyHDPath)
}

func TestConfig_HDPath(t *testing.T) {
	t.Parallel()

	assert.Equal(t, DefaultHDPath, (&Config{}).hdPath())

	custom := []uint32{44, 118, 0, 0, 1}
	assert.Equal(t, custom, (&Config{HDPath: custom}).hdPath())
}

func TestNewSignerFromConfig_Disabled(t *testing.T) {
	t.Parallel()

	_, err := NewSignerFromConfig(DefaultConfig())
	require.ErrorIs(t, err, errDisabled)
}
