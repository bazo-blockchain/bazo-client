package main

import (
	"github.com/bazo-blockchain/bazo-client/cli"
	"github.com/bazo-blockchain/bazo-client/util"
	cli2 "github.com/urfave/cli"
	"os"
)

func main() {
	logger := util.InitLogger()

	app := cli2.NewApp()

	app.Name = "bazo-client"
	app.Usage = "the command line interface for interacting with the Bazo blockchain implemented in Go."
	app.Version = "1.0.0"
	app.Commands = []cli2.Command {
		cli.GetAccountCommand(logger),
		cli.GetFundsCommand(logger),
		cli.GetNetworkCommand(logger),
		cli.GetRestCommand(),
		cli.GetStakingCommand(logger),
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
