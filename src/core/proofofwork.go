package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64
)

// 目标（调整这个数字可以变更计算的难度，数字越小哈希的前置0越少）
const targetBits = 10

type ProofOfWork struct {
	block  *Block   // 区块
	target *big.Int // 目标
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	// 左移运算，target 的二进制位全部左移uint(256-targetBits)位
	// 也理解为 target 乘以 2 的 uint(256-targetBits) 次方
	target.Lsh(target, uint(256-targetBits))

	return &ProofOfWork{block, target}
}

//Run执行工作证明
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	fmt.Printf("proof of work run,block will contains transactions:%s\n", pow.block.Transactions)

	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		fmt.Printf("generate hash: %x  nonce:%d\n", hash, nonce)

		// 比较hash，小于目标值
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}

	return nonce, hash[:]
}

// 工作量计算
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PreHash,
			pow.block.SerializeTransactions(),
			Int2Hex(pow.block.Timestamp),
			Int2Hex(int64(targetBits)),
			Int2Hex(int64(nonce)),
		}, []byte{},
	)
	return data
}

func Int2Hex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// 验证hash
func (pow *ProofOfWork) Validate() bool {
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	var hashInt big.Int
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}
