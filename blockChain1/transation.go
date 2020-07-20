package main

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
	"strings"
)

const subsidy int = 100

type Transation struct {
	ID   []byte //交易的哈希值
	Vin  []TXInput
	Vout []TXOutput
}

type TXInput struct {
	TXid      []byte//前一个交易的hash值，即父hash
	Voutindex int//前一个交易的索引号
	Signature []byte //私钥的签名
	Pubkey    []byte //公钥
}

type TXOutput struct {
	Value      int//交易数额
	PubkeyHash []byte //公钥的哈希
}

type TXOutputs struct {
	Outputs []TXOutput
}

func (outputs TXOutputs) Serialize() []byte{
	encoded := new(bytes.Buffer)
	enc := gob.NewEncoder(encoded)
	err := enc.Encode(outputs)
	if err != nil {
		panic(err)
	}
	return encoded.Bytes()
}

//根据比特币地址得到公钥哈希
func (out *TXOutput) Lock(address []byte) {
	decodeAddress, _ := Base58Decode(address)
	pubkeyhash := decodeAddress[1 : len(decodeAddress)-4]
	out.PubkeyHash = pubkeyhash
}

func (tx Transation) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("___Transaction %x", tx.ID))
	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("     Input： %d", i))
		lines = append(lines, fmt.Sprintf("     TXID: %x", input.TXid))
		lines = append(lines, fmt.Sprintf("     Out: %d", input.Voutindex))
		lines = append(lines, fmt.Sprintf("     Signature: %x", input.Signature))
	}
	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output： %d", i))
		lines = append(lines, fmt.Sprintf("     Value: %d", output.Value))
		lines = append(lines, fmt.Sprintf("     Script: %x", output.PubkeyHash))
	}
	return strings.Join(lines, "\n")
}

//序列化
func (tx Transation) Serialize() []byte {
	encoded := new(bytes.Buffer)
	enc := gob.NewEncoder(encoded)
	err := enc.Encode(tx)
	if err != nil {
		panic(err)
	}
	return encoded.Bytes()
}

func DeserializeOutputs(data []byte) TXOutputs{
	var outputs TXOutputs
	dec:= gob.NewDecoder(bytes.NewReader(data))
	err:= dec.Decode(&outputs)
	if err != nil {
		panic(err)
	}
	return outputs
}

func (tx *Transation) Hash() []byte {
	txcopy := *tx
	txcopy.ID = []byte{}
	hash := sha256.Sum256(txcopy.Serialize())
	return hash[:]
}

//根据金额与地址新建一个输出
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte (address))
	//txo.PubkeyHash = []byte(address)
	return txo
}

func NewCoinbaseTX(to, data string) *Transation {
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transation{nil, []TXInput{txin}, []TXOutput{*txout}} //[]TXInput{txin}, []TXOutput{*txout}
	tx.ID = tx.Hash()
	return &tx
}

func (out *TXOutput) CanBeUnlockedWith(pubkeyhash []byte) bool {
	return bytes.Compare(out.PubkeyHash, pubkeyhash) == 0
}

func (in *TXInput) BanBeUnlockedWith(unlockdata []byte) bool {
	lockinghash := HashPubkey(in.Pubkey)
	return bytes.Compare(lockinghash, unlockdata) == 0
}

func (tx Transation) IsCoinBase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].TXid) == 0 && tx.Vin[0].Voutindex == -1
}

func (tx Transation) Sign(prikey ecdsa.PrivateKey, prevTXs map[string]Transation) {
	if tx.IsCoinBase() {
		return
	}
	//检查过程
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.TXid)].ID == nil {
			log.Panic("ERROR")
		}
	}

	txcopy := tx.TrimmedCopy()
	for inID, vin := range txcopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.TXid)] //前一笔交易结构体
		txcopy.Vin[inID].Signature = nil
		txcopy.Vin[inID].Pubkey = prevTx.Vout[vin.Voutindex].PubkeyHash //这笔交易的这比输入的引用的前一笔交易的输出的公钥哈希值
		txcopy.ID = txcopy.Hash()
		r, s, err := ecdsa.Sign(rand.Reader, &prikey, txcopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature
	}
}

func (tx Transation) TrimmedCopy() Transation {
	var inputs []TXInput
	var outputs []TXOutput
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.TXid, vin.Voutindex, nil, nil})
	}
	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubkeyHash})
	}
	txCopy := Transation{tx.ID, inputs, outputs}
	return txCopy
}

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transation {
	var inputs []TXInput
	var outputs []TXOutput
	wallets, err := NewWallets()
	if err != nil {
		log.Panic()
	}
	wallet := wallets.GetWallet(from)

	acc, validoutputs := bc.FindSpendableOutputs(HashPubkey(wallet.PublicKey), amount)
	if acc < amount {
		log.Panic("Error:Not enough funds")
	}
	for txid, outs := range validoutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, []byte(from), wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	tx := Transation{nil, inputs, outputs}
	tx.ID = tx.Hash()
	bc.SignTransation(&tx, wallet.PrivateKey)
	return &tx
}
func (tx Transation) Verify(prevTxs map[string]Transation) bool {
	if tx.IsCoinBase() {
		return true
	}
	for _, vin := range tx.Vin {
		if prevTxs[hex.EncodeToString(vin.TXid)].ID == nil {
			log.Panic("ERROR")
		}
	}
	txcopy := tx.TrimmedCopy()
	//椭圆曲线
	curve := elliptic.P256()
	for inID,vin:= range tx.Vin{
		prevTx := prevTxs[hex.EncodeToString(vin.TXid)]
		txcopy.Vin[inID].Signature = nil
		txcopy.Vin[inID].Pubkey = prevTx.Vout[vin.Voutindex].PubkeyHash
		txcopy.ID = txcopy.Hash()

		r:= big.Int{}
		s:=big.Int{}

		siglen:= len(vin.Signature)
		r.SetBytes(vin.Signature[:siglen/2])
		s.SetBytes(vin.Signature[siglen/2:])

		x:=big.Int{}
		y:=big.Int{}

		keylen := len(vin.Pubkey)
		x.SetBytes(vin.Pubkey[:keylen/2])
		y.SetBytes(vin.Pubkey[keylen/2:])

		rawPubkey := ecdsa.PublicKey{curve,&x,&y}
		if  ecdsa.Verify(&rawPubkey,txcopy.ID,&r,&s) ==false {
			return false
		}

		txcopy.Vin[inID].Pubkey = nil
	}
	return true
}
