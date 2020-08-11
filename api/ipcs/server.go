// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ipcs

import (
	"fmt"
	"net/http"

	_ "go.nanomsg.org/mangos/v3/transport/ipc" // registers the IPC transport

	"github.com/gorilla/rpc/v2"

	"github.com/ava-labs/gecko/api"
	"github.com/ava-labs/gecko/chains"
	"github.com/ava-labs/gecko/ipcs"
	"github.com/ava-labs/gecko/snow/engine/common"
	"github.com/ava-labs/gecko/utils/json"
	"github.com/ava-labs/gecko/utils/logging"
)

// IPCServer maintains the IPCs
type IPCServer struct {
	httpServer   *api.Server
	chainManager chains.Manager
	log          logging.Logger
	ipcs         *ipcs.ChainIPCs
}

// NewService returns a new IPCs API service
func NewService(log logging.Logger, chainManager chains.Manager, httpServer *api.Server, ipcs *ipcs.ChainIPCs) (*common.HTTPHandler, error) {
	ipcServer := &IPCServer{
		log:          log,
		chainManager: chainManager,
		httpServer:   httpServer,

		ipcs: ipcs,
	}

	newServer := rpc.NewServer()
	codec := json.NewCodec()
	newServer.RegisterCodec(codec, "application/json")
	newServer.RegisterCodec(codec, "application/json;charset=UTF-8")
	newServer.RegisterService(ipcServer, "ipcs")
	return &common.HTTPHandler{Handler: newServer}, nil
}

// PublishBlockchainArgs are the arguments for calling PublishBlockchain
type PublishBlockchainArgs struct {
	BlockchainID string `json:"blockchainID"`
}

// PublishBlockchainReply are the results from calling PublishBlockchain
type PublishBlockchainReply struct {
	ConsensusURL string `json:"consensusURL"`
	DecisionsURL string `json:"decisionsURL"`
}

// PublishBlockchain publishes the finalized accepted transactions from the blockchainID over the IPC
func (ipc *IPCServer) PublishBlockchain(r *http.Request, args *PublishBlockchainArgs, reply *PublishBlockchainReply) error {
	ipc.log.Info("IPCs: PublishBlockchain called with BlockchainID: %s", args.BlockchainID)
	chainID, err := ipc.chainManager.Lookup(args.BlockchainID)
	if err != nil {
		ipc.log.Error("unknown blockchainID: %s", err)
		return err
	}

	ipcs, err := ipc.ipcs.Publish(chainID)
	if err != nil {
		ipc.log.Error("couldn't publish blockchainID: %s", err)
		return err
	}

	reply.ConsensusURL = ipcs.ConsensusURL()
	reply.DecisionsURL = ipcs.DecisionsURL()

	return nil
}

// UnpublishBlockchainArgs are the arguments for calling UnpublishBlockchain
type UnpublishBlockchainArgs struct {
	BlockchainID string `json:"blockchainID"`
}

// UnpublishBlockchainReply are the results from calling UnpublishBlockchain
type UnpublishBlockchainReply struct {
	Success bool `json:"success"`
}

// UnpublishBlockchain closes publishing of a blockchainID
func (ipc *IPCServer) UnpublishBlockchain(r *http.Request, args *UnpublishBlockchainArgs, reply *UnpublishBlockchainReply) error {
	ipc.log.Info("IPCs: UnpublishBlockchain called with BlockchainID: %s", args.BlockchainID)
	chainID, err := ipc.chainManager.Lookup(args.BlockchainID)
	if err != nil {
		ipc.log.Error("unknown blockchainID %s: %s", args.BlockchainID, err)
		return err
	}

	ok, err := ipc.ipcs.Unpublish(chainID)
	if !ok {
		return fmt.Errorf("blockchainID not publishing: %s", chainID)
	}

	reply.Success = true
	return err
}