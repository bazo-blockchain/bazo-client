package account

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/urfave/cli"
	"log"
	"os"
)

type createAccountArgs struct {
	header			int
	fee				int
	rootkeyFile		string
	file			string
}

func getCreateAccountCommand(logger *log.Logger) cli.Command {
	return cli.Command {
		Name: "create",
		Usage: "create a new account",
		Action: func(c *cli.Context) error {
			args := &createAccountArgs {
				header: 		c.Int("header"),
				fee: 			c.Int("fee"),
				rootkeyFile: 	c.String("rootkey"),
				file: 			c.String("file"),
			}

			return createAccount(args, logger)
		},
		Flags: []cli.Flag {
			headerFlag,
			feeFlag,
			rootkeyFlag,
			cli.StringFlag {
				Name: 	"file",
				Usage: 	"save new account's private key to `FILE`",
			},
		},
	}
}

func createAccount(args *createAccountArgs, logger *log.Logger) error {
	err := args.ValidateInput()
	if err != nil {
		return err
	}

	privKey, err := crypto.ExtractECDSAKeyFromFile(args.rootkeyFile)
	if err != nil {
		return err
	}

	var newKey *ecdsa.PrivateKey
	//Write the public key to the given textfile
	file, err := os.Create(args.file)
	if err != nil {
		return err
	}

	tx, newKey, err := protocol.ConstrAccTx(byte(args.header), uint64(args.fee), [64]byte{}, privKey, nil, nil)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(newKey.X.Text(16)) + "\n")
	_, err = file.WriteString(string(newKey.Y.Text(16)) + "\n")
	_, err = file.WriteString(string(newKey.D.Text(16)) + "\n")

	if err != nil {
		return errors.New(fmt.Sprintf("failed to write key to file %v", args.file))
	}

	return sendAccountTx(tx, logger)
}

func (args createAccountArgs) ValidateInput() error {
	if args.fee <= 0 {
		return errors.New("invalid argument: fee must be > 0")
	}

	if len(args.rootkeyFile) == 0 {
		return errors.New("argument missing: rootkeyFile")
	}

	if len(args.file) == 0 {
		return errors.New("argument missing: accountFile")
	}

	return nil
}