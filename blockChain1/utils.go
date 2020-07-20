package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)
//计算两个数的最小值
func min(a int, b int) int {
	if a > b {
		return b
	}
	return a
}
//将int类型转换为【】byte ，小端在后
func IntToHex(num int32) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.LittleEndian, num)

	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
//将int类型转换为【】byte ，大端在后
func IntToHex2(num int32) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)

	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// ReverseBytes 翻转字节
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
func main1() {
	//前一个区块哈希值
	prev, _ := hex.DecodeString("000000000000000016145aa12fa7e81a304c38aec3d7c5208f1d33b587f966a6")
	ReverseBytes(prev)
	//摩卡根哈希值
	merkleroot, _ := hex.DecodeString("3a4f410269fcc4c7885770bc8841ce6781f15dd304ae5d2770fc93a21dbd70d7")
	ReverseBytes(merkleroot)

	block := &Block{2,
		prev,
		merkleroot,
		[]byte{},
		1418755780,
		404454260,
		0,
		[]*Transation{},
		0,
	}
	targetHash:= CalculaterTargetFast(IntToHex2(block.Bits))
	var target big.Int
	target.SetBytes(targetHash)
	block.Nonce = 1865996500
	var currentHash big.Int
	for block.Nonce < int32(maxnonce) {
		data:= block.serialize()
		firshHash:= sha256.Sum256(data)
		secondHash := sha256.Sum256(firshHash[:])
		ReverseBytes(secondHash[:])
		currentHash.SetBytes(secondHash[:])
		fmt.Printf("nonce:%d,  currenthash:%x\n",block.Nonce,secondHash)
		if currentHash.Cmp(&target) == -1{
			fmt.Printf("挖矿成功\n")
			break
		}else{
			block.Nonce++
		}
	}

}