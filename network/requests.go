package network

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

func BlockReq(blockHash []byte) error {
	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := p2p.BuildPacket(p2p.BLOCK_REQ, blockHash[:])
	sendData(p, packet)

	return nil
}

func BlockHeaderReq(blockHash []byte) error {
	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := p2p.BuildPacket(p2p.BLOCK_HEADER_REQ, blockHash[:])
	sendData(p, packet)

	return nil
}

func TxReq(txType uint8, txHash [32]byte) error {
	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := p2p.BuildPacket(txType, txHash[:])
	sendData(p, packet)

	return nil
}

func AccReq(root bool, address [64]byte) error {
	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	var packet []byte
	if root {
		packet = p2p.BuildPacket(p2p.ROOTACC_REQ, address[:])
	} else {
		packet = p2p.BuildPacket(p2p.ACC_REQ, address[:])
	}

	sendData(p, packet)

	return nil
}

func SendTx(dial string, tx protocol.Transaction, typeID uint8) (err error) {
	if conn := p2p.Connect(dial); conn != nil {
		packet := p2p.BuildPacket(typeID, tx.Encode())
		conn.Write(packet)

		header, payload, err := p2p.RcvData_(conn)
		if err != nil || header.TypeID == p2p.NOT_FOUND {
			err = errors.New(string(payload[:]))
		}
		conn.Close()

		return err
	}

	txHash := tx.Hash()
	return errors.New(fmt.Sprintf("Sending tx %x failed.", txHash[:8]))
}

func IntermediateNodesReq(blockHash [32]byte, txHash [32]byte) error {
	p := peers.getRandomPeer()
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	var data [][]byte
	data = append(data, blockHash[:])
	data = append(data, txHash[:])

	packet := p2p.BuildPacket(p2p.INTERMEDIATE_NODES_REQ, protocol.Encode(data, 32))
	sendData(p, packet)

	return nil
}

func neighborReq() {
	p := peers.getRandomPeer()
	if p == nil {
		logger.Print("Could not fetch a random peer.\n")
		return
	}

	packet := p2p.BuildPacket(p2p.NEIGHBOR_REQ, nil)
	sendData(p, packet)
}
