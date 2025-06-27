package scalar

import (
	"fmt"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/stretchr/testify/require"
)

const (
	MNEMONIC = "ring start still sentence gas kid shy design ahead device adult movie appear rack provide scissors bag flat there sort soccer juice depend size"
)

func TestCreateKeyring(t *testing.T) {
	kr, sigAlgo, err := createKeyring(os.Stdin, "test", "secp256k1", "~/.scalar")
	if err != nil {
		t.Fatalf("failed to create keyring: %v", err)
	}
	info, err := kr.NewAccount("govenance", MNEMONIC, keyring.DefaultBIP39Passphrase, "m/2'/0'/0'/0/0", sigAlgo)
	if err != nil {
		t.Logf("failed to create account in keyring: %v", err)
		info, err = kr.Key("govenance")
	}
	if err != nil {
		t.Fatalf("failed to get account from keyring: %v", err)
	}
	fmt.Println("account: ", info.GetAddress().String())
	require.Equal(t, info.GetAddress().String(), "scalar1eaat9ffjf9ls9uafmpkrq0sjrwpqym8ptrndug")
	t.Logf("account: %v", info.GetAddress().String())
}
