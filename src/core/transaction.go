package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

// 交易信息
type Transaction struct {
	ID   []byte     // 交易ID
	Vin  []TXInput  // 交易输入
	Vout []TXOutput // 交易输出
}

// 发送货币，将这个操作创建成一个交易，放到一个块里
// 然后有人挖出这个块，放到链上，这个人会活动这个交易对应的奖励
// from to可看做转账钱包地址
func NewTransaction(wallet *Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {

	var inputs []TXInput
	var outputs []TXOutput

	// acc：此次消费可以用来花费的数量  validOutputs：此次消费可以用来花费的输出
	pubKeyHash := HashPubKey(wallet.PublicKey)
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("ERROR:not enough amount")
	}

	// 迭代可花费的输出集合，将所之前剩余的输出都花费掉，多余的通过找零的方式处理  key:交易ID  value:输出集合
	for txid, outs := range validOutputs {
		txID, _ := hex.DecodeString(txid)

		// 之前交易的输出（即剩下的余额），可作为这次交易的输入
		for _, out := range outs {
			input := NewTxin(txID, out, wallet.PublicKey)
			inputs = append(inputs, input)
		}
	}

	// 输出1：这是实际转移给接受者地址的输出
	from := fmt.Sprintf("%s", wallet.GetAddress())
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		// 输出2：找零，只有当未花费输出超过新交易所需时产生
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
	}

	// 创建交易
	tx := Transaction{nil, inputs, outputs}
	// 填充交易ID
	tx.SetID()

	UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)
	return &tx
}

// 创建奖励交易
// 奖励交易只有一个输出，输入的Txid 为空数组，Vout 等于 -1
func NewRewardTX(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}

	txin := NewRewardTxin(data)
	txout := *NewTXOutput(subsidy, to)
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

// 是否是奖励交易
func (tx *Transaction) IsRewardTx() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// 签名
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsRewardTx() {
		// 没有实际的输入，所以不需要签名
		return
	}

	txCopy := tx.TrimmedCopy()

	// 这个副本包含了所有的输入和输出，但是 TXInput.Signature 和 TXIput.PubKey 被设置为 nil
	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r, s, _ := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}
}

//创建一个副本
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, NewTxin(vin.Txid, vin.Vout, nil))
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}
