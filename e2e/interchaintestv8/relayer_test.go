package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
	ibchostv2 "github.com/cosmos/ibc-go/v10/modules/core/24-host/v2"

	"github.com/strangelove-ventures/interchaintest/v8/testutil"

	"github.com/cosmos/solidity-ibc-eureka/packages/go-abigen/ibcerc20"
	"github.com/cosmos/solidity-ibc-eureka/packages/go-abigen/ics20transfer"

	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/chainconfig"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/e2esuite"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/operator"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/testvalues"
	ethereumtypes "github.com/srdtrk/solidity-ibc-eureka/e2e/v8/types/ethereum"
	relayertypes "github.com/srdtrk/solidity-ibc-eureka/e2e/v8/types/relayer"
)

// RelayerTestSuite is a suite of tests that wraps IbcEurekaTestSuite
// and can provide additional functionality
type RelayerTestSuite struct {
	IbcEurekaTestSuite
}

// TestWithIbcEurekaTestSuite is the boilerplate code that allows the test suite to be run
func TestWithRelayerTestSuite(t *testing.T) {
	suite.Run(t, new(RelayerTestSuite))
}

func (s *RelayerTestSuite) Test_10_RecvPacketToEth_Groth16() {
	ctx := context.Background()
	s.FilteredRecvPacketToEthTest(ctx, operator.ProofTypeGroth16, 10, nil)
}

func (s *RelayerTestSuite) Test_5_RecvPacketToEth_Plonk() {
	ctx := context.Background()
	s.FilteredRecvPacketToEthTest(ctx, operator.ProofTypePlonk, 5, []uint64{1, 2, 3, 4, 5})
}

func (s *RelayerTestSuite) Test_10_FilteredRecvPacketToEth_Groth16() {
	ctx := context.Background()
	s.FilteredRecvPacketToEthTest(ctx, operator.ProofTypeGroth16, 10, []uint64{2, 6})
}

func (s *RelayerTestSuite) FilteredRecvPacketToEthTest(
	ctx context.Context, proofType operator.SupportedProofType, numOfTransfers int, recvFilter []uint64,
) {
	s.Require().GreaterOrEqual(numOfTransfers, len(recvFilter))
	s.Require().Greater(numOfTransfers, 0)

	s.SetupSuite(ctx, proofType)

	eth, simd := s.EthChain, s.CosmosChains[0]

	ics26Address := ethcommon.HexToAddress(s.contractAddresses.Ics26Router)
	transferAmount := big.NewInt(testvalues.TransferAmount)
	totalTransferAmount := big.NewInt(testvalues.TransferAmount * int64(numOfTransfers))
	if totalTransferAmount.Int64() > testvalues.InitialBalance {
		s.FailNow("Total transfer amount exceeds the initial balance")
	}
	var totalRecvAmount *big.Int
	if len(recvFilter) == 0 {
		totalRecvAmount = totalTransferAmount
	} else {
		totalRecvAmount = big.NewInt(testvalues.TransferAmount * int64(len(recvFilter)))
	}
	ethereumUserAddress := crypto.PubkeyToAddress(s.key.PublicKey)
	cosmosUserWallet := s.CosmosUsers[0]
	cosmosUserAddress := cosmosUserWallet.FormattedAddress()

	var (
		transferCoin sdk.Coin
		sendTxHashes [][]byte
	)
	s.Require().True(s.Run("Send transfers on Cosmos chain", func() {
		for range numOfTransfers {
			timeout := uint64(time.Now().Add(30 * time.Minute).Unix())
			transferCoin = sdk.NewCoin(simd.Config().Denom, sdkmath.NewIntFromBigInt(transferAmount))

			transferPayload := transfertypes.FungibleTokenPacketData{
				Denom:    transferCoin.Denom,
				Amount:   transferCoin.Amount.String(),
				Sender:   cosmosUserAddress,
				Receiver: strings.ToLower(ethereumUserAddress.Hex()),
				Memo:     "",
			}
			encodedPayload, err := transfertypes.EncodeABIFungibleTokenPacketData(&transferPayload)
			s.Require().NoError(err)

			payload := channeltypesv2.Payload{
				SourcePort:      transfertypes.PortID,
				DestinationPort: transfertypes.PortID,
				Version:         transfertypes.V1,
				Encoding:        transfertypes.EncodingABI,
				Value:           encodedPayload,
			}
			msgSendPacket := channeltypesv2.MsgSendPacket{
				SourceClient:     testvalues.FirstWasmClientID,
				TimeoutTimestamp: timeout,
				Payloads: []channeltypesv2.Payload{
					payload,
				},
				Signer: cosmosUserWallet.FormattedAddress(),
			}

			resp, err := s.BroadcastMessages(ctx, simd, cosmosUserWallet, 200_000, &msgSendPacket)
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.TxHash)

			txHash, err := hex.DecodeString(resp.TxHash)
			s.Require().NoError(err)

			sendTxHashes = append(sendTxHashes, txHash)
		}

		s.Require().True(s.Run("Verify balances on Cosmos chain", func() {
			// Check the balance of UserB
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simd, &banktypes.QueryBalanceRequest{
				Address: cosmosUserAddress,
				Denom:   transferCoin.Denom,
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(testvalues.InitialBalance-totalTransferAmount.Int64(), resp.Balance.Amount.Int64())
		}))
	}))

	s.Require().True(s.Run("Receive packets on Ethereum", func() {
		var relayTx []byte
		s.Require().True(s.Run("Retrieve relay tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:           simd.Config().ChainID,
				DstChain:           eth.ChainID.String(),
				SourceTxIds:        sendTxHashes,
				SrcClientId:        testvalues.FirstWasmClientID,
				DstClientId:        testvalues.CustomClientID,
				SrcPacketSequences: recvFilter,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Equal(resp.Address, ics26Address.String())

			relayTx = resp.Tx
		}))

		s.Require().True(s.Run("Submit relay tx", func() {
			receipt, err := eth.BroadcastTx(ctx, s.EthRelayerSubmitter, 5_000_000, ics26Address, relayTx)
			s.Require().NoError(err)
			s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status, fmt.Sprintf("Tx failed: %+v", receipt))
		}))

		s.Require().True(s.Run("Verify balances on Ethereum", func() {
			denomOnEthereum := transfertypes.NewDenom(transferCoin.Denom, transfertypes.NewHop(transfertypes.PortID, testvalues.CustomClientID))

			ibcERC20Addr, err := s.ics20Contract.IbcERC20Contract(nil, denomOnEthereum.Path())
			s.Require().NoError(err)

			ibcERC20, err := ibcerc20.NewContract(ethcommon.HexToAddress(ibcERC20Addr.Hex()), s.EthChain.RPCClient)
			s.Require().NoError(err)

			// User balance on Ethereum
			userBalance, err := ibcERC20.BalanceOf(nil, ethereumUserAddress)
			s.Require().NoError(err)
			s.Require().Equal(totalRecvAmount, userBalance)
		}))
	}))
}

