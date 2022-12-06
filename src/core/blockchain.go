package core

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisData = "genesis"

// 挖出新块的奖励金。在比特币中，实际并没有存储这个数字，而是基于区块总数进行计算而得
const subsidy = 1

type Blockchain struct {
	// 不在里面存储所有的区块了，而是仅存储区块链的 tip
	// Blocks []*Block
	tip []byte
	// 存储数据库连接，一旦打开，就要一直运行到程序结束
	Db *bolt.DB
}

// 添加数据到链条
func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	lastHash := bc.getLastHash()
	newBlock := NewBlock(transactions, lastHash)
	bc.putBlock2Db(newBlock)
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
			gtx := NewGenesisTX(address, genesisData)
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

// 在区块链的最初，也就是第一个块，叫做创世块。正是这个创世块，产生了区块链最开始的输出。
//对于创世块，不需要引用之前的交易输出。因为在创世块之前根本不存在交易，也就没有不存在交易输出
func NewGenesisTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	// 创世块只有一个输出，Txid 为空数组，Vout 等于 -1
	// 交易也没有在 ScriptSig 中存储脚本，而只是存储了一个任意的字符串 data
	txin := TXInput{[]byte{}, -1, data}

	// 输出为挖矿奖励subsidy
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()
	return &tx
}

// 这个方法对所有的未花费交易进行迭代，并对它的值进行累加。
//当累加值大于或等于我们想要传送的值时，它就会停止并返回累加值，同时返回的还有通过交易 ID 进行分组的输出索引。我们只需取出足够支付的钱就够了
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
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
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsGenesisTx() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PreHash) == 0 {
			break
		}
	}

	return unspentTXs
}
