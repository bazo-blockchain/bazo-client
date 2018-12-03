package cstorage

import "github.com/boltdb/bolt"

func DeleteBlockHeader(hash [32]byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKHEADERS_BUCKET))
		return b.Delete(hash[:])
	})
}
