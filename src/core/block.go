package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	// 创建时间戳
	Timestamp int64
	// 区块数据（这里指的是交易信息）
	Transactions []*Transaction
	// 前一个区块hash
	PreHash []byte
	// 当前区块hash，用于校验区块数据有效性
	Hash []byte
	// 工作量
	Nonce int
}

// 创建block
func NewBlock(transactions []*Transaction, preHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, preHash, []byte{}, 0}

	// 挖矿过程，计算一个特殊的满足要求的数值
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// 创世纪区块
func NewGenesisBlock(genesis *Transaction) *Block {
	return NewBlock([]*Transaction{genesis}, []byte{})
}

// block序列化
// 选择 encoding/gob，是因为简单，而且也是go标准库的一部分
func (block *Block) SerializeBlock() []byte {
	// 定义一个 buffer 存储序列化之后的数据
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	// 序列化区块到buffer中
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// 反序列化block
func DeserializeBlock(data []byte) *Block {
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)

	// 反序列化到block
	var block Block
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}

// 序列化交易列表
func (block *Block) SerializeTransactions() []byte {
	// 定义一个 buffer 存储序列化之后的数据
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	// 序列化区块到buffer中
	err := encoder.Encode(block.Transactions)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// 对交易ID的集合进行sha256
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}
