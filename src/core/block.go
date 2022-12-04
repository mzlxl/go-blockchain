package core

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	// 创建时间戳
	Timestamp int64
	// 区块数据
	Data []byte
	// 前一个区块hash
	PreHash []byte
	// 当前区块hash，用于校验区块数据有效性
	Hash []byte

	Nonce int
}

// 创建block
func NewBlock(data string, preHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), preHash, []byte{}, 0}

	// 挖矿过程，计算一个特殊的满足要求的数值
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// 创世纪区块
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
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
	err := decoder.Decode(block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
