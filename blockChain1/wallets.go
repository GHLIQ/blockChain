package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const WalletFile string = "wallet.dat"

type Wallets struct {
	Walletsstore map[string] *Wallet //string 比特币地址  映射到钱包
}

func NewWallets()(*Wallets,error){
	wallets:= Wallets{}
	wallets.Walletsstore = make(map[string] *Wallet)
    err :=	wallets.LoadFromFile()
    return &wallets,err
}


//创建钱包
func (ws *Wallets) CreateWallet() string{
	wallet:= NewWallet()
	//把字节数组变字符串
	address:= fmt.Sprintf("%s",wallet.GetAddress())
	ws.Walletsstore[address] = wallet
	return address
}
//获取当前传入比特币地址的钱包
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Walletsstore[address]
}

func (ws *Wallets) GetAddress() []string {
	var addresses []string
	for address,_:= range ws.Walletsstore{
		addresses = append(addresses,address)
	}
	return  addresses
}
//读取文件操作
func (ws *Wallets) LoadFromFile() error  {
	//判断文件是否存在
	if _,err := os.Stat(WalletFile);os.IsNotExist(err){
		return err
	}
	//读取文件
	fileContent,err := ioutil.ReadFile(WalletFile)
	if err!= nil{
		log.Panic(err)
	}
	//反序列化 存入wallets
	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil{
		log.Panic(err)
	}
    ws.Walletsstore = wallets.Walletsstore
    return  nil
}
//写入文件操作
func (ws *Wallets) SaveToFile()  {
	var content bytes.Buffer
	gob.Register(elliptic.P256())
	encoder:= gob.NewEncoder(&content)
	err:= encoder.Encode(ws)
	if err!=nil{
		log.Panic(err)
	}

	err = ioutil.WriteFile(WalletFile,content.Bytes(),0777)//0777 代表最高权限
	if err!= nil {
		log.Panic(err)
	}
}