package main


import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/ripemd160"
	"math/big"
)

var base58Alphabets = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")
func main(){
	publickey,_ := hex.DecodeString("7559DA5FCAD1EA5C54A20130406CCE9D533FDD6F08CF74CAE5CB97E4F2138D815493B62B09154F7869B3DD0191C15B09148320779959709F3AC48266441F36D3")
	address := generateAddress(publickey)
	fmt.Printf("\n%s\n",address)


	resByte, resStr := Base58Decode([]byte("E1PhExf3HutwhWUbaWaERYRB192RcZJyw"))
	fmt.Println("resByte=", resByte)
	fmt.Println("resByte=", resByte[10:])
	fmt.Println("resStr=", resStr)

	// bewr := []byte{0x00}
	//fmt.Println("resBytsdfse=", bewr)

}
//生成比特币地址
func generateAddress(pubkey []byte) []byte{
	pubkeyHash256 := sha256.Sum256(pubkey)
	PIPEMD160Hasher := ripemd160.New()
	_,err:= PIPEMD160Hasher.Write(pubkeyHash256[:])
	if err!=nil {
		fmt.Println("error")
	}
	publicRIPEMD160 := PIPEMD160Hasher.Sum(nil)
	fmt.Printf("%d\n",publicRIPEMD160 )
	versionPayload := append([]byte{0x00},publicRIPEMD160...)
	firstSHA := sha256.Sum256(versionPayload)
	secondSHA:=sha256.Sum256(firstSHA[:])
	checksum := secondSHA[:4]
	fullPayload:= append(versionPayload,checksum...)
	address,_:= Base58Encode(fullPayload)
	return  address
}
//生成base58编码
func Base58Encode(input []byte) ([]byte, string) {
	x := big.NewInt(0).SetBytes(input)
	//fmt.Println("x=", x)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := &big.Int{}
	var result []byte
	// 被除数/除数=商……余数
	//fmt.Println("开始循环-------")
	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		//fmt.Println("mod=", mod)
		//fmt.Println("x=", x)
		result = append(result, base58Alphabets[mod.Int64()])
		//fmt.Println("一次循环结束-------")
	}
	ReverseBytes(result)
	return result, string(result)
}

// ReverseBytes 翻转字节
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

}



// Base58Decode 解码
func Base58Decode(input []byte) ([]byte, string) {
	result := big.NewInt(0)
	for _, b := range input {
		charIndex := bytes.IndexByte(base58Alphabets, b)
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}
	decoded := result.Bytes()
	//if input[0] == base58Alphabets[0] {
		decoded = append([]byte{0x00}, decoded...)
	//}
	return decoded, string(decoded)
}