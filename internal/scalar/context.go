package scalar

import (
	"bufio"
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/client"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/go-bip39"
	"github.com/rs/zerolog/log"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func defaultKeyringOption(options *keyring.Options) {
	options.SupportedAlgos = keyring.SigningAlgoList{hd.Secp256k1}
	options.SupportedAlgosLedger = keyring.SigningAlgoList{hd.Secp256k1}
}
func CreateClientContext(cmdClientCtx client.Context, params *ScalarParams) (client.Context, error) {
	codec := GetProtoCodec()
	clientCtx := cmdClientCtx.WithChainID(params.ChainID).
		WithTxConfig(tx.NewTxConfig(codec, []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT})).
		//WithCodec(codec).
		WithInput(cmdClientCtx.Input).
		WithFromName(params.Sender).
		WithOutputFormat("json")
	if params.SenderAddress != "" {
		accAddr, err := types.AccAddressFromBech32(params.SenderAddress)
		if err != nil {
			log.Debug().Err(err).Str("FromAddress", params.SenderAddress).Msg("Account address invalid")
		}
		clientCtx = clientCtx.WithFromAddress(accAddr)
	}
	if params.ScalarRPCUrl != "" {
		log.Info().Msgf("Create rpcClient using RPC URL %s", params.ScalarRPCUrl)
		clientCtx = clientCtx.WithNodeURI(params.ScalarRPCUrl)
		rpcClient, err := client.NewClientFromNode(params.ScalarRPCUrl)
		if err != nil {
			return clientCtx, fmt.Errorf("failed to create RPC client: %w", err)
		}
		clientCtx = clientCtx.WithClient(rpcClient)
	} else {
		log.Debug().Msg("Missing RPC URL, using local node")
	}
	kr, sigAlgo, err := createKeyring(cmdClientCtx.Input, params.KeyringBackend, params.KeyringAlgoName, params.WorkingDir)
	if err != nil {
		return clientCtx, fmt.Errorf("failed to create keyring: %w", err)
	}
	clientCtx = clientCtx.WithKeyring(kr)

	if params.Mnemonic != "" {
		// Create MemoryKeyring from mnemonic
		// kr := keyring.NewInMemory()
		// Set mnemonic to the in-memory keyring with name {FromName}
		bip44Path := fmt.Sprintf("%s/%d", BIP44_BASE_SCALAR_PATH, 0) //Fist index for call to scalarnode1
		info, err := kr.NewAccount(params.Sender, params.Mnemonic, keyring.DefaultBIP39Passphrase, bip44Path, sigAlgo)
		if err != nil {
			log.Debug().Err(err).Msg("Create account form mnemonic")
			info, err = kr.Key(params.Sender)
			if err != nil {
				log.Debug().Err(err).Msg("Get account from keyring")
				return clientCtx, fmt.Errorf("failed to get account from keyring: %w", err)
			}
		}
		log.Debug().Str("KeyName", params.Sender).
			Str("account", info.GetAddress().String()).
			Str("mnemonic", params.Mnemonic).
			Str("bip44Path", bip44Path).
			Msg("Created scalar client account in keyring")
		clientCtx = clientCtx.WithFromAddress(info.GetAddress())
	}
	return clientCtx, nil
}
func createKeyring(input io.Reader, krBackend string, algoName string, nodeDir string) (keyring.Keyring, keyring.SignatureAlgo, error) {
	serviceName := types.KeyringServiceName()
	log.Debug().Str("serviceName", serviceName).Str("krBackend", krBackend).Str("algoName", algoName).Str("nodeDir", nodeDir).Msg("Creating keyring")
	inBuf := bufio.NewReader(input)
	kb, err := keyring.New(serviceName, krBackend, nodeDir, inBuf, defaultKeyringOption)
	if err != nil {
		return nil, nil, err
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	sigAlgo, err := keyring.NewSigningAlgoFromString(algoName, keyringAlgos)
	if err != nil {
		return nil, nil, err
	}
	return kb, sigAlgo, nil
}
func CreateEd25519AccountFromMnemonic(mnemonic string) (ed25519.PrivKey, types.AccAddress, error) {
	privKey := ed25519.GenPrivKeyFromSecret([]byte(mnemonic))
	//privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	//addr := types.AccAddress(privKey.PubKey().Address())
	pubkey, err := cryptocodec.FromTmPubKeyInterface(privKey.PubKey())
	if err != nil {
		return nil, nil, err
	}
	addr := types.AccAddress(pubkey.Address())
	return privKey, addr, nil
}

func CreateBip39AccountFromMnemonic(mnemonic string) (*secp256k1.PrivKey, types.AccAddress, error) {
	// Derive the seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// Create master key and derive the private key
	// Using "m/44'/118'/0'/0/0" for Cosmos
	master, ch := hd.ComputeMastersFromSeed(seed)
	privKeyBytes, err := hd.DerivePrivateKeyForPath(master, ch, BIP44PATH)
	if err != nil {
		return nil, nil, err
	}

	// Create private key and get address
	privKey := &secp256k1.PrivKey{Key: privKeyBytes}
	addr := types.AccAddress(privKey.PubKey().Address())
	log.Debug().Msgf("Created scalar broadcaster account address: %s from mnemonic", addr.String())
	return privKey, addr, nil
}
