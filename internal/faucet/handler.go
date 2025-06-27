package faucet

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/scalarorg/scalar-service/config"
	"github.com/scalarorg/scalar-service/internal/scalar"
)

const (
	FAUCET_AMOUNT = 1_000_000_000
)

func handleFaucet(c echo.Context) error {
	var rawBody map[string]interface{}
	err := c.Bind(&rawBody)
	if err != nil {
		return err
	}
	value := rawBody["address"]
	userAddress, ok := value.(string)
	if !ok {
		return c.JSON(400, map[string]string{"error": "Invalid address"})
	}
	clientCtx := c.Get("clientCtx").(*client.Context)
	scalarClient, err := scalar.NewScalarClient(clientCtx, &scalar.ScalarParams{
		ChainID:           config.Env.CHAIN_ID,
		ConfigPath:        config.Env.CONFIG_PATH,
		ConnectionString:  config.Env.CONNECTION_STRING,
		EvmPrivateKey:     config.Env.EVM_PRIVATE_KEY,
		EvmMnemonic:       config.Env.EVM_MNEMONIC,
		KeyringBackend:    config.Env.KEYRING_BACKEND,
		KeyringAlgoName:   config.Env.KEYRING_ALGO_NAME,
		Mnemonic:          config.Env.MNEMONIC,
		Name:              config.Env.NAME,
		ScalarRPCUrl:      config.Env.SCALAR_RPC_URL,
		Sender:            config.Env.SENDER,
		SenderAddress:     config.Env.SENDER_ADDRESS,
		SQLitePath:        config.Env.SQLITE_PATH,
		Timeout:           config.Env.TIMEOUT,
		WorkingDir:        config.Env.WORKING_DIR,
		RecipientMnemonic: config.Env.RECIPIENT_MNEMONIC,
	})
	currentBalance, err := scalarClient.GetBalance(userAddress)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Failed to get balance"})
	}
	if currentBalance.Amount.Int64() >= FAUCET_AMOUNT {
		return c.JSON(400, map[string]string{"error": fmt.Sprintf("Address already has %d %s", currentBalance.Amount.Int64(), DENOM)})
	}
	privKeyBytes, err := hex.DecodeString(config.Env.FAUCET_PRIVATE_KEY)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Failed to decode private key"})
	}
	privKey := secp256k1.PrivKey{Key: privKeyBytes}
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Failed to get private key"})
	}
	faucetAddress, err := sdk.Bech32ifyAddressBytes(scalar.ACCOUNT_PREFIX_SCALAR, privKey.PubKey().Address())
	if err != nil {
		log.Fatal().Msgf("Error generating address: %++v", err)
	}
	msg := banktypes.MsgSend{
		FromAddress: faucetAddress,
		ToAddress:   userAddress,
		Amount:      sdk.NewCoins(sdk.NewCoin(scalar.BaseAsset, sdk.NewInt(FAUCET_AMOUNT))),
	}
	txResponse, err := scalarClient.broadcastMsgs(&privKey, &msg)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Failed to broadcast transaction"})
	}

	return c.JSON(200, map[string]string{"message": "Transaction broadcasted successfully", "tx_hash": txResponse.TxHash})
}
