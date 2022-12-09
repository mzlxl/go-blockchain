package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// 交易信息
type Transaction struct {
	ID   []byte     // 交易ID
	Vin  []TXInput  // 交易输入
	Vout []TXOutput // 交易输出
}

// 交易输入
type TXInput struct {
	Txid      []byte // 交易ID
	Vout      int    // 存储的是该输出在那笔交易中所有输出的索引
	ScriptSig string // 提供了可解锁输出结构里面 ScriptPubKey 字段的数据（暂存转账用户地址）
}

// 交易输出，本次交易的输出可以看做是余额
type TXOutput struct {
	Value        int    // 交易数量
	ScriptPubKey string // 锁定脚本(暂存收款账户地址)
}

// 发送货币，将这个操作创建成一个交易，放到一个块里
// 然后有人挖出这个块，放到链上，这个人会活动这个交易对应的奖励
// from to可看做转账钱包地址
func NewTransaction(from, to string, amount int, bc *Blockchain) *Transaction {

	var inputs []TXInput
	var outputs []TXOutput

	// acc：此次消费可以用来花费的数量  validOutputs：此次消费可以用来花费的输出
	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("ERROR:not enough amount")
	}

	// 迭代可花费的输出集合，将所之前剩余的输出都花费掉，多余的通过找零的方式处理  key:交易ID  value:输出集合
	for txid, outs := range validOutputs {
		txID, _ := hex.DecodeString(txid)

		// 之前交易的输出（即剩下的余额），可作为这次交易的输入
		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	// 输出1：这是实际转移给接受者地址的输出
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		// 输出2：找零，只有当未花费输出超过新交易所需时产生
		outputs = append(outputs, TXOutput{acc - amount, from}) // a change
	}

	// 创建交易
	tx := Transaction{nil, inputs, outputs}
	// 填充交易ID
	tx.SetID()
	return &tx
}

// 创建奖励交易
func NewRewardTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

func (tx *Transaction) SetID() {
	tx.ID = tx.Hash()
}

// 将序列化结果做hash处理
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

// 序列化
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// 判断这笔属于是否属于我
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// 判断这笔交易是否属于我的
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// 是否是奖励交易
func (tx *Transaction) IsRewardTx() bool {
	return tx.ID == nil
}
