package cstorage

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/boltdb/bolt"
)

func ReadBlockHeader(hash [32]byte) (header *protocol.Block, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKHEADERS_BUCKET))
		encodedHeader := b.Get(hash[:])
		header = header.Decode(encodedHeader)
		return nil
	})

	if err != nil {
		return nil, err
	}

	if header == nil {
		return nil, errors.New(fmt.Sprintf("header not found for hash %x\n", hash))
	}

	return header, nil
}

func ReadLastBlockHeader() (header *protocol.Block, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LASTBLOCKHEADER_BUCKET))
		cb := b.Cursor()
		_, encodedHeader := cb.First()
		header = header.Decode(encodedHeader)
		return nil
	})

	if err != nil {
		return nil, err
	}

	if header == nil {
		return nil, errors.New("last block header not found")
	}

	return header, nil
}

func ReadMerkleProof(hash [32]byte) (proof *protocol.MerkleProof, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MERKLEPROOF_BUCKET))
		encdoedProof := b.Get(hash[:])
		proof = proof.Decode(encdoedProof)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return proof, nil
}

func ReadMerkleProofs() (proofs []*protocol.MerkleProof, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MERKLEPROOF_BUCKET))

		return b.ForEach(func(k, v []byte) error {
			var proof *protocol.MerkleProof
			proofs = append(proofs, proof.Decode(v))
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return proofs, nil
}