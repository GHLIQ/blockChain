package main

import (
	"fmt"
	"log"
)

func TestCreateMerkleTreeRoot() {
	txin := TXInput{[]byte{}, -1, nil,[]byte{}}
	txout := NewTXOutput(subsidy, "first")
	tx := Transation{nil, []TXInput{txin}, []TXOutput{*txout}}

	txin2 := TXInput{[]byte{}, -1, nil,[]byte{}}
	txout2 := NewTXOutput(subsidy, "second")
	tx2 := Transation{nil, []TXInput{txin2}, []TXOutput{*txout2}}

	var Transations []*Transation
	Transations = append(Transations, &tx, &tx2)

	block := &Block{2,
		[]byte{},
		[]byte{},
		[]byte{},
		1418755780,
		404454260,
		0,
		[]*Transation{},
		0,
	}
	block.createMerklTreeRoot(Transations)
	fmt.Printf("%x\n",block.Merkleroot)
}
func  TestNewSerialize(){
	block := &Block{2,
		[]byte{},
		[]byte{},
		[]byte{},
		1418755780,
		404454260,
		0,
		[]*Transation{},
		0,
	}
	deBlock := DeserializeBlock(block.Serialize())
	deBlock.String()
}

func TestPow(){
	block := &Block{2,
		[]byte{},
		[]byte{},
		[]byte{},
		1418755780,
		404454260,
		0,
		[]*Transation{},
		0,
	}

	pow := NewProofofWork(block)
	nonce,_:=pow.Run()
	block.Nonce = nonce
	fmt.Println(pow.validata())
}

func TestBoltDB()  {
	blockchain:= NewBlocchain("2UMvqNgxrn1KrFWyhM3twkoKx9EJpnCDc")
	blockchain.MineBlock([]*Transation{})
	blockchain.MineBlock([]*Transation{})
	blockchain.printBlockchain()

}

func Testaa() {
	wallets ,err := NewWallets()
	if err!=nil{
		log.Panic()
	}
	wallet:= wallets.GetWallet("2UMvqNgxrn1KrFWyhM3twkoKx9EJpnCDc")
	aa:= HashPubkey(wallet.PublicKey)
	fmt.Printf("%d\n",aa )       //[16 36 211 163 87 14 39 90 200 8 179 164 22 27 7 211 154 57 231 171]


}