package utils

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// func ConvertPubKeyToAddress(pubKeyHex string, network string) (string, error) {
// 	params := &chaincfg.TestNet3Params
// 	parts := strings.Split(network, "|")
// 	if len(parts) != 2 {
// 		return "", fmt.Errorf("invalid network format")
// 	}
// 	switch parts[1] {
// 	case "0":
// 		params = &chaincfg.MainNetParams
// 	case "4":
// 		params = &chaincfg.TestNet3Params
// 	default:
// 		params = &chaincfg.TestNet3Params
// 	}

// 	pubKeySerialized, err := hex.DecodeString(pubKeyHex)
// 	if err != nil {
// 		return "", err
// 	}
// 	addressPubKey, err := btcutil.NewAddressPubKey(pubKeySerialized, params)
// 	if err != nil {
// 		return "", err
// 	}
// 	return addressPubKey.AddressPubKeyHash().String(), nil
// }

func ScriptPubKeyToAddress(scriptHex string, network string) (btcutil.Address, error) {
	params := &chaincfg.TestNet3Params
	parts := strings.Split(network, "|")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid network format")
	}
	switch parts[1] {
	case "0":
		params = &chaincfg.MainNetParams
	case "4":
		params = &chaincfg.TestNet3Params
	default:
		params = &chaincfg.TestNet3Params
	}
	// Decode the hex string into bytes
	script, err := hex.DecodeString(scriptHex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %v", err)
	}
	// Extract the type of script
	_, addresses, _, err := txscript.ExtractPkScriptAddrs(script, params)
	if err != nil {
		return nil, err
	}

	// Usually we take the first address, but some scripts might have multiple
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no addresses found")
	}

	// TODO: Just support the simple case for now
	return addresses[0], nil
}
