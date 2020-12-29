package relayer

// DONTCOVER

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	tmKv "github.com/tendermint/tendermint/libs/kv"
	tmLog "github.com/tendermint/tendermint/libs/log"
	tmClient "github.com/tendermint/tendermint/rpc/client/http"
	tmTypes "github.com/tendermint/tendermint/types"

	"github.com/Sifchain/sifnode/cmd/ebrelayer/txs"
	"github.com/Sifchain/sifnode/cmd/ebrelayer/types"

	ethbridge "github.com/Sifchain/sifnode/x/ethbridge/types"
)

const (
	// ReturnCethThreshold If more than 0.1 eth remaining after transaction
	ReturnCethThreshold = 10000000000000000
)

// TODO: Move relay functionality out of CosmosSub into a new Relayer parent struct

// CosmosSub defines a Cosmos listener that relays events to Ethereum and Cosmos
type CosmosSub struct {
	TmProvider              string
	EthProvider             string
	RegistryContractAddress common.Address
	PrivateKey              *ecdsa.PrivateKey
	CosmosContext           *types.CosmosContext
	Logger                  tmLog.Logger
}

// NewCosmosSub initializes a new CosmosSub
func NewCosmosSub(tmProvider, ethProvider string, registryContractAddress common.Address,
	privateKey *ecdsa.PrivateKey, cosmosContext *types.CosmosContext, logger tmLog.Logger) CosmosSub {
	return CosmosSub{
		TmProvider:              tmProvider,
		EthProvider:             ethProvider,
		RegistryContractAddress: registryContractAddress,
		PrivateKey:              privateKey,
		CosmosContext:           cosmosContext,
		Logger:                  logger,
	}
}

// Start a Cosmos chain subscription
func (sub CosmosSub) Start(completionEvent *sync.WaitGroup) {
	defer completionEvent.Done()
	time.Sleep(time.Second)
	client, err := tmClient.New(sub.TmProvider, "/websocket")
	if err != nil {
		sub.Logger.Error("failed to initialize a client", "err", err)
		completionEvent.Add(1)
		go sub.Start(completionEvent)
		return
	}
	client.SetLogger(sub.Logger)

	if err := client.Start(); err != nil {
		sub.Logger.Error("failed to start a client", "err", err)
		completionEvent.Add(1)
		go sub.Start(completionEvent)
		return
	}

	defer client.Stop() //nolint:errcheck

	// Subscribe to all tendermint transactions
	query := "tm.event = 'Tx'"
	out, err := client.Subscribe(context.Background(), "test", query, 1000)
	if err != nil {
		sub.Logger.Error("failed to subscribe to query", "err", err, "query", query)
		completionEvent.Add(1)
		go sub.Start(completionEvent)
		return
	}

	defer client.Unsubscribe(context.Background(), "test", query)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer close(quit)

	for {
		select {
		case result := <-out:
			tx, ok := result.Data.(tmTypes.EventDataTx)
			if !ok {
				sub.Logger.Error("new tx: error while extracting event data from new tx")
			}
			sub.Logger.Info("New transaction witnessed")

			// Iterate over each event in the transaction
			for _, event := range tx.Result.Events {
				claimType := getOracleClaimType(event.GetType())

				switch claimType {
				case types.MsgBurn, types.MsgLock:
					// Parse event data, then package it as a ProphecyClaim and relay to the Ethereum Network
					err := sub.handleBurnLockMsg(event.GetAttributes(), claimType)
					if err != nil {
						sub.Logger.Error(err.Error())
					}
				}
			}
		case <-quit:
			return
		}
	}
}

// getOracleClaimType sets the OracleClaim's claim type based upon the witnessed event type
func getOracleClaimType(eventType string) types.Event {
	var claimType types.Event
	switch eventType {
	case types.MsgBurn.String():
		claimType = types.MsgBurn
	case types.MsgLock.String():
		claimType = types.MsgLock
	default:
		claimType = types.Unsupported
	}
	return claimType
}

// Parses event data from the msg, event, builds a new ProphecyClaim, and relays it to Ethereum
func (sub CosmosSub) handleBurnLockMsg(attributes []tmKv.Pair, claimType types.Event) error {
	cosmosMsg, err := txs.BurnLockEventToCosmosMsg(claimType, attributes)
	if err != nil {
		fmt.Println(err)
		return err
	}
	sub.Logger.Info(cosmosMsg.String())

	// Only deal with submit message
	if cosmosMsg.MessageType != ethbridge.MsgSubmit {
		return nil
	}

	prophecyClaim := txs.CosmosMsgToProphecyClaim(cosmosMsg)
	gasUsed, err := txs.RelayProphecyClaimToEthereum(sub.EthProvider, sub.RegistryContractAddress,
		claimType, prophecyClaim, sub.PrivateKey, cosmosMsg.CethAmount.BigInt())

	// If failed to send prophecy claim to Ethereum
	if err != nil {
		fmt.Println(err)

		if err.Error() == "ceth paid not enough" {
			cosmosMsg.MessageType = ethbridge.MsgRevert
			txs.SendMsgToCosmos(sub.CosmosContext, &cosmosMsg)
		}

		return err
	}

	// return remaining gas to sifchain account
	cethAmount := cosmosMsg.CethAmount.BigInt()
	tempAmount := big.NewInt(ReturnCethThreshold)
	tempAmount.Add(tempAmount, big.NewInt(int64(gasUsed*3)))

	if cethAmount.Cmp(tempAmount) > 0 {
		cosmosMsg.MessageType = ethbridge.MsgReturnCeth
		cosmosMsg.CethAmount = cethAmount.Sub(big.NewInt(int64(gasUsed * 3)))
		txs.SendMsgToCosmos(sub.CosmosContext, &cosmosMsg)
	}

	return nil
}
