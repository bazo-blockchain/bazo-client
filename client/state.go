package client

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-client/cstorage"
	"github.com/bazo-blockchain/bazo-client/network"
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

var (
	//All blockheaders of the whole chain
	blockHeaders []*protocol.Block

	activeParameters miner.Parameters

	UnsignedContractTx    = make(map[[32]byte]*protocol.ContractTx)
	UnsignedConfigTx = make(map[[32]byte]*protocol.ConfigTx)
	UnsignedFundsTx  = make(map[[32]byte]*protocol.FundsTx)
)

//Update allBlockHeaders to the latest header. Start listening to broadcasted headers after.
func SyncBeforeTx(address [64]byte) {
	loadBlockHeaders()
	incomingBlockHeaders(true)
	GetAccount(address)
}

func Sync() {
	loadBlockHeaders()
	go incomingBlockHeaders(false)
}

func loadBlockHeaders() {
	var last *protocol.Block

	//youngest = fetchBlockHeader(nil)
	if last, _ = cstorage.ReadLastBlockHeader(); last != nil {
		var loaded []*protocol.Block
		loaded = loadDB(last, [32]byte{}, loaded)
		blockHeaders = append(blockHeaders, loaded...)
	}

	//The client is up to date with the network and can start listening for incoming headers.
	network.Uptodate = true
}

func incomingBlockHeaders(untilSynced bool) {
	for {
		blockHeaderIn := <-network.BlockHeaderIn

		var last *protocol.Block
		var lastHash [32]byte

		//Get the last header in the blockHeaders array. Its hash is relevant for appending the incoming header or the abort condition for recursive header fetching.
		if len(blockHeaders) > 0 {
			last = blockHeaders[len(blockHeaders)-1]
			lastHash = last.Hash
		} else {
			lastHash = [32]byte{}
		}

		//The incoming block header is already the last saved in the array.
		if blockHeaderIn.Hash == lastHash {
			continue
		}

		//The client is out of sync. Header cannot be appended to the array. The client must sync first.
		if last == nil || blockHeaderIn.PrevHash != lastHash {
			//Set the uptodate flag to false in order to avoid listening to new incoming block headers.
			network.Uptodate = false

			var loaded []*protocol.Block

			if last == nil || len(blockHeaders) <= 100 {
				blockHeaders = []*protocol.Block{}
				loaded = loadNetwork(blockHeaderIn, [32]byte{}, loaded)
			} else {
				//Remove the last 100 headers. This is precaution if the array contains rolled back blocks.
				blockHeaders = blockHeaders[:len(blockHeaders)-100]
				loaded = loadNetwork(blockHeaderIn, blockHeaders[len(blockHeaders)-1].Hash, loaded)
			}

			blockHeaders = append(blockHeaders, loaded...)
			cstorage.WriteLastBlockHeader(blockHeaders[len(blockHeaders)-1])

			network.Uptodate = true
		} else if blockHeaderIn.PrevHash == lastHash {
			saveAndLogBlockHeader(blockHeaderIn)

			blockHeaders = append(blockHeaders, blockHeaderIn)
			cstorage.WriteLastBlockHeader(blockHeaderIn)

			if untilSynced {
				return
			}
		}
	}
}

func fetchBlockHeader(blockHash []byte) (blockHeader *protocol.Block) {
	var errormsg string
	if blockHash != nil {
		errormsg = fmt.Sprintf("Loading header %x failed: ", blockHash[:8])
	}

	err := network.BlockHeaderReq(blockHash[:])
	if err != nil {
		logger.Println(errormsg + err.Error())
		return nil
	}

	blockHeaderI, err := network.Fetch(network.BlockHeaderChan)
	if err != nil {
		logger.Println(errormsg + err.Error())
		return nil
	}

	blockHeader = blockHeaderI.(*protocol.Block)

	logger.Printf("Fetch header with height %v\n", blockHeader.Height)

	return blockHeader
}

func loadDB(last *protocol.Block, abort [32]byte, loaded []*protocol.Block) []*protocol.Block {
	var ancestor *protocol.Block

	if last.PrevHash != abort {
		if ancestor, _ = cstorage.ReadBlockHeader(last.PrevHash); ancestor == nil {
			logger.Fatal()
		}

		loaded = loadDB(ancestor, abort, loaded)
	}

	logger.Printf("Header %x with height %v loaded from DB\n",
		last.Hash[:8],
		last.Height)

	loaded = append(loaded, last)

	return loaded
}

func loadNetwork(last *protocol.Block, abort [32]byte, loaded []*protocol.Block) []*protocol.Block {
	var ancestor *protocol.Block
	if ancestor = fetchBlockHeader(last.PrevHash[:]); ancestor == nil {
		for ancestor == nil {
			logger.Printf("Try to fetch header %x with height %v again\n", last.Hash[:8], last.Height)
			ancestor = fetchBlockHeader(last.PrevHash[:])
		}
	}

	if last.PrevHash != abort {
		loaded = loadNetwork(ancestor, abort, loaded)
	}

	saveAndLogBlockHeader(last)

	loaded = append(loaded, last)

	return loaded
}

