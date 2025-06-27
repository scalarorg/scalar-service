package scalar

import (
	//Importing the types package is necessary to register the codec

	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	_ "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/std"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gogo/protobuf/proto"

	// "github.com/scalarorg/scalar-core/app/params"
	// multisigtypes "github.com/scalarorg/scalar-core/x/multisig/types"
	// scalarnettypes "github.com/scalarorg/scalar-core/x/scalarnet/types"
	// tsstypes "github.com/scalarorg/scalar-core/x/tss/types"
	"google.golang.org/grpc/encoding"
	encproto "google.golang.org/grpc/encoding/proto"
)

type customRegistry interface {
	RegisterCustomTypeURL(iface interface{}, typeURL string, impl proto.Message)
}

var protoCodec *codec.ProtoCodec
var interfaceRegistry codectypes.InterfaceRegistry

// This registers a codec that can encode custom Golang types defined by gogoproto extensions, which newer versions of the grpc module cannot.
// The fix has been extracted into its own module in order to minimize the number of dependencies
// that get imported before this init() function is called.
func init() {
	interfaceRegistry = codectypes.NewInterfaceRegistry()
	RegisterLegacyMsgInterfaces(interfaceRegistry)
	RegisterInterfaces(interfaceRegistry)
	RegisterImplementations(interfaceRegistry)
	gogoCodec := GogoEnabled{Codec: encoding.GetCodec(encproto.Name)}
	encoding.RegisterCodec(gogoCodec)
	protoCodec = codec.NewProtoCodec(interfaceRegistry)
}

// MakeEncodingConfig creates an EncodingConfig for testing
//
//	func MakeEncodingConfig() params.EncodingConfig {
//		encodingConfig := params.MakeEncodingConfig()
//		std.RegisterLegacyAminoCodec(encodingConfig.Amino)
//		RegisterLegacyMsgInterfaces(encodingConfig.InterfaceRegistry)
//		RegisterImplementations(encodingConfig.InterfaceRegistry)
//		RegisterInterfaces(encodingConfig.InterfaceRegistry)
//		return encodingConfig
//	}
func RegisterImplementations(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*proto.Message)(nil), &authtypes.BaseAccount{})
	registry.RegisterImplementations((*proto.Message)(nil), &secp256k1.PubKey{})
}
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	authtypes.RegisterInterfaces(registry)
	std.RegisterInterfaces(registry)
}
func RegisterLegacyMsgInterfaces(registry codectypes.InterfaceRegistry) {
	_, ok := registry.(customRegistry)
	if !ok {
		panic(fmt.Errorf("failed to convert registry type %T", registry))
	}
}

func GetInterfaceRegistry() codectypes.InterfaceRegistry {
	return interfaceRegistry
}

func GetProtoCodec() *codec.ProtoCodec {
	return protoCodec
}

type GogoEnabled struct {
	encoding.Codec
}

func (c GogoEnabled) Marshal(v interface{}) ([]byte, error) {
	if vv, ok := v.(proto.Marshaler); ok {
		return vv.Marshal()
	}
	return c.Codec.Marshal(v)
}

func (c GogoEnabled) Unmarshal(data []byte, v interface{}) error {
	if vv, ok := v.(proto.Unmarshaler); ok {
		return vv.Unmarshal(data)
	}
	return c.Codec.Unmarshal(data, v)
}
