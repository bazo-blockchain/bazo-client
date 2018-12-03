package client

import (
	"github.com/bazo-blockchain/bazo-client/cstorage"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"log"
)

var (
	logger     *log.Logger
)

func Init() error {
	p2p.InitLogging()
	logger = util.InitLogger()

	util.Config = util.LoadConfiguration()

	network.Init()
	err := cstorage.Init("client.db")

	if err != nil {
		return err
	}

	return nil
}

func put(slice []*FundsTxJson, tx *FundsTxJson) {
	for i := 0; i < 9; i++ {
		slice[i] = slice[i+1]
	}

	slice[9] = tx
}
