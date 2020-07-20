package main

import (
	"bytes"
	//"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"
)

var (
	maxnonce = math.MaxInt32
)

//定义一个区块结构体
type Block struct {
	Version      int32
	PreBlockHash []byte
	Merkleroot   []byte
	Hash         []byte
	Time         int32
	Bits         int32
	Nonce        int32
	Transations  []*Transation
	Height       int32
}
//区块序列化
func (block *Block) serialize() []byte {
	result := bytes.Join([][]byte{IntToHex(block.Version),
		block.PreBlockHash,
		block.Merkleroot,
		block.Hash,
		IntToHex(block.Time),
		IntToHex(block.Bits),
		IntToHex(block.Nonce)},
		[]byte{})
	return result
}
//区块序列化方式二
func (b*Block) Serialize()[]byte {
	encoded:=new(bytes.Buffer)
	enc := gob.NewEncoder(encoded)
	err := enc.Encode(b)
	if err != nil{
		panic(err)
	}
	return encoded.Bytes()
}

//反序列化
func DeserializeBlock(d []byte) *Block{
	var block Block
	decode:=gob.NewDecoder(bytes.NewReader(d))
	err := decode.Decode(&block)
	if err!=nil{
		log.Panic(err)
	}
	return  &block
}

//计算比特币目标值
func CalculaterTargetFast(bits []byte) []byte {
	var result []byte
	//第一个字节 计算指数
	exponent := bits[:1]
	fmt.Printf("%x\n", exponent)
	//计算后面3个系数
	coeffient := bits[1:]
	fmt.Printf("%x\n", coeffient)

	str := hex.EncodeToString(exponent)
	exp, _ := strconv.ParseInt(str, 16, 8)
	fmt.Printf("%d\n", exp)
	//在前面拼接00
	result = append(bytes.Repeat([]byte{0x00}, 32-int(exp)), coeffient...)
	//在后面拼接00
	result = append(result, bytes.Repeat([]byte{0x00}, 32-len(result))...)
	return result
}
func (b * Block) String(){
	fmt.Printf("version: %s\n",strconv.FormatInt(int64(b.Version),10))
	fmt.Printf("PreBlockHash:%x\n", b.PreBlockHash)
	fmt.Printf("Merkleroot:%x\n", b.Merkleroot)
	fmt.Printf("Hash:%x\n", b.Hash)
	fmt.Printf("Time:%s\n", strconv.FormatInt(int64(b.Time),10))
	fmt.Printf("Bits:%s\n", strconv.FormatInt(int64(b.Bits),10))
	fmt.Printf("Nonce:%s\n", strconv.FormatInt(int64(b.Nonce),10))
	fmt.Printf("blockHeight:%s\n", strconv.FormatInt(int64(b.Height),10))
	fmt.Printf("_______________________________________\n")

}
func NewBlock(transactions []*Transation,prevBlockHash []byte,height int32) * Block{
	block := &Block{2,
		prevBlockHash,
		[]byte{},
		[]byte{},
		1418755780,
		404454260,
		0,
		transactions,
		height,
	}
	pow := NewProofofWork(block)
	nonce,hash:=   pow.Run()
	block.Nonce = nonce
	block.Hash = hash
	return  block
}

func (b *Block) createMerklTreeRoot(transations []*Transation) {
	var tranHash [][]byte
	for _, tx := range transations {
		tranHash = append(tranHash, tx.Hash())
		mTree := NewMerkleTree(tranHash)
		b.Merkleroot = mTree.RootNode.Data
	}
}


func NewGensisBlock(transactions []*Transation) * Block {
	block := &Block{2,
		[]byte{},
		[]byte{},
		[]byte{},
		int32(time.Now().Unix()),
		404454260,
		0,
		transactions,
		0,
	}
	pow := NewProofofWork(block)
	nonce,hash:=   pow.Run()
	block.Nonce = nonce
	block.Hash = hash
	return  block

}


