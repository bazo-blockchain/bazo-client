package account

import (
	"errors"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-client/cstorage"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/urfave/cli"
	"log"
	"math/big"
)

type checkAccountArgs struct {
	address	string
	file	string
}

func getCheckAccountCommand(logger *log.Logger) cli.Command {
	return cli.Command {
		Name: "check",
		Usage: "check account state",
			Action: func(c *cli.Context) error {
			args := &checkAccountArgs {
				address:	c.String("address"),
				file:		c.String("file"),
			}

			return checkAccount(args, logger)
		},
		Flags: []cli.Flag {
			cli.StringFlag {
				Name: 	"address",
				Usage: 	"the account's 128 byte address",
			},
			cli.StringFlag {
				Name: 	"file",
				Usage: 	"load the account's 128 byte address from `FILE`",
				Value: 	"account.txt",
			},
		},
	}
}

func checkAccount(args *checkAccountArgs, logger *log.Logger) error {
	err := args.ValidateInput()
	if err != nil {
		return err
	}

	network.Init()
	cstorage.Init("client.db")

	var address [64]byte
	if len(args.address) == 128 {
		newPubInt, _ := new(big.Int).SetString(args.address, 16)
		copy(address[:], newPubInt.Bytes())
	} else {
		privKey, err := crypto.ExtractECDSAKeyFromFile(args.file)
		if err != nil {
			logger.Printf("%v\n", err)
			return err
		}

		address = crypto.GetAddressFromPubKey(&privKey.PublicKey)
	}

	logger.Printf("My address: %x\n", address)

	acc, _, err := client.CheckAccount(address)
	if err != nil {
		logger.Println(err)
		return err
	} else {
		logger.Printf(acc.String())
	}

	return nil
}

func (args checkAccountArgs) ValidateInput() error {
	if len(args.address) == 0 && len(args.file) == 0 {
		return errors.New("argument missing: address or file")
	}

	if len(args.file) == 0 && len(args.address) != 128 {
		return errors.New("invalid argument: address")
	}

	return nil
}