package chainconfig

import (
	"cosmossdk.io/x/authz"
	banktypes "cosmossdk.io/x/bank/types"
	govv1 "cosmossdk.io/x/gov/types/v1"
	govv1beta1 "cosmossdk.io/x/gov/types/v1beta1"
	grouptypes "cosmossdk.io/x/group"
	proposaltypes "cosmossdk.io/x/params/types/proposal"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdktestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	wasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/host/types"
	feetypes "github.com/cosmos/ibc-go/v9/modules/apps/29-fee/types"
	transfertypes "github.com/cosmos/ibc-go/v9/modules/apps/transfer/types"
	v7migrations "github.com/cosmos/ibc-go/v9/modules/core/02-client/migrations/v7"
	clienttypes "github.com/cosmos/ibc-go/v9/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v9/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v9/modules/core/04-channel/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v9/modules/core/04-channel/v2/types"
	solomachine "github.com/cosmos/ibc-go/v9/modules/light-clients/06-solomachine"
	ibctmtypes "github.com/cosmos/ibc-go/v9/modules/light-clients/07-tendermint"
)

// Codec returns the global E2E protobuf codec.
func Codec() *codec.ProtoCodec {
	cdc, _ := codecAndEncodingConfig()
	return cdc
}

// SDKEncodingConfig returns the global E2E encoding config.
func SDKEncodingConfig() *sdktestutil.TestEncodingConfig {
	_, cfg := codecAndEncodingConfig()
	return &sdktestutil.TestEncodingConfig{
		InterfaceRegistry: cfg.InterfaceRegistry,
		Codec:             cfg.Codec,
		TxConfig:          cfg.TxConfig,
		Amino:             cfg.Amino,
	}
}

// codecAndEncodingConfig returns the codec and encoding config used in the E2E tests.
// Note: any new types added to the codec must be added here.
func codecAndEncodingConfig() (*codec.ProtoCodec, sdktestutil.TestEncodingConfig) {
	cfg := sdktestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{})

	// ibc types
	icacontrollertypes.RegisterInterfaces(cfg.InterfaceRegistry)
	icahosttypes.RegisterInterfaces(cfg.InterfaceRegistry)
	feetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	solomachine.RegisterInterfaces(cfg.InterfaceRegistry)
	v7migrations.RegisterInterfaces(cfg.InterfaceRegistry)
	transfertypes.RegisterInterfaces(cfg.InterfaceRegistry)
	clienttypes.RegisterInterfaces(cfg.InterfaceRegistry)
	channeltypes.RegisterInterfaces(cfg.InterfaceRegistry)
	connectiontypes.RegisterInterfaces(cfg.InterfaceRegistry)
	ibctmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	channeltypesv2.RegisterInterfaces(cfg.InterfaceRegistry)

	// all other types
	upgradetypes.RegisterInterfaces(cfg.InterfaceRegistry)
	banktypes.RegisterInterfaces(cfg.InterfaceRegistry)
	govv1beta1.RegisterInterfaces(cfg.InterfaceRegistry)
	govv1.RegisterInterfaces(cfg.InterfaceRegistry)
	authtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	cryptocodec.RegisterInterfaces(cfg.InterfaceRegistry)
	grouptypes.RegisterInterfaces(cfg.InterfaceRegistry)
	proposaltypes.RegisterInterfaces(cfg.InterfaceRegistry)
	authz.RegisterInterfaces(cfg.InterfaceRegistry)
	txtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	cdc := codec.NewProtoCodec(cfg.InterfaceRegistry)
	return cdc, cfg
}