func saveAndLogBlockHeader(blockHeader *protocol.Block) {
	cstorage.WriteBlockHeader(blockHeader)
	logger.Printf("Header %x with height %v loaded from network\n",
		blockHeader.Hash[:8],
		blockHeader.Height)
}

func getState(acc *Account, lastTenTx []*FundsTxJson) (err error) {
	//Get blocks if the Acc address:
	//* sent funds
	//* received funds
	//* is block's beneficiary
	//* nr of configTx in block is > 0 (in order to maintain params in light-client)

	relevantHeadersBeneficiary, relevantHeadersConfigBF := getRelevantBlockHeaders(acc.Address)

	acc.Balance += activeParameters.Block_reward * uint64(len(relevantHeadersBeneficiary))

	relevantBlocks, err := getRelevantBlocks(relevantHeadersConfigBF)
	for _, block := range relevantBlocks {
		if block == nil {
			continue
		}

		err = updateConfigParameters(block)
		if err != nil {
			return err
		}

		// Check if bloomfilter returns false, if yes, the block has nothing related to account's address
		if !block.BloomFilter.Test(acc.Address[:]) {
			continue
		}

		fundsTxs, err := requestFundsTx(block)

		// Check if it's a block with a Bloomfilter that returns false positive
		if len(fundsTxs) == 0 && block.Beneficiary != acc.Address {
			// TODO @rmnblm
		}

		err = balanceFunds(fundsTxs, block, acc, lastTenTx)
		if err != nil {
			return err
		}

		if block.Beneficiary == acc.Address {
			acc.Balance += block.TotalFees
		}
	}

	return nil
}

func requestFundsTx(block *protocol.Block) (fundsTxs []*protocol.FundsTx, err error) {
	for _, txHash := range block.FundsTxData {
		err := network.TxReq(p2p.FUNDSTX_REQ, txHash)
		if err != nil {
			return nil, err
		}

		txI, err := network.Fetch(network.FundsTxChan)
		if err != nil {
			return nil, err
		}

		fundsTx := txI.(*protocol.FundsTx)
		fundsTxs = append(fundsTxs, fundsTx)
	}

	return fundsTxs, nil
}

func balanceFunds(fundsTxs []*protocol.FundsTx, block *protocol.Block, acc *Account, lastTenTx []*FundsTxJson) error {
	bucket := protocol.NewTxBucket(acc.Address)

	for _, fundsTx := range fundsTxs {
		if fundsTx.From == acc.Address || fundsTx.To == acc.Address {
			bucket.AddFundsTx(fundsTx)
		}
	}

	bucketHash := bucket.Hash()
	if err := validateBucket(block, bucketHash); err != nil {
		return err
	}

	for _, fundsTx := range bucket.Transactions {
		// Check if account is sender of a transaction
		if fundsTx.From == acc.Address {
			//If Acc is no root, balance funds
			if !acc.IsRoot {
				acc.Balance -= fundsTx.Amount
				acc.Balance -= fundsTx.Fee
			}
			acc.TxCnt += 1
		}

		if fundsTx.To == acc.Address {
			acc.Balance += fundsTx.Amount
			put(lastTenTx, ConvertFundsTx(fundsTx, "verified"))
		}
	}

	// Create the Merkle proof for this block
	merkleTree := block.BuildMerkleTree()
	mhashes, err := merkleTree.MerkleProof(bucketHash)
	if err != nil {
		return err
	}

	proof := protocol.NewMerkleProof(
		block.Height,
		mhashes,
		bucket.Address,
		bucket.RelativeBalance,
		bucket.CalculateMerkleRoot())

	err = cstorage.WriteMerkleProof(&proof)
	if err != nil {
		return err
	}

	logger.Printf("Merkle proof written to client storage for tx at block height %v", block.Height)

	return nil
}

func updateConfigParameters(block *protocol.Block) error {
	for _, txHash := range block.ConfigTxData {
		err := network.TxReq(p2p.CONFIGTX_REQ, txHash)
		if err != nil {
			return err
		}

		txI, err := network.Fetch(network.ConfigTxChan)
		if err != nil {
			return err
		}

		tx := txI.(protocol.Transaction)
		configTx := txI.(*protocol.ConfigTx)

		//Validate tx
		if err := validateTx(block, tx, txHash); err != nil {
			return err
		}

		configTxSlice := []*protocol.ConfigTx{configTx}
		miner.CheckAndChangeParameters(&activeParameters, &configTxSlice)
	}

	return nil
}