package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const nodeversion = 0x00
const commnoLenth = 12

var nodeAddress string
var blockInTransit [][]byte

type Version struct {
	Version    int
	BestHeight int32
	AddFrom    string
}

func (ver *Version) String() {
	fmt.Printf("Version:%d\n", ver.Version)
	fmt.Printf("BestHeight:%d\n", ver.BestHeight)
	fmt.Printf("AddFrom:%s\n", ver.AddFrom)
}

//存储的是节点已经探测到的网络
var knowNodes = []string{"localhost:3000"}

func StartServer(nodeID, minerAddress string, bc *Blockchain) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	ln, _ := net.Listen("tcp", nodeAddress)
	defer ln.Close()
	//bc := NewBlocchain("2UMvqNgxrn1KrFWyhM3twkoKx9EJpnCDc")
	if nodeAddress != knowNodes[0] {
		sendVersion(knowNodes[0], bc)
	}
	for {
		conn, err2 := ln.Accept()
		if err2 != nil {
			log.Panic(err2)
		}
		go handleConnection(conn, bc)
	}
}

func handleConnection(conn net.Conn, bc *Blockchain) {
	request, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	//获取命令
	command := BytesToCommand(request[:commnoLenth])
	switch command {
	case "version":
		handleVersion(request, bc)
	case "getblocks":
		handleGetBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "block":
		handleBlock(request,bc)
	}


}

func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload blocksend
	buff.Write(request[commnoLenth:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err!=nil{
		log.Panic(err)
	}
	blockdata:=payload.Block

	block:=DeserializeBlock(blockdata)
bc.AddBlock(block)
	fmt.Printf("Recieve a new Block")

	if len(blockInTransit)>0{
		blockHash:= blockInTransit[0]
		sendGetData(payload.AddFrom,"block",blockHash)
		blockInTransit = blockInTransit[1:]
	}else{
	set :=UTXOSet{bc}
	set.Reindex()
	}
}

func handleGetData(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getdate
	buff.Write(request[commnoLenth:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			log.Panic(err)
		}
		sendBlock(payload.AddrFrom, &block)
	}

}

type blocksend struct {
	AddFrom string
	Block   []byte
}

func sendBlock(addr string, block *Block) {
	data := blocksend{nodeAddress, block.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)
	sendData(addr, request)
}

func handleInv(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload inv
	buff.Write(request[commnoLenth:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("Recieve inventory %d %s", len(payload.Items), payload.Type)
	//for _,b := range payload.Items{
	//	fmt.Printf("%x\n",b)
	//}
	if payload.Type == "block" {
		blockInTransit = payload.Items
		blockHash := payload.Items[0]
		sendGetData(payload.AddFrom, "block", blockHash)
		newInTransit := [][]byte{}
		for _, b := range blockInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}

		blockInTransit = newInTransit
	}
}

type getdate struct {
	AddrFrom string
	Type     string
	ID       []byte
}

func sendGetData(addr string, kind string, id []byte) {
	payload := gobEncode(getdate{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)
	sendData(addr, request)
}

func handleGetBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload getblocks
	buff.Write(request[commnoLenth:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	block := bc.Getblockhash()
	sendInv(payload.Addfrom, "block", block)
}

type inv struct {
	AddFrom string
	Type    string
	Items   [][]byte
}

func sendInv(addr string, kind string, items [][]byte) {
	inventory := inv{nodeAddress, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(addr, request)
}

func handleVersion(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload Version
	buff.Write(request[commnoLenth:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	payload.String()
	myBestHeight := bc.GetBaseHeight()
	foreignerBestHeight := payload.BestHeight
	if myBestHeight > foreignerBestHeight {
		sendGetBlock(payload.AddFrom)
	} else {
		sendVersion(payload.AddFrom, bc)
	}
	if !nodeIsKnow(payload.AddFrom) {
		knowNodes = append(knowNodes, payload.AddFrom)
	}
}

type getblocks struct {
	Addfrom string
}

func sendGetBlock(address string) {
	payload := gobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)
	sendData(address, request)
}

func nodeIsKnow(addr string) bool {
	for _, node := range knowNodes {
		if node == addr {
			return true
		}
	}
	return false
}

func sendVersion(addr string, bc *Blockchain) {
	bestHeight := bc.GetBaseHeight()
	payload := gobEncode(Version{nodeversion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)

}

func sendData(addr string, data []byte) {
	//先跟地址网络建立连接
	con, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("%s is not available", addr)

		var updateNodes []string

		for _, node := range knowNodes {
			if node != addr {
				updateNodes = append(updateNodes, node)
			}
		}
		knowNodes = updateNodes
	}
	defer con.Close()

	_, err = io.Copy(con, bytes.NewReader(data))

	if err != nil {
		log.Panic(err)
	}
}

func commandToBytes(command string) []byte {
	var bytes [commnoLenth]byte
	for i, c := range command {
		bytes[i] = byte(c)

	}
	return bytes[:]
}
func BytesToCommand(bytes []byte) string {
	var command []byte
	for _, b := range bytes {
		if b != 0x00 {
			command = append(command, b)
		}
	}
	//把动态字节数组转换为string
	return fmt.Sprintf("%s", command)
}

//序列化
func gobEncode(data interface{}) []byte {
	var buff = new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic()
	}
	return buff.Bytes()
}
