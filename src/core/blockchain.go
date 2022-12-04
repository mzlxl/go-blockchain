package core

import "github.com/boltdb/bolt"

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

type Blockchain struct {
	// 不在里面存储所有的区块了，而是仅存储区块链的 tip
	// Blocks []*Block
	tip []byte
	// 存储数据库连接，一旦打开，就要一直运行到程序结束
	Db *bolt.DB
}

// 添加数据到链条
func (bc *Blockchain) AddBlock(data string) {
	lastHash := bc.getLastHash()
	newBlock := NewBlock(data, lastHash)
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
func NewBlockchain() *Blockchain {
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
			genesis := NewGenesisBlock()
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
