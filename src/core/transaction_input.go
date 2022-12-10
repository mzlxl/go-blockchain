package core

import "bytes"

// 交易输入
type TXInput struct {
	Txid      []byte // 交易ID
	Vout      int    // 存储的是该输出在那笔交易中所有输出的索引
	Signature []byte // 签名
	PubKey    []byte // 公钥
}

func NewTxin(Txid []byte, Vout int, PubKey []byte) TXInput {
	return TXInput{Txid, Vout, nil, PubKey}
}
func NewRewardTxin(data string) TXInput {
	return TXInput{[]byte{}, -1, nil, []byte(data)}
}

// 检查该地址是否发起了事务
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
