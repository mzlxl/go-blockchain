package core

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisData = "genesis"

// 挖出新块的奖励金。在比特币中，实际并没有存储这个数字，而是基于区块总数进行计算而得
// 挖出创世块的奖励是 50 BTC，每挖出 210000 个块后，奖励减半
const subsidy = 10

type Blockchain struct {
	// 不在里面存储所有的区块了，而是仅存储区块链的 tip
	// Blocks []*Block
	tip []byte
	// 存储数据库连接，一旦打开，就要一直运行到程序结束
	Db *bolt.DB
}

// 添加数据到链条
func (bc *Blockchain) AddBlock(transactions []*Transaction) *Block {

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	lastHash := bc.getLastHash()
	newBlock := NewBlock(transactions, lastHash)
	bc.putBlock2Db(newBlock)
	return newBlock
}

// VerifyTransaction verifies transaction input signatures
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsRewardTx() {
		return true
	}

	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

// 获取数据库中最后一个区块的hash
func (bc *Blockchain) getLastHash() []byte {
	var lastHash []byte
	_ = bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("last"))
		return nil
	})

	return lastHash
}

func (bc *Blockchain) putBlock2Db(newBlock *Block) {
	_ = bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		_ = b.Put(newBlock.Hash, newBlock.SerializeBlock())
		_ = b.Put([]byte("last"), newBlock.Hash)
		bc.tip = newBlock.Hash
		return nil
	})
}

// 创建一个新的区块链条
// 数据库选择，BoltDB。理由：简单、go实现、不需要单独运行服务、keyvalue形式的字节数据存储
func NewBlockchain(address string) *Blockchain {
	var tip []byte
	// 打开一个数据库文件
	db, _ := bolt.Open(dbFile, 0600, nil)

	// 数据库操作通过一个事务（transaction）进行操作。有两种类型的事务：只读（read-only）和读写（read-write）
	// 打开一个读写事务（db.Update(...)），因为我们可能会向数据库中添加创世块
	_ = db.Update(func(tx *bolt.Tx) error {

		// 读取存储区块的bucket
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			// 创建存储区块的bucket，并将创世块保存进去
			gtx := NewRewardTX(address, genesisData)
			genesis := NewGenesisBlock(gtx)
			b, _ := tx.CreateBucket([]byte(blocksBucket))
			_ = b.Put(genesis.Hash, genesis.SerializeBlock())
			// last键存储链最后一个区块的hash，用于快捷获取PreHash
			_ = b.Put([]byte("last"), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("last"))
		}

		return nil
	})

	bc := &Blockchain{tip, db}

	return bc
}

func GetBlockchain() *Blockchain {
	if dbExists(dbFile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("last"))
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	bc := Blockchain{tip, db}
	return &bc
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// FindUTXO finds all unspent transaction outputs and returns transactions with spent outputs removed
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if tx.IsRewardTx() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PreHash) == 0 {
			break
		}
	}

	return UTXO
}

// SignTransaction signs inputs of a Transaction
func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

// FindTransaction finds a transaction by its ID
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PreHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found")
}
