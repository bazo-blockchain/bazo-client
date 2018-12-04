package cstorage

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/boltdb/bolt"
)

func WriteBlockHeader(header *protocol.Block) (err error) {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKHEADERS_BUCKET))
		return b.Put(header.Hash[:], header.EncodeHeader())
	})
}

//Before saving the last block header, delete all existing entries.
func WriteLastBlockHeader(header *protocol.Block) (err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LASTBLOCKHEADER_BUCKET))
		return b.ForEach(func(k, v []byte) error {
			return b.Delete(k)
		})
	})

	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LASTBLOCKHEADER_BUCKET))
		return b.Put(header.Hash[:], header.EncodeHeader())
	})
}

func WriteMerkleProof(proof *protocol.MerkleProof) (err error) {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MERKLEPROOF_BUCKET))
		return b.Put(proof.Hash()[:], proof.Encode())
	})
}