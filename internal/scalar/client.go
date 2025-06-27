package scalar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/rs/zerolog/log"

	// chainstypes "github.com/scalarorg/scalar-core/x/chains/types"
	// covtypes "github.com/scalarorg/scalar-core/x/covenant/types"
	// multisigtypes "github.com/scalarorg/scalar-core/x/multisig/types"
	// scalarnet "github.com/scalarorg/scalar-core/x/scalarnet/exported"
	tmclient "github.com/tendermint/tendermint/rpc/client"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

const (
	DEFAULT_GAS_ADJUSTMENT        = 1.3
	GAS_PRICE                     = 0.0125
	ACCOUNT_PREFIX_SCALAR         = "scalar"
	BaseAsset                     = "ascal"
	BIP44PATH              string = "m/44'/118'/0'/0/0"
	BIP44_BASE_SCALAR_PATH string = "m/44'/118'/0'/0"
	BIP44_BASE_EVM_PATH    string = "m/44'/60'/0'/0"
)

type ScalarParams struct {
	ChainID           string
	ConfigPath        string
	ConnectionString  string
	EvmPrivateKey     string
	EvmMnemonic       string
	KeyringBackend    string
	KeyringAlgoName   string
	Mnemonic          string
	Name              string
	ScalarRPCUrl      string
	Sender            string
	SenderAddress     string
	SQLitePath        string
	Timeout           time.Duration
	WorkingDir        string
	RecipientMnemonic string
	//StartAccountIndex uint32 //Default 0
	RecipientNumber uint32 //index from 1
	Denom           string
	Symbol          string
	Amount          int64

	EvmRpcUrl       string
	ContractAddress string
	//TokenAddress    string
}

type ScalarClient struct {
	Params          *ScalarParams
	TxFactory       tx.Factory
	ClientCtx       client.Context
	RpcClient       tmclient.Client
	AuthQueryClient auth.QueryClient
	BankQueryClient bank.QueryClient
	grpcConn        ScalarGrpcConn
}

func NewScalarClient(cmdClientCtx client.Context, params *ScalarParams) (*ScalarClient, error) {
	log.Info().Msgf("Create ScalarClient with params: %+v", params)
	clientCtx, err := CreateClientContext(cmdClientCtx, params)
	if err != nil {
		return nil, err
	}
	rpcClient, err := client.NewClientFromNode(params.ScalarRPCUrl)
	if err != nil {
		return nil, err
	}
	txFactory := tx.Factory{}
	grpcConn := ScalarGrpcConn{
		Codec:         encoding.GetCodec(proto.Name),
		ClientContext: clientCtx,
	}
	return &ScalarClient{
		Params:          params,
		ClientCtx:       clientCtx,
		RpcClient:       rpcClient,
		AuthQueryClient: auth.NewQueryClient(clientCtx),
		BankQueryClient: bank.NewQueryClient(clientCtx),
		TxFactory:       txFactory.WithAccountNumber(0).WithSequence(0).WithMemo(""),
		grpcConn:        grpcConn,
	}, nil
}

func (c *ScalarClient) GetTxServiceClient() txtypes.ServiceClient {
	return txtypes.NewServiceClient(c.grpcConn)
}

// Inject account number and sequence number into txFactory for signing
func (c *ScalarClient) createTxFactory() tx.Factory {
	txFactory := c.TxFactory.
		WithTxConfig(c.ClientCtx.TxConfig).
		WithAccountRetriever(c.ClientCtx.AccountRetriever).
		WithKeybase(c.ClientCtx.Keyring).
		WithChainID(c.ClientCtx.ChainID).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithGas(0). // Adjust in estimateGas()
		WithGasAdjustment(DEFAULT_GAS_ADJUSTMENT)
	// txFactory = txFactory.WithGasPrices(config.GasPrice)
	// txFactory = txFactory.WithFees(sdk.NewCoin("uaxl", sdk.NewInt(20000)).String())
	if txFactory.AccountRetriever() == nil {
		log.Error().Msg("AccountRetriever is nil")
	}
	resp, err := c.AuthQueryClient.Account(context.Background(), &auth.QueryAccountRequest{
		Address: c.ClientCtx.FromAddress.String(),
	})
	if err != nil {
		log.Error().Err(err).Msg("[createTxFactory] failed to get account")
	} else if resp.Account == nil {
		log.Error().Msgf("account not found")
	} else {
		var account auth.BaseAccount
		err = c.UnmarshalAccount(resp, &account)
		if err != nil {
			log.Error().Err(err).Msg("failed to unmarshal account")
		}
		log.Debug().Uint64("accoutNumber", account.AccountNumber).Uint64("sequence", account.Sequence).Msg("Got account from network")
		txFactory = txFactory.WithAccountNumber(account.AccountNumber)
		//If sequence number is greater than current sequence number, update the sequence number
		//This is to avoid the situation where the transaction is not included in the next block
		//Then account sequence number is not updated on the server side
		if account.Sequence >= txFactory.Sequence() {
			txFactory = txFactory.WithSequence(account.Sequence)
		}
	}
	return txFactory
}

// Todo: Add code for more correct unmarshal
func (c *ScalarClient) UnmarshalAccount(resp *auth.QueryAccountResponse, account *auth.BaseAccount) error {
	// err = account.Unmarshal(resp.Account.Value)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to unmarshal account: %w", err)
	// }
	buf := &bytes.Buffer{}
	clientCtx := c.ClientCtx.WithOutput(buf)
	err := clientCtx.PrintProto(resp.Account)
	if err != nil {
		return fmt.Errorf("failed to print proto: %w", err)
	}
	var accountMap map[string]any
	err = json.Unmarshal(buf.Bytes(), &accountMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal account: %w", err)
	}
	log.Debug().Msgf("accountMap: %v", accountMap)
	account.Address = accountMap["address"].(string)
	account.AccountNumber, err = strconv.ParseUint(accountMap["account_number"].(string), 10, 64)
	if err != nil {
		log.Error().Msgf("failed to parse account number: %+v", err)
	}
	account.Sequence, err = strconv.ParseUint(accountMap["sequence"].(string), 10, 64)
	if err != nil {
		log.Error().Msgf("failed to parse sequence: %+v", err)
	}
	//pubKey := secp256k1.PubKey{}
	//pubKey.Key = accountMap["public_key"].(map[string]any)["key"].(string)
	//account.PubKey = &pubKey
	return nil
}

func (c *ScalarClient) GetBalance(address string) (*sdk.Coin, error) {
	// Convert address string to sdk.AccAddress
	accAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Query balance using bank module
	resp, err := c.BankQueryClient.Balance(context.Background(), &bank.QueryBalanceRequest{
		Address: accAddr.String(),
		Denom:   c.Params.Denom,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query balance: %w", err)
	}

	return resp.Balance, nil
}

func (c *ScalarClient) generateTx(msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	txFactory := c.createTxFactory()
	var buffer bytes.Buffer
	ctx := c.ClientCtx.WithOutput(&buffer)
	err := tx.GenerateTx(ctx, txFactory, msgs...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate tx")
		return nil, fmt.Errorf("failed to generate tx: %w", err)
	}
	var txResponse sdk.TxResponse
	ctx.Codec.UnmarshalJSON(buffer.Bytes(), &txResponse)
	return &txResponse, nil
}

func (c *ScalarClient) broadcastTxs(msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	txFactory := c.createTxFactory()
	//Estimate fees
	simRes, adjusted, err := tx.CalculateGas(c.ClientCtx, txFactory, msgs...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to calculate gas")
		return nil, fmt.Errorf("failed to calculate gas: %w", err)
	}
	fees := int64(txFactory.GasAdjustment() * float64(simRes.GasInfo.GasUsed) * GAS_PRICE)
	log.Debug().Msgf("[ScalarClient] [broadcastTx] Adjusted gas: %d, Fees: %d", adjusted, fees)
	txFactory = txFactory.WithGas(adjusted).
		WithFees(sdk.NewCoin(BaseAsset, sdk.NewInt(fees)).String())
	var buffer bytes.Buffer
	ctx := c.ClientCtx.WithOutput(&buffer)
	//Cosmos SDK Sign transaction with key name from context of in the keyring
	err = tx.GenerateOrBroadcastTxWithFactory(ctx, txFactory, msgs...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate key")
		return nil, err
	}
	var txResponse sdk.TxResponse
	ctx.Codec.UnmarshalJSON(buffer.Bytes(), &txResponse)
	return &txResponse, nil
}

func (c *ScalarClient) estimateFee(privateKey cryptotypes.PrivKey, msgs ...sdk.Msg) (*sdk.GasInfo, error) {
	txFactory := c.createTxFactory()
	ctx := c.ClientCtx
	txBuilder := ctx.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msgs...)
	sigMode := c.ClientCtx.TxConfig.SignModeHandler().DefaultMode()
	sigV2 := signing.SignatureV2{
		PubKey: privateKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  sigMode,
			Signature: nil,
		},
		Sequence: txFactory.Sequence(),
	}
	err := txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signatures: %w", err)
	}
	// Second round: all signer infos are set, so each signer can sign.
	signerData := xauthsigning.SignerData{
		ChainID:       c.ClientCtx.ChainID,
		AccountNumber: txFactory.AccountNumber(),
		Sequence:      txFactory.Sequence(),
	}
	sigV2, err = tx.SignWithPrivKey(
		sigMode, signerData,
		txBuilder, privateKey, c.ClientCtx.TxConfig, txFactory.Sequence())
	if err != nil {
		return nil, err
	}
	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signatures: %w", err)
	}

	//Estimate gas
	txBytes, err := ctx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}
	txSvcClient := txtypes.NewServiceClient(c.grpcConn)
	simRes, err := txSvcClient.Simulate(context.Background(), &txtypes.SimulateRequest{
		TxBytes: txBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to simulate tx: %w", err)
	}
	// adjusted := uint64(txFactory.GasAdjustment() * float64(simRes.GasInfo.GasUsed))
	// fees := int64(float64(adjusted) * GasPrice)
	return simRes.GasInfo, nil
}

// Broadcast Msgs with private key.
// Account and private key created by Keyring are different from the account and private key created by ScalarClient
func (c *ScalarClient) broadcastMsgs(privateKey cryptotypes.PrivKey, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	// Estimate fees
	var gasInfo *sdk.GasInfo
	var err error
	for i := 0; i < 3; i++ {
		gasInfo, err = c.estimateFee(privateKey, msgs...)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to estimate fee. Retry %d", i)
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}
	if err != nil {
		return nil, fmt.Errorf("failed to estimate fee: %w", err)
	}
	txFactory := c.createTxFactory()
	adjusted := uint64(txFactory.GasAdjustment() * float64(gasInfo.GasUsed))
	fees := int64(float64(adjusted) * GAS_PRICE)
	log.Debug().Msgf("[ScalarClient] [broadcastTx] Adjusted gas: %d, Fees: %d", adjusted, fees)

	//var buffer bytes.Buffer
	//ctx := c.ClientCtx.WithOutput(&buffer)
	ctx := c.ClientCtx
	txBuilder := ctx.TxConfig.NewTxBuilder()
	txBuilder.SetMsgs(msgs...)
	txBuilder.SetGasLimit(adjusted)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(BaseAsset, sdk.NewInt(fees))))
	// txBuilder.SetMemo(ctx.Memo)
	// txBuilder.SetTimeoutHeight(ctx.Height)

	//Sign tx
	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigMode := c.ClientCtx.TxConfig.SignModeHandler().DefaultMode()
	sigV2 := signing.SignatureV2{
		PubKey: privateKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  sigMode,
			Signature: nil,
		},
		Sequence: txFactory.Sequence(),
	}
	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signatures: %w", err)
	}
	// Second round: all signer infos are set, so each signer can sign.
	signerData := xauthsigning.SignerData{
		ChainID:       c.ClientCtx.ChainID,
		AccountNumber: txFactory.AccountNumber(),
		Sequence:      txFactory.Sequence(),
	}
	log.Debug().Msgf("[ScalarClient] [broadcastMsgs] Signer data: %+v", signerData)
	sigV2, err = tx.SignWithPrivKey(
		sigMode, signerData,
		txBuilder, privateKey, c.ClientCtx.TxConfig, txFactory.Sequence())
	if err != nil {
		return nil, err
	}
	// verifySignature(privateKey, sigV2)
	err = txBuilder.SetSignatures(sigV2)
	if err != nil {
		return nil, fmt.Errorf("failed to set signatures: %w", err)
	}
	// broadcast to a Tendermint node
	txBytes, err := ctx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}
	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *ScalarClient) broardcastTxWithRetry(msg sdk.Msg, interval time.Duration, maxRetry *int) (*sdk.TxResponse, error) {
	counter := 0
	for {
		counter += 1
		txRes, err := c.broadcastTxs(msg)
		if err == nil {
			return txRes, nil
		}
		if maxRetry != nil && counter >= *maxRetry {
			return nil, err
		}
		if counter%10 == 0 {
			log.Debug().Err(err).Msgf("Broadcast tx failed. Retry counter %d", counter)
		}
		time.Sleep(interval)
	}
}
