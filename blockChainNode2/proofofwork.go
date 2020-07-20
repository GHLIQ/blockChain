package main

import (
	"bytes"
	"crypto/sha256"
	"math/big"
)

type ProofofWork struct {
	block   *Block
	tartget *big.Int
}

const targetBits = 24

func NewProofofWork(b *Block) *ProofofWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofofWork{b, target}
	return pow
}

func (pow *ProofofWork) prepareData(nonce int32) []byte {
	data := bytes.Join([][]byte{IntToHex(pow.block.Version),
		pow.block.PreBlockHash,
		pow.block.Merkleroot,
		pow.block.Hash,
		IntToHex(pow.block.Time),
		IntToHex(pow.block.Bits),
		IntToHex(nonce)},
		[]byte{})
	return data
}

func (pow *ProofofWork) Run() (int32, []byte) {

	var nonce int32
	var secondHash [32]byte
	nonce = 0
	var currentHash big.Int
	for nonce < int32(maxnonce) {
		//序列化
		data := pow.prepareData(nonce)
		//double hash
		firshHash := sha256.Sum256(data)
		secondHash = sha256.Sum256(firshHash[:])
		currentHash.SetBytes(secondHash[:])
		//fmt.Printf("%x\n",currentHash)
		if currentHash.Cmp(pow.tartget) == -1 {
			break
		} else {
			nonce++
		}
	}
	return  nonce,secondHash[:]
}

func (pow * ProofofWork) validata() bool{
	var  hashInt big.Int
	data:= pow.prepareData(pow.block.Nonce)
	firshHash := sha256.Sum256(data)
	secondHash := sha256.Sum256(firshHash[:])
	hashInt.SetBytes(secondHash[:])
	isValid := hashInt.Cmp(pow.tartget) ==-1
	return  isValid
}