package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/ripemd160"
)

const  version =  byte(0x00)
type Wallet  struct{

	PrivateKey ecdsa.PrivateKey
    PublicKey []byte

}

func NewWallet()  *Wallet {
   private,public := newKeyPair()
   wallet := Wallet{private,public}
   return  &wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	//生产椭圆曲线，go内置secp256r1曲线  比特币中的曲线是secp256k1
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)

	if err != nil {
		fmt.Println("error")
	}
	pubkey := append(private.PublicKey.X.Bytes(), private.Y.Bytes()...)
	return *private, pubkey
}

func (w Wallet) GetAddress() []byte{
	pubkeyHash := HashPubkey(w.PublicKey)
	versionPayload := append([]byte{version},pubkeyHash...)
	check:=checksum(versionPayload)
	fullPayload:= append(versionPayload,check...)
	address,_:= Base58Encode(fullPayload)
	return  address
}
func HashPubkey(pubkey []byte) []byte {
	pubkeyHash256:= sha256.Sum256(pubkey)
	PIPEMD160Hasher := ripemd160.New()
	_,err:= PIPEMD160Hasher.Write(pubkeyHash256[:])
	if err!=nil {
		fmt.Println("error")
	}
	publicRIPEMD160 := PIPEMD160Hasher.Sum(nil)
	return  publicRIPEMD160
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA:=sha256.Sum256(firstSHA[:])
	//checksum是前面的4个字节
	checksum := secondSHA[:4]
	return checksum
}

func ValidateAddress(address []byte) bool{
	pubkeyhash,_ := Base58Decode(address)
	actualCheckSum := pubkeyhash[len(pubkeyhash)-4:]
	publickHash := pubkeyhash[1:len(pubkeyhash)-4]
	targetChecksum:=checksum(append([]byte{0x00},publickHash...))
	return bytes.Compare(actualCheckSum,targetChecksum)==0
}