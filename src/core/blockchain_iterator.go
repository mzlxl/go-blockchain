package core

import "github.com/boltdb/bolt"

// 对数据库区块进行顺序迭代并打印
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// 为Blockchain创建一个迭代器，里面存储了当前迭代的块哈希（currentHash）和数据库的连接（db）
// 迭代器的初始状态为链中的 tip，因此区块将从尾到头迭代
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.Db}
	return bci
}

// 返回链中下一个块
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	_ = i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	i.currentHash = block.PreHash

	return block
}