// TestConcurrentRecvPacketToEth_Groth16 tests the concurrent relaying of 2 packets from Cosmos to Ethereum
// NOTE: This test is not included in the CI pipeline as it is flaky
func (s *RelayerTestSuite) Test_2_ConcurrentRecvPacketToEth_Groth16() {
	// I've noticed that the prover network drops the requests when sending too many
	ctx := context.Background()
	s.ConcurrentRecvPacketToEthTest(ctx, operator.ProofTypeGroth16, 2)
}

func (s *RelayerTestSuite) ConcurrentRecvPacketToEthTest(
	ctx context.Context, proofType operator.SupportedProofType, numConcurrentTransfers int,
) {
	s.Require().Greater(numConcurrentTransfers, 0)

	s.SetupSuite(ctx, proofType)

	_, simd := s.EthChain, s.CosmosChains[0]

	ics26Address := ethcommon.HexToAddress(s.contractAddresses.Ics26Router)
	transferAmount := big.NewInt(testvalues.TransferAmount)
	transferCoin := sdk.NewCoin(simd.Config().Denom, sdkmath.NewIntFromBigInt(transferAmount))
	totalTransferAmount := big.NewInt(testvalues.TransferAmount * int64(numConcurrentTransfers))
	if totalTransferAmount.Int64() > testvalues.InitialBalance {
		s.FailNow("Total transfer amount exceeds the initial balance")
	}
	ethereumUserAddress := crypto.PubkeyToAddress(s.key.PublicKey)
	cosmosUserWallet := s.CosmosUsers[0]
	cosmosUserAddress := cosmosUserWallet.FormattedAddress()

	var sendTxHashes [][]byte
	s.Require().True(s.Run("Send transfers on Cosmos chain", func() {
		for range numConcurrentTransfers {
			timeout := uint64(time.Now().Add(30 * time.Minute).Unix())

			transferPayload := transfertypes.FungibleTokenPacketData{
				Denom:    transferCoin.Denom,
				Amount:   transferCoin.Amount.String(),
				Sender:   cosmosUserAddress,
				Receiver: strings.ToLower(ethereumUserAddress.Hex()),
				Memo:     "",
			}
			encodedPayload, err := transfertypes.EncodeABIFungibleTokenPacketData(&transferPayload)
			s.Require().NoError(err)

			payload := channeltypesv2.Payload{
				SourcePort:      transfertypes.PortID,
				DestinationPort: transfertypes.PortID,
				Version:         transfertypes.V1,
				Encoding:        transfertypes.EncodingABI,
				Value:           encodedPayload,
			}
			msgSendPacket := channeltypesv2.MsgSendPacket{
				SourceClient:     testvalues.FirstWasmClientID,
				TimeoutTimestamp: timeout,
				Payloads: []channeltypesv2.Payload{
					payload,
				},
				Signer: cosmosUserWallet.FormattedAddress(),
			}

			resp, err := s.BroadcastMessages(ctx, simd, cosmosUserWallet, 200_000, &msgSendPacket)
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.TxHash)

			txHash, err := hex.DecodeString(resp.TxHash)
			s.Require().NoError(err)

			sendTxHashes = append(sendTxHashes, txHash)
		}

		s.Require().True(s.Run("Verify balances on Cosmos chain", func() {
			// Check the balance of UserB
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simd, &banktypes.QueryBalanceRequest{
				Address: cosmosUserAddress,
				Denom:   transferCoin.Denom,
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(testvalues.InitialBalance-totalTransferAmount.Int64(), resp.Balance.Amount.Int64())
		}))
	}))

	s.Require().True(s.Run("Install circuit artifacts on machine", func() {
		// When running multiple instances of the relayer, the circuit artifacts need to be installed on the machine
		// to avoid the overhead of installing the artifacts for each relayer instance (which also panics).
		// This is why we make a single request which installs the artifacts on the machine, and discard the response.

		resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
			SrcChain:    simd.Config().ChainID,
			DstChain:    s.EthChain.ChainID.String(),
			SourceTxIds: sendTxHashes,
			SrcClientId: testvalues.FirstWasmClientID,
			DstClientId: testvalues.CustomClientID,
		})
		s.Require().NoError(err)
		s.Require().NotEmpty(resp.Tx)
		s.Require().Equal(resp.Address, ics26Address.String())
	}))

	var wg sync.WaitGroup
	wg.Add(numConcurrentTransfers)
	s.Require().True(s.Run("Make concurrent requests", func() {
		// loop over the txHashes and send them concurrently
		for _, txHash := range sendTxHashes {
			// we send the request while the previous request is still being processed
			time.Sleep(3 * time.Second)
			go func() {
				defer wg.Done() // decrement the counter when the request completes
				resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
					SrcChain:    simd.Config().ChainID,
					DstChain:    s.EthChain.ChainID.String(),
					SourceTxIds: [][]byte{txHash},
					SrcClientId: testvalues.FirstWasmClientID,
					DstClientId: testvalues.CustomClientID,
				})
				s.Require().NoError(err)
				s.Require().NotEmpty(resp.Tx)
				s.Require().Equal(resp.Address, ics26Address.String())
			}()
		}
	}))

	s.Require().True(s.Run("Wait for all requests to complete", func() {
		// wait for all requests to complete
		// If the request never completes, we rely on the test timeout to kill the test
		wg.Wait()
	}))
}

