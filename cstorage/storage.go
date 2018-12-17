package cstorage

import (
	"github.com/bazo-blockchain/bazo-client/util"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

var (
	db     	*bolt.DB
	logger 	*log.Logger
	Buckets	[]string
)

const (
	ERROR_MSG = "Initiate storage aborted: "
	BLOCKHEADERS_BUCKET = "blockheaders"
	LASTBLOCKHEADER_BUCKET = "lastblockheader"
	MERKLEPROOF_BUCKET = "merkleproofs"
)

//Entry function for the storage package
func Init(dbname string) (err error) {
	logger = util.InitLogger()

	Buckets = []string {
		BLOCKHEADERS_BUCKET,
		LASTBLOCKHEADER_BUCKET,
		MERKLEPROOF_BUCKET,
	}

	db, err = bolt.Open(dbname, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		logger.Fatal(ERROR_MSG, err)
	}

	for _, bucket := range Buckets {
		err = storage.CreateBucket(bucket, db)
		if err != nil {
			return err
		}
	}

	return nil
}