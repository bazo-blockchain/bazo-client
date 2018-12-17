package client

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

//TODO Validate block merkle root.

func validateTx(block *protocol.Block, tx protocol.Transaction, txHash [32]byte) error {
	valid := true

	err := network.IntermediateNodesReq(block.Hash, txHash)
	if err != nil {
		//TODO
		valid = false
	}

	nodes, err := network.Fetch32Bytes(network.IntermediateNodesChan)
	if err != nil {
		//TODO
		valid = false
	}

	if txHash != tx.Hash() {
		valid = false
	}

	leafHash := txHash
	for i := 0; i < len(nodes); i += 2 {
		var parentHash [32]byte
		concatHash := append(leafHash[:], nodes[i][:]...)
		if parentHash = protocol.MTHash(concatHash); parentHash != nodes[i+1] {
			concatHash = append(nodes[i][:], leafHash[:]...)
			if parentHash = protocol.MTHash(concatHash); parentHash != nodes[i+1] {
				valid = false
			}
		}
		leafHash = parentHash
	}

	if !valid {
		return errors.New(fmt.Sprintf("Tx validation failed for %x\n", txHash))
	}

	return nil
}

func validateBucket(block *protocol.Block, bucketHash [32]byte) error {
	err := network.IntermediateNodesReq(block.Hash, bucketHash)
	if err != nil {
		return err
	}

	nodes, err := network.Fetch32Bytes(network.IntermediateNodesChan)
	if err != nil {
		return err
	}

	leafHash := bucketHash
	for i := 0; i < len(nodes); i += 2 {
		var parentHash [32]byte
		concatHash := append(leafHash[:], nodes[i][:]...)
		if parentHash = protocol.MTHash(concatHash); parentHash != nodes[i+1] {
			concatHash = append(nodes[i][:], leafHash[:]...)
			if parentHash = protocol.MTHash(concatHash); parentHash != nodes[i+1] {
				return errors.New(fmt.Sprintf("Bucket validation failed for %x\n", bucketHash))
			}
		}
		leafHash = parentHash
	}

	return nil
}