func (s *RelayerTestSuite) Test_10_BatchedAckPacketToEth_Groth16() {
	ctx := context.Background()
	s.ICS20TransferERC20TokenBatchedAckToEthTest(ctx, operator.ProofTypeGroth16, 10, big.NewInt(testvalues.TransferAmount), nil)
}

func (s *RelayerTestSuite) Test_5_BatchedAckPacketToEth_Plonk() {
	ctx := context.Background()
	s.ICS20TransferERC20TokenBatchedAckToEthTest(ctx, operator.ProofTypePlonk, 5, big.NewInt(testvalues.TransferAmount), []uint64{1, 2, 3, 4, 5})
}

func (s *RelayerTestSuite) Test_10_FilteredBatchedAckPacketToEth_Groth16() {
	ctx := context.Background()
	s.ICS20TransferERC20TokenBatchedAckToEthTest(ctx, operator.ProofTypeGroth16, 10, big.NewInt(testvalues.TransferAmount), []uint64{2, 6})
}

// Note that the relayer still only relays one tx, the batching is done
// on the cosmos transaction itself. So that it emits multiple IBC events.
func (s *RelayerTestSuite) ICS20TransferERC20TokenBatchedAckToEthTest(
	ctx context.Context, proofType operator.SupportedProofType, numOfTransfers int, transferAmount *big.Int, ackFilter []uint64,
) {
	s.Require().GreaterOrEqual(numOfTransfers, len(ackFilter))
	s.Require().Greater(numOfTransfers, 0)

	s.SetupSuite(ctx, proofType)

	eth, simd := s.EthChain, s.CosmosChains[0]

	ics26Address := ethcommon.HexToAddress(s.contractAddresses.Ics26Router)
	ics20Address := ethcommon.HexToAddress(s.contractAddresses.Ics20Transfer)
	erc20Address := ethcommon.HexToAddress(s.contractAddresses.Erc20)

	totalTransferAmount := new(big.Int).Mul(transferAmount, big.NewInt(int64(numOfTransfers)))
	ethereumUserAddress := crypto.PubkeyToAddress(s.key.PublicKey)
	cosmosUserWallet := s.CosmosUsers[0]
	cosmosUserAddress := cosmosUserWallet.FormattedAddress()

	ics20transferAbi, err := abi.JSON(strings.NewReader(ics20transfer.ContractABI))
	s.Require().NoError(err)

	s.Require().True(s.Run("Approve the ICS20Transfer.sol contract to spend the erc20 tokens", func() {
		tx, err := s.erc20Contract.Approve(s.GetTransactOpts(s.key, eth), ics20Address, totalTransferAmount)
		s.Require().NoError(err)

		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)
		s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)

		allowance, err := s.erc20Contract.Allowance(nil, ethereumUserAddress, ics20Address)
		s.Require().NoError(err)
		s.Require().Equal(totalTransferAmount, allowance)
	}))

	var (
		sendTxHashes  [][]byte
		escrowAddress ethcommon.Address
	)
	s.Require().True(s.Run(fmt.Sprintf("Send %d transfers on Ethereum", numOfTransfers), func() {
		timeout := uint64(time.Now().Add(30 * time.Minute).Unix())
		transferMulticall := make([][]byte, numOfTransfers)

		msgSendPacket := ics20transfer.IICS20TransferMsgsSendTransferMsg{
			SourceClient:     testvalues.CustomClientID,
			Denom:            erc20Address,
			Amount:           transferAmount,
			Receiver:         cosmosUserAddress,
			TimeoutTimestamp: timeout,
			Memo:             "",
		}

		encodedMsg, err := ics20transferAbi.Pack("sendTransfer", msgSendPacket)
		s.Require().NoError(err)
		for i := range numOfTransfers {
			transferMulticall[i] = encodedMsg
		}

		tx, err := s.ics20Contract.Multicall(s.GetTransactOpts(s.key, eth), transferMulticall)
		s.Require().NoError(err)

		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)
		s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)
		s.T().Logf("Multicall send %d transfers gas used: %d", numOfTransfers, receipt.GasUsed)
		sendTxHashes = append(sendTxHashes, tx.Hash().Bytes())

		s.True(s.Run("Verify balances on Ethereum", func() {
			// User balance on Ethereum
			userBalance, err := s.erc20Contract.BalanceOf(nil, ethereumUserAddress)
			s.Require().NoError(err)
			s.Require().Equal(new(big.Int).Sub(testvalues.StartingERC20Balance, totalTransferAmount), userBalance)

			// Get the escrow address
			escrowAddress, err = s.ics20Contract.GetEscrow(nil, testvalues.CustomClientID)
			s.Require().NoError(err)

			// ICS20 contract balance on Ethereum
			escrowBalance, err := s.erc20Contract.BalanceOf(nil, escrowAddress)
			s.Require().NoError(err)
			s.Require().Equal(totalTransferAmount, escrowBalance)
		}))
	}))

	var ackTxHash []byte
	s.Require().True(s.Run("Receive packets on Cosmos chain", func() {
		var relayTxBodyBz []byte
		s.Require().True(s.Run("Retrieve relay tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:    eth.ChainID.String(),
				DstChain:    simd.Config().ChainID,
				SourceTxIds: sendTxHashes,
				SrcClientId: testvalues.CustomClientID,
				DstClientId: testvalues.FirstWasmClientID,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Empty(resp.Address)

			relayTxBodyBz = resp.Tx
		}))

		s.Require().True(s.Run("Broadcast relay tx", func() {
			resp := s.BroadcastSdkTxBody(ctx, simd, s.SimdRelayerSubmitter, 2_000_000, relayTxBodyBz)

			ackTxHash, err = hex.DecodeString(resp.TxHash)
			s.Require().NoError(err)
			s.Require().NotEmpty(ackTxHash)
		}))

		s.Require().True(s.Run("Verify balances on Cosmos chain", func() {
			denomOnCosmos := transfertypes.NewDenom(s.contractAddresses.Erc20, transfertypes.NewHop(transfertypes.PortID, testvalues.FirstWasmClientID))

			// User balance on Cosmos chain
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simd, &banktypes.QueryBalanceRequest{
				Address: cosmosUserAddress,
				Denom:   denomOnCosmos.IBCDenom(),
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(totalTransferAmount, resp.Balance.Amount.BigInt())
			s.Require().Equal(denomOnCosmos.IBCDenom(), resp.Balance.Denom)
		}))
	}))

	s.Require().True(s.Run("Acknowledge packets on Ethereum", func() {
		s.Require().True(s.Run("Verify commitment exists", func() {
			for i := range numOfTransfers {
				seq := uint64(i) + 1
				packetCommitmentPath := ibchostv2.PacketCommitmentKey(testvalues.CustomClientID, seq)
				var ethPath [32]byte
				copy(ethPath[:], crypto.Keccak256(packetCommitmentPath))

				resp, err := s.ics26Contract.GetCommitment(nil, ethPath)
				s.Require().NoError(err)
				s.Require().NotZero(resp)
			}
		}))

		var relayTx []byte
		s.Require().True(s.Run("Retrieve relay tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:           simd.Config().ChainID,
				DstChain:           s.EthChain.ChainID.String(),
				SourceTxIds:        [][]byte{ackTxHash},
				SrcClientId:        testvalues.FirstWasmClientID,
				DstClientId:        testvalues.CustomClientID,
				DstPacketSequences: ackFilter,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Equal(resp.Address, ics26Address.String())

			relayTx = resp.Tx
		}))

		s.Require().True(s.Run("Submit relay tx", func() {
			receipt, err := eth.BroadcastTx(ctx, s.EthRelayerSubmitter, 5_000_000, ics26Address, relayTx)
			s.Require().NoError(err)
			s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)

			// Verify the ack packet event exists
			_, err = e2esuite.GetEvmEvent(receipt, s.ics26Contract.ParseAckPacket)
			s.Require().NoError(err)
		}))

		s.Require().True(s.Run("Verify commitment removed", func() {
			for _, seq := range ackFilter {
				packetCommitmentPath := ibchostv2.PacketCommitmentKey(testvalues.CustomClientID, seq)
				var ethPath [32]byte
				copy(ethPath[:], crypto.Keccak256(packetCommitmentPath))

				resp, err := s.ics26Contract.GetCommitment(nil, ethPath)
				s.Require().NoError(err)
				s.Require().Zero(resp)
			}
		}))
	}))
}

func (s *RelayerTestSuite) TestMultiPeriodClientUpdateToCosmos() {
	ctx := context.Background()

	s.SetupSuite(ctx, operator.ProofTypeGroth16) // Doesn't matter, since we won't relay to eth in this test

	eth, simd := s.EthChain, s.CosmosChains[0]

	ics20Address := ethcommon.HexToAddress(s.contractAddresses.Ics20Transfer)
	erc20Address := ethcommon.HexToAddress(s.contractAddresses.Erc20)

	ethereumUserAddress := crypto.PubkeyToAddress(s.key.PublicKey)
	cosmosUserWallet := s.CosmosUsers[0]
	cosmosUserAddress := cosmosUserWallet.FormattedAddress()

	transferAmount := big.NewInt(testvalues.TransferAmount)

	s.Require().True(s.Run("Approve the ICS20Transfer.sol contract to spend the erc20 tokens", func() {
		tx, err := s.erc20Contract.Approve(s.GetTransactOpts(s.key, eth), ics20Address, transferAmount)
		s.Require().NoError(err)

		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)
		s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)
	}))

	s.Require().True(s.Run("Wait for period to change twice on Eth before relaying", func() {
		_, ethClientState := s.GetEthereumClientState(ctx, simd, testvalues.FirstWasmClientID)
		clientPeriod := ethClientState.LatestSlot / (ethClientState.EpochsPerSyncCommitteePeriod * ethClientState.SlotsPerEpoch)
		err := testutil.WaitForCondition(time.Minute*30, time.Second*30, func() (bool, error) {
			finalityUpdate, err := eth.BeaconAPIClient.GetFinalityUpdate()
			if err != nil {
				return false, err
			}

			finalitySlot, err := strconv.Atoi(finalityUpdate.Data.FinalizedHeader.Beacon.Slot)
			if err != nil {
				return false, err
			}
			finalityPeriod := uint64(finalitySlot) / (ethClientState.EpochsPerSyncCommitteePeriod * ethClientState.SlotsPerEpoch)

			return finalityPeriod == clientPeriod+2, nil
		})
		s.Require().NoError(err)
	}))

	var (
		sendTxHash    []byte
		escrowAddress ethcommon.Address
	)
	s.Require().True(s.Run("Send transfers on Ethereum", func() {
		timeout := uint64(time.Now().Add(30 * time.Minute).Unix())

		msgSendTransfer := ics20transfer.IICS20TransferMsgsSendTransferMsg{
			Denom:            erc20Address,
			SourceClient:     testvalues.CustomClientID,
			DestPort:         transfertypes.PortID,
			Amount:           transferAmount,
			Receiver:         cosmosUserAddress,
			TimeoutTimestamp: timeout,
			Memo:             "",
		}

		tx, err := s.ics20Contract.SendTransfer(s.GetTransactOpts(s.key, eth), msgSendTransfer)
		s.Require().NoError(err)

		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)
		s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)

		sendTxHash = tx.Hash().Bytes()

		s.True(s.Run("Verify balances on Ethereum", func() {
			// User balance on Ethereum
			userBalance, err := s.erc20Contract.BalanceOf(nil, ethereumUserAddress)
			s.Require().NoError(err)
			s.Require().Equal(new(big.Int).Sub(testvalues.StartingERC20Balance, transferAmount), userBalance)

			// Get the escrow address
			escrowAddress, err = s.ics20Contract.GetEscrow(nil, testvalues.CustomClientID)
			s.Require().NoError(err)

			// ICS20 contract balance on Ethereum
			escrowBalance, err := s.erc20Contract.BalanceOf(nil, escrowAddress)
			s.Require().NoError(err)
			s.Require().Equal(transferAmount, escrowBalance)
		}))
	}))

	s.Require().True(s.Run("Receive packets on Cosmos chain", func() {
		var relayTxBodyBz []byte
		s.Require().True(s.Run("Retrieve relay tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:    eth.ChainID.String(),
				DstChain:    simd.Config().ChainID,
				SourceTxIds: [][]byte{sendTxHash},
				SrcClientId: testvalues.CustomClientID,
				DstClientId: testvalues.FirstWasmClientID,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Empty(resp.Address)

			relayTxBodyBz = resp.Tx

			s.wasmFixtureGenerator.AddFixtureStep("receive_packets", ethereumtypes.RelayerMessages{
				RelayerTxBody: hex.EncodeToString(relayTxBodyBz),
			})
		}))

		var ackTxHash []byte
		s.Require().True(s.Run("Broadcast relay tx", func() {
			resp := s.BroadcastSdkTxBody(ctx, simd, s.SimdRelayerSubmitter, 2_000_000, relayTxBodyBz)

			var err error
			ackTxHash, err = hex.DecodeString(resp.TxHash)
			s.Require().NoError(err)
			s.Require().NotEmpty(ackTxHash)
		}))

		s.Require().True(s.Run("Verify balances on Cosmos chain", func() {
			denomOnCosmos := transfertypes.NewDenom(s.contractAddresses.Erc20, transfertypes.NewHop(transfertypes.PortID, testvalues.FirstWasmClientID))

			// User balance on Cosmos chain
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simd, &banktypes.QueryBalanceRequest{
				Address: cosmosUserAddress,
				Denom:   denomOnCosmos.IBCDenom(),
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(transferAmount, resp.Balance.Amount.BigInt())
			s.Require().Equal(denomOnCosmos.IBCDenom(), resp.Balance.Denom)
		}))
	}))
}

func (s *RelayerTestSuite) Test_Electra_Fork() {
	if os.Getenv(testvalues.EnvKeyEthTestnetType) != testvalues.EthTestnetTypePoS {
		s.T().Skip("Test is only relevant for PoS networks")
	}

	ctx := context.Background()

	electraForkEpoch := 12
	chainconfig.KurtosisConfig.NetworkParams.ElectraForkEpoch = uint64(electraForkEpoch)

	s.SetupSuite(ctx, operator.ProofTypeGroth16) // Doesn't matter, since we won't relay to eth in this test

	s.FilteredRecvPacketToCosmosTest(ctx, 1, big.NewInt(testvalues.TransferAmount), nil)

	spec, err := s.EthChain.BeaconAPIClient.GetSpec()
	s.Require().NoError(err)
	err = testutil.WaitForCondition(time.Minute*30, time.Second*30, func() (bool, error) {
		finalityUpdate, err := s.EthChain.BeaconAPIClient.GetFinalityUpdate()
		if err != nil {
			return false, err
		}

		finalitySlot, err := strconv.Atoi(finalityUpdate.Data.FinalizedHeader.Beacon.Slot)
		if err != nil {
			return false, err
		}

		latestFinalityEpoch := uint64(finalitySlot) / spec.SlotsPerEpoch

		fmt.Printf("Waiting for epoch %d, current epoch: %d\n", electraForkEpoch, latestFinalityEpoch)
		return latestFinalityEpoch >= uint64(electraForkEpoch), nil
	})
	s.Require().NoError(err)

	s.FilteredRecvPacketToCosmosTest(ctx, 1, big.NewInt(testvalues.TransferAmount), nil)
}

func (s *RelayerTestSuite) Test_10_RecvPacketToCosmos() {
	ctx := context.Background()
	s.SetupSuite(ctx, operator.ProofTypeGroth16) // Doesn't matter, since we won't relay to eth in this test
	s.FilteredRecvPacketToCosmosTest(ctx, 10, big.NewInt(testvalues.TransferAmount), nil)
}

func (s *RelayerTestSuite) Test_10_FilteredRecvPacketToCosmos() {
	ctx := context.Background()
	s.SetupSuite(ctx, operator.ProofTypeGroth16) // Doesn't matter, since we won't relay to eth in this test
	s.FilteredRecvPacketToCosmosTest(ctx, 10, big.NewInt(testvalues.TransferAmount), []uint64{2, 6})
}

func (s *RelayerTestSuite) FilteredRecvPacketToCosmosTest(ctx context.Context, numOfTransfers int, transferAmount *big.Int, recvFilter []uint64) {
	s.Require().GreaterOrEqual(numOfTransfers, len(recvFilter))
	s.Require().Greater(numOfTransfers, 0)

	eth, simd := s.EthChain, s.CosmosChains[0]

	ics20Address := ethcommon.HexToAddress(s.contractAddresses.Ics20Transfer)
	erc20Address := ethcommon.HexToAddress(s.contractAddresses.Erc20)

	totalTransferAmount := new(big.Int).Mul(transferAmount, big.NewInt(int64(numOfTransfers)))
	var relayedAmount *big.Int
	if len(recvFilter) == 0 {
		relayedAmount = totalTransferAmount
	} else {
		relayedAmount = new(big.Int).Mul(transferAmount, big.NewInt(int64(len(recvFilter))))
	}
	ethereumUserAddress := crypto.PubkeyToAddress(s.key.PublicKey)
	cosmosUserWallet := s.CosmosUsers[0]
	cosmosUserAddress := cosmosUserWallet.FormattedAddress()
	escrowAddress, err := s.ics20Contract.GetEscrow(nil, testvalues.CustomClientID)
	s.Require().NoError(err)

	denomOnCosmos := transfertypes.NewDenom(s.contractAddresses.Erc20, transfertypes.NewHop(transfertypes.PortID, testvalues.FirstWasmClientID))

	escrowStartingBalance, err := s.erc20Contract.BalanceOf(nil, escrowAddress)
	s.Require().NoError(err)
	startBalanceEthereum, err := s.erc20Contract.BalanceOf(nil, ethereumUserAddress)
	s.Require().NoError(err)
	startBalanceCosmosResp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simd, &banktypes.QueryBalanceRequest{
		Address: cosmosUserAddress,
		Denom:   denomOnCosmos.IBCDenom(),
	})
	s.Require().NoError(err)
	startBalanceCosmos := startBalanceCosmosResp.Balance.Amount.BigInt()

	s.Require().True(s.Run("Approve the ICS20Transfer.sol contract to spend the erc20 tokens", func() {
		tx, err := s.erc20Contract.Approve(s.GetTransactOpts(s.key, eth), ics20Address, totalTransferAmount)
		s.Require().NoError(err)

		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)
		s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)

		allowance, err := s.erc20Contract.Allowance(nil, ethereumUserAddress, ics20Address)
		s.Require().NoError(err)
		s.Require().Equal(totalTransferAmount, allowance)
	}))

	var sendTxHashes [][]byte
	s.Require().True(s.Run(fmt.Sprintf("Send %d transfers on Ethereum", numOfTransfers), func() {
		timeout := uint64(time.Now().Add(30 * time.Minute).Unix())

		msgSendTransfer := ics20transfer.IICS20TransferMsgsSendTransferMsg{
			Denom:            erc20Address,
			SourceClient:     testvalues.CustomClientID,
			DestPort:         transfertypes.PortID,
			Amount:           transferAmount,
			Receiver:         cosmosUserAddress,
			TimeoutTimestamp: timeout,
			Memo:             "",
		}

		for range numOfTransfers {
			tx, err := s.ics20Contract.SendTransfer(s.GetTransactOpts(s.key, eth), msgSendTransfer)
			s.Require().NoError(err)

			receipt, err := eth.GetTxReciept(ctx, tx.Hash())
			s.Require().NoError(err)
			s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)

			sendTxHashes = append(sendTxHashes, tx.Hash().Bytes())
		}

		s.True(s.Run("Verify balances on Ethereum", func() {
			// User balance on Ethereum
			userBalance, err := s.erc20Contract.BalanceOf(nil, ethereumUserAddress)
			s.Require().NoError(err)
			s.Require().Equal(new(big.Int).Sub(startBalanceEthereum, totalTransferAmount), userBalance)

			// Get the escrow address
			escrowAddress, err = s.ics20Contract.GetEscrow(nil, testvalues.CustomClientID)
			s.Require().NoError(err)

			// ICS20 contract balance on Ethereum
			escrowBalance, err := s.erc20Contract.BalanceOf(nil, escrowAddress)
			s.Require().NoError(err)
			s.Require().Equal(new(big.Int).Add(escrowStartingBalance, totalTransferAmount), escrowBalance)
		}))
	}))

	s.Require().True(s.Run("Receive packets on Cosmos chain", func() {
		var relayTxBodyBz []byte
		s.Require().True(s.Run("Retrieve relay tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:           eth.ChainID.String(),
				DstChain:           simd.Config().ChainID,
				SourceTxIds:        sendTxHashes,
				SrcClientId:        testvalues.CustomClientID,
				DstClientId:        testvalues.FirstWasmClientID,
				SrcPacketSequences: recvFilter,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Empty(resp.Address)

			relayTxBodyBz = resp.Tx

			s.wasmFixtureGenerator.AddFixtureStep("receive_packets", ethereumtypes.RelayerMessages{
				RelayerTxBody: hex.EncodeToString(relayTxBodyBz),
			})
		}))

		var ackTxHash []byte
		s.Require().True(s.Run("Broadcast relay tx", func() {
			resp := s.BroadcastSdkTxBody(ctx, simd, s.SimdRelayerSubmitter, 2_000_000, relayTxBodyBz)

			var err error
			ackTxHash, err = hex.DecodeString(resp.TxHash)
			s.Require().NoError(err)
			s.Require().NotEmpty(ackTxHash)
		}))

		s.Require().True(s.Run("Verify balances on Cosmos chain", func() {
			// User balance on Cosmos chain
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simd, &banktypes.QueryBalanceRequest{
				Address: cosmosUserAddress,
				Denom:   denomOnCosmos.IBCDenom(),
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(new(big.Int).Add(startBalanceCosmos, relayedAmount), resp.Balance.Amount.BigInt())
			s.Require().Equal(denomOnCosmos.IBCDenom(), resp.Balance.Denom)
		}))
	}))
}

func (s *RelayerTestSuite) Test_10_BatchedAckPacketToCosmos() {
	ctx := context.Background()
	s.ICS20TransferERC20TokenBatchedFilteredAckToCosmosTest(ctx, operator.ProofTypeGroth16, 10, nil)
}

func (s *RelayerTestSuite) Test_10_BatchedFilteredAckPacketToCosmos() {
	ctx := context.Background()
	s.ICS20TransferERC20TokenBatchedFilteredAckToCosmosTest(ctx, operator.ProofTypeGroth16, 10, []uint64{1, 2, 8})
}

// Note that the relayer still only relays one tx, the batching is done
// on the cosmos transaction itself. So that it emits multiple IBC events.
// This test also tests the filtering of the acks by packet sequence
func (s *RelayerTestSuite) ICS20TransferERC20TokenBatchedFilteredAckToCosmosTest(
	ctx context.Context, proofType operator.SupportedProofType, numOfTransfers int, ackFilter []uint64,
) {
	s.Require().GreaterOrEqual(numOfTransfers, len(ackFilter))
	s.Require().Greater(numOfTransfers, 0)

	s.SetupSuite(ctx, proofType)

	eth, simd := s.EthChain, s.CosmosChains[0]

	ics26Address := ethcommon.HexToAddress(s.contractAddresses.Ics26Router)
	transferAmount := big.NewInt(testvalues.TransferAmount)
	totalTransferAmount := big.NewInt(testvalues.TransferAmount * int64(numOfTransfers))
	if totalTransferAmount.Int64() > testvalues.InitialBalance {
		s.FailNow("Total transfer amount exceeds the initial balance")
	}
	ethereumUserAddress := crypto.PubkeyToAddress(s.key.PublicKey)
	cosmosUserWallet := s.CosmosUsers[0]
	cosmosUserAddress := cosmosUserWallet.FormattedAddress()
	sendMemo := "batched ack to cosmos test memo"

	var (
		transferCoin sdk.Coin
		sendTxHashes [][]byte
	)
	s.Require().True(s.Run("Send transfers on Cosmos chain", func() {
		for range numOfTransfers {
			timeout := uint64(time.Now().Add(30 * time.Minute).Unix())
			transferCoin = sdk.NewCoin(simd.Config().Denom, sdkmath.NewIntFromBigInt(transferAmount))

			transferPayload := transfertypes.FungibleTokenPacketData{
				Denom:    transferCoin.Denom,
				Amount:   transferCoin.Amount.String(),
				Sender:   cosmosUserAddress,
				Receiver: strings.ToLower(ethereumUserAddress.Hex()),
				Memo:     sendMemo,
			}
			encodedPayload, err := transfertypes.EncodeABIFungibleTokenPacketData(&transferPayload)
			s.Require().NoError(err)

			payload := channeltypesv2.Payload{
				SourcePort:      transfertypes.PortID,
				DestinationPort: transfertypes.PortID,
				Version:         transfertypes.V1,
				Encoding:        transfertypes.EncodingABI,
				Value:           encodedPayload,
			}
			msgSendPacket := channeltypesv2.MsgSendPacket{
				SourceClient:     testvalues.FirstWasmClientID,
				TimeoutTimestamp: timeout,
				Payloads: []channeltypesv2.Payload{
					payload,
				},
				Signer: cosmosUserWallet.FormattedAddress(),
			}

			resp, err := s.BroadcastMessages(ctx, simd, cosmosUserWallet, 200_000, &msgSendPacket)
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.TxHash)

			txHash, err := hex.DecodeString(resp.TxHash)
			s.Require().NoError(err)

			sendTxHashes = append(sendTxHashes, txHash)
		}

		s.Require().True(s.Run("Verify balances on Cosmos chain", func() {
			// Check the balance of UserB
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simd, &banktypes.QueryBalanceRequest{
				Address: cosmosUserAddress,
				Denom:   transferCoin.Denom,
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(testvalues.InitialBalance-totalTransferAmount.Int64(), resp.Balance.Amount.Int64())
		}))
	}))

	var ackTxHash []byte
	s.Require().True(s.Run("Receive packets on Ethereum", func() {
		var multicallTx []byte
		s.Require().True(s.Run("Retrieve relay tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:    simd.Config().ChainID,
				DstChain:    s.EthChain.ChainID.String(),
				SourceTxIds: sendTxHashes,
				SrcClientId: testvalues.FirstWasmClientID,
				DstClientId: testvalues.CustomClientID,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Equal(resp.Address, ics26Address.String())

			multicallTx = resp.Tx
		}))

		s.Require().True(s.Run("Submit relay tx", func() {
			receipt, err := eth.BroadcastTx(ctx, s.EthRelayerSubmitter, 5_000_000, ics26Address, multicallTx)
			s.Require().NoError(err)
			s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status, fmt.Sprintf("Tx failed: %+v", receipt))

			ackTxHash = receipt.TxHash.Bytes()
		}))
	}))

	s.Require().True(s.Run("Acknowledge packets on Cosmos", func() {
		s.Require().True(s.Run("Verify commitments exists", func() {
			for i := range numOfTransfers {
				resp, err := e2esuite.GRPCQuery[channeltypesv2.QueryPacketCommitmentResponse](ctx, simd, &channeltypesv2.QueryPacketCommitmentRequest{
					ClientId: testvalues.FirstWasmClientID,
					Sequence: uint64(i) + 1,
				})
				s.Require().NoError(err)
				s.Require().NotEmpty(resp.Commitment)
			}
		}))

		var relayTxBodyBz []byte
		s.Require().True(s.Run("Retrieve relay tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:           s.EthChain.ChainID.String(),
				DstChain:           simd.Config().ChainID,
				SourceTxIds:        [][]byte{ackTxHash},
				SrcClientId:        testvalues.CustomClientID,
				DstClientId:        testvalues.FirstWasmClientID,
				DstPacketSequences: ackFilter,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Empty(resp.Address)

			relayTxBodyBz = resp.Tx
		}))

		s.Require().True(s.Run("Broadcast relay tx", func() {
			_ = s.BroadcastSdkTxBody(ctx, simd, s.SimdRelayerSubmitter, 2_000_000, relayTxBodyBz)
		}))

		s.Require().True(s.Run("Verify commitments removed", func() {
			for i := range numOfTransfers {
				seq := uint64(i) + 1
				resp, err := e2esuite.GRPCQuery[channeltypesv2.QueryPacketCommitmentResponse](ctx, simd, &channeltypesv2.QueryPacketCommitmentRequest{
					ClientId: testvalues.FirstWasmClientID,
					Sequence: seq,
				})
				if len(ackFilter) == 0 || slices.Contains(ackFilter, seq) {
					// If the sequence is in the filter, we expect the commitment to be removed
					s.Require().ErrorContains(err, "packet commitment hash not found")
				} else {
					// Otherwise, we expect the commitment to still exist
					s.Require().NoError(err)
					s.Require().NotEmpty(resp.Commitment)
				}

			}
		}))
	}))
}

// TestConcurrentRecvPacketToEth_Groth16 tests the concurrent relaying of 50 packets from Ethereum to Cosmos
func (s *RelayerTestSuite) Test_50_concurrent_RecvPacketToCosmosTest() {
	ctx := context.Background()
	s.ConcurrentRecvPacketToCosmos(ctx, operator.ProofTypeGroth16, 50)
}

func (s *RelayerTestSuite) ConcurrentRecvPacketToCosmos(
	ctx context.Context, proofType operator.SupportedProofType, numConcurrentTransfers int,
) {
	s.Require().Greater(numConcurrentTransfers, 0)

	s.SetupSuite(ctx, proofType)

	eth, simd := s.EthChain, s.CosmosChains[0]

	ics20Address := ethcommon.HexToAddress(s.contractAddresses.Ics20Transfer)
	erc20Address := ethcommon.HexToAddress(s.contractAddresses.Erc20)
	transferAmount := big.NewInt(testvalues.TransferAmount)
	totalTransferAmount := big.NewInt(testvalues.TransferAmount * int64(numConcurrentTransfers))
	if totalTransferAmount.Int64() > testvalues.InitialBalance {
		s.FailNow("Total transfer amount exceeds the initial balance")
	}
	ethereumUserAddress := crypto.PubkeyToAddress(s.key.PublicKey)
	cosmosUserWallet := s.CosmosUsers[0]
	cosmosUserAddress := cosmosUserWallet.FormattedAddress()

	s.Require().True(s.Run("Approve the ICS20Transfer.sol contract to spend the erc20 tokens", func() {
		tx, err := s.erc20Contract.Approve(s.GetTransactOpts(s.key, eth), ics20Address, totalTransferAmount)
		s.Require().NoError(err)

		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)
		s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)

		allowance, err := s.erc20Contract.Allowance(nil, ethereumUserAddress, ics20Address)
		s.Require().NoError(err)
		s.Require().Equal(totalTransferAmount, allowance)
	}))

	var (
		sendTxHashes  [][]byte
		escrowAddress ethcommon.Address
	)
	s.Require().True(s.Run(fmt.Sprintf("Send %d transfers on Ethereum", numConcurrentTransfers), func() {
		// Setting the timeout high to avoid timeouts on mainnet preset
		timeout := uint64(time.Now().Add(120 * time.Minute).Unix())

		msgSendPacket := ics20transfer.IICS20TransferMsgsSendTransferMsg{
			SourceClient:     testvalues.CustomClientID,
			Denom:            erc20Address,
			Amount:           transferAmount,
			Receiver:         cosmosUserAddress,
			TimeoutTimestamp: timeout,
			Memo:             "",
		}

		for range numConcurrentTransfers {
			tx, err := s.ics20Contract.SendTransfer(s.GetTransactOpts(s.key, eth), msgSendPacket)
			s.Require().NoError(err)

			receipt, err := eth.GetTxReciept(ctx, tx.Hash())
			s.Require().NoError(err)
			s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)
			sendTxHashes = append(sendTxHashes, tx.Hash().Bytes())
		}

		s.True(s.Run("Verify balances on Ethereum", func() {
			// User balance on Ethereum
			userBalance, err := s.erc20Contract.BalanceOf(nil, ethereumUserAddress)
			s.Require().NoError(err)
			s.Require().Equal(new(big.Int).Sub(testvalues.StartingERC20Balance, totalTransferAmount), userBalance)

			// Get the escrow address
			escrowAddress, err = s.ics20Contract.GetEscrow(nil, testvalues.CustomClientID)
			s.Require().NoError(err)

			// ICS20 contract balance on Ethereum
			escrowBalance, err := s.erc20Contract.BalanceOf(nil, escrowAddress)
			s.Require().NoError(err)
			s.Require().Equal(totalTransferAmount, escrowBalance)
		}))
	}))

	s.Require().True(s.Run("Relay the last packet", func() {
		relayResp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
			SrcChain:    eth.ChainID.String(),
			DstChain:    simd.Config().ChainID,
			SourceTxIds: [][]byte{sendTxHashes[len(sendTxHashes)-1]},
			SrcClientId: testvalues.CustomClientID,
			DstClientId: testvalues.FirstWasmClientID,
		})
		s.Require().NoError(err)
		s.Require().NotEmpty(relayResp.Tx)
		s.Require().Empty(relayResp.Address)

		relayTxBodyBz := relayResp.Tx

		_ = s.BroadcastSdkTxBody(ctx, simd, s.SimdRelayerSubmitter, 2_000_000, relayTxBodyBz)

		// Remove the last txHash from the list
		sendTxHashes = sendTxHashes[:len(sendTxHashes)-1]
	}))

	var eg errgroup.Group
	s.Require().True(s.Run("Relay all the remaining requests concurrently", func() {
		// to avoid the submitter from getting account sequence mismatch, we need to lock when submitting
		time.Sleep(10 * time.Second) // Just to make sure we are up to date

		var relayTxHashes [][]byte
		// loop over the txHashes and send them concurrently
		for _, txHash := range sendTxHashes {
			txHash := txHash
			eg.Go(func() error {
				resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
					SrcChain:    eth.ChainID.String(),
					DstChain:    simd.Config().ChainID,
					SourceTxIds: [][]byte{txHash},
					SrcClientId: testvalues.CustomClientID,
					DstClientId: testvalues.FirstWasmClientID,
				})
				if err != nil {
					return err
				}

				relayTxBodyBz := resp.Tx
				relayTxHashes = append(relayTxHashes, relayTxBodyBz)
				return nil
			})
		}

		s.Require().NoError(eg.Wait())

		// relay 10 packets at a time
		for i := 0; i < len(relayTxHashes); i += 10 {
			txHashes := relayTxHashes[i:min(i+10, len(relayTxHashes))]

			var msgs []sdk.Msg
			for _, txHash := range txHashes {
				var txBody txtypes.TxBody
				err := proto.Unmarshal(txHash, &txBody)
				s.Require().NoError(err)

				for _, msg := range txBody.Messages {
					// Make sure there are no update client messages
					s.Require().NotEqual("ibc.core.client.v1.MsgUpdateClient", msg.TypeUrl)

					var sdkMsg sdk.Msg
					err = simd.Config().EncodingConfig.InterfaceRegistry.UnpackAny(msg, &sdkMsg)
					s.Require().NoError(err)

					msgs = append(msgs, sdkMsg)
				}
			}
			s.Require().NotZero(len(msgs))

			_, err := s.BroadcastMessages(ctx, simd, s.SimdRelayerSubmitter, 5_000_000, msgs...)
			s.Require().NoError(err)
		}
	}))

	s.Require().True(s.Run("Verify balances on Cosmos chain", func() {
		denomOnCosmos := transfertypes.NewDenom(s.contractAddresses.Erc20, transfertypes.NewHop(transfertypes.PortID, testvalues.FirstWasmClientID))

		// User balance on Cosmos chain
		resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simd, &banktypes.QueryBalanceRequest{
			Address: cosmosUserAddress,
			Denom:   denomOnCosmos.IBCDenom(),
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp.Balance)
		s.Require().Equal(totalTransferAmount, resp.Balance.Amount.BigInt())
		s.Require().Equal(denomOnCosmos.IBCDenom(), resp.Balance.Denom)
	}))
}
