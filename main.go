package main

import (
	"github.com/bazo-blockchain/bazo-client/REST"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-client/cstorage"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"os"
)

func main() {
	client.Init()

	if len(os.Args) > 1 {
		if os.Args[1] == "accTx" || os.Args[1] == "fundsTx" || os.Args[1] == "configTx" || os.Args[1] == "stakeTx" {
			client.ProcessTx(os.Args[1:])
		}

		return
	}

	//For querying an account state or starting the REST service, the client must establish a connection to the Bazo network.
	network.Init(p2p.CLIENT_PING)
	cstorage.Init("client.db")

	if len(os.Args) == 2 {
		client.ProcessState(os.Args[1])

		return
	}

	client.Sync()
	REST.Init()
}
