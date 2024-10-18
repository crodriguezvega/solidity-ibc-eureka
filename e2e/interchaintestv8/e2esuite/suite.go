package e2esuite

import (
	"context"
	"time"

	dockerclient "github.com/docker/docker/client"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	sdkmath "cosmossdk.io/math"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/chain/ethereum"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"

	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/chainconfig"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/testvalues"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/visualizerclient"
)

const DefaultVisualizerServerPort = 6969

// TestSuite is a suite of tests that require two chains and a relayer
type TestSuite struct {
	suite.Suite

	ChainA       *ethereum.EthereumChain
	ChainB       *cosmos.CosmosChain
	UserA        ibc.Wallet
	UserB        ibc.Wallet
	dockerClient *dockerclient.Client
	network      string
	logger       *zap.Logger
	ExecRep      *testreporter.RelayerExecReporter

	VisualizerClient    *visualizerclient.VisualizerClient
	CurrentNetworkState *visualizerclient.NetworkState
}

// SetupSuite sets up the chains, relayer, user accounts, clients, and connections
func (s *TestSuite) SetupSuite(ctx context.Context) {
	chainSpecs := chainconfig.DefaultChainSpecs

	if len(chainSpecs) != 2 {
		panic("TestSuite requires exactly 2 chain specs")
	}

	t := s.T()

	s.VisualizerClient = visualizerclient.NewVisualizerClient(DefaultVisualizerServerPort, t.Name())
	s.CurrentNetworkState = &visualizerclient.NetworkState{
		Name: "IBC Eureka",
	}
	s.setNotStartedNetworkState()
	s.VisualizerClient.SendMessage("Spinning up chains, relayer, accounts, etc...", "setup")

	s.logger = zaptest.NewLogger(t)
	s.dockerClient, s.network = interchaintest.DockerSetup(t)

	cf := interchaintest.NewBuiltinChainFactory(s.logger, chainSpecs)

	chains, err := cf.Chains(t.Name())
	s.Require().NoError(err)
	s.ChainA = chains[0].(*ethereum.EthereumChain)
	s.ChainB = chains[1].(*cosmos.CosmosChain)

	s.ExecRep = testreporter.NewNopReporter().RelayerExecReporter(t)

	ic := interchaintest.NewInterchain().
		AddChain(s.ChainA).
		AddChain(s.ChainB)

	s.Require().NoError(ic.Build(ctx, s.ExecRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           s.dockerClient,
		NetworkID:        s.network,
		SkipPathCreation: true,
	}))

	s.setStartedNetworkState()
	s.VisualizerClient.SendMessage("Chains started...", "setup")
	go func() {
		time.Sleep(10 * time.Second)
		s.RemoveColors()
		s.RemoveStatusTexts()
	}()

	// map all query request types to their gRPC method paths for cosmos chains
	s.Require().NoError(populateQueryReqToPath(ctx, s.ChainB))

	// Fund user accounts
	cosmosUserFunds := sdkmath.NewInt(testvalues.InitialBalance)
	cosmosUsers := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), cosmosUserFunds, s.ChainB)
	s.UserB = cosmosUsers[0]
	ethUsers := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), testvalues.StartingEthBalance, s.ChainA)
	s.UserA = ethUsers[0]

	s.VisualizerClient.SendMessage("Test users created and funded...", "setup")

	t.Cleanup(
		func() {
			s.RemoveArrows()
			s.RemoveColors()
			if t.Failed() {
				s.VisualizerClient.SendPopupMessage("Test failed...", "teardown")
				return
			} else {
				s.VisualizerClient.SendPopupMessage("Test passed!", "teardown")
			}
		},
	)
}
