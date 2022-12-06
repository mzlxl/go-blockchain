package core

import (
	"encoding/hex"
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
	ScriptSig string // 提供了可解锁输出结构里面 ScriptPubKey 字段的数据。如果 ScriptSig 提供的数据是正确的，那么输出就会被解锁
}

// 交易输出
type TXOutput struct {
	Value        int    // 交易数量
	ScriptPubKey string // 锁定脚本(ScriptPubKey)，要花这笔钱，必须要解锁该脚本
}

// 发送货币，将这个操作创建成一个交易，放到一个块里
// 然后有人挖出这个块，放到链上，这个人会活动这个交易对应的奖励
func NewTransaction(from, to string, amount int, bc *Blockchain) *Transaction {

	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("ERROR:not enough amount")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, _ := hex.DecodeString(txid)

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

func (tx *Transaction) SetID() {
	tx.ID = nil
}

func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// 判断是否是创世块交易
func (tx *Transaction) IsGenesisTx() bool {
	return tx.ID == nil
}
