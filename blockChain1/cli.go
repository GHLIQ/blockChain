package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type CLI struct {
	bc *Blockchain
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 1 {
		fmt.Println("参数小于1")
		os.Exit(1)
	}
	fmt.Println(os.Args)
}
func(cli *CLI) printUsage()  {
	fmt.Printf("Usage:\n")
	fmt.Printf("creatwallet-创建钱包\n")
	fmt.Printf("listaddress-打印地址\n")
	fmt.Printf("send-转账\n")
	fmt.Printf("getbalance-转账\n")
	fmt.Printf("addblock-增加区块\n")
	fmt.Printf("printChain-打印区块链\n")

}
func (cli *CLI) printChain(){
	cli.bc.printBlockchain()
}

func (cli *CLI) addBlock() {
	cli.bc.MineBlock([]*Transation{})
}
func (cli *CLI) getBalance(address string){
	balance := 0
	decodeAddress,_ := Base58Decode([]byte(address))
	pubkeyHash := decodeAddress[1:len(decodeAddress)-4]

	set := UTXOSet{cli.bc}
	UTXOs:=set.FindUTXObyPublicHash(pubkeyHash)
	//UTXOs := cli.bc.FindUTXO(pubkeyHash)
	for _,out := range UTXOs{
		balance += out.Value
	}
	fmt.Printf("\nbalance of'%s':%d\n",address,balance)
}

func(cli *CLI) send(from,to string,amount int)  {
	tx:= NewUTXOTransaction(from,to,amount,cli.bc)
	newblock:= cli.bc.MineBlock([]*Transation{tx})
	set:=UTXOSet{cli.bc}
	set.update(newblock)
	//cli.getBalance("JU2DBLjYCPq4Fsk9JeQZXARrnTqXj5zHa")
	//cli.getBalance("2UMvqNgxrn1KrFWyhM3twkoKx9EJpnCDc")
	//cli.getBalance("9pMu8Gx6JKMNdkEfBTpTzyhP7GiaLtTCx")
	fmt.Printf("Success")
}

func (cli *CLI) creatWallet() {
	wallets,_ := NewWallets()
	address:= wallets.CreateWallet()
	wallets.SaveToFile()
	fmt.Printf("your address :%s\n",address)
}
func (cli *CLI) listAddress() {
	wallets,err := NewWallets()
	if err!= nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddress()
	for _,address:= range addresses{
		fmt.Printf("%s\n",address)
	}

}
func (cli *CLI) getbestHeght() {
  fmt.Println(cli.bc.GetBaseHeight())
}
func (cli *CLI) Run() {
	cli.validateArgs()
    //所以需要先在命令行执行 export NODE_ID = 3000
	nodeID:= os.Getenv("NODE_I")
	if nodeID == ""{
		fmt.Printf("NODE_ID is not set")
		os.Exit(1)
	}

	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	startNodeminner := startNodeCmd.String("minner","","startNodeminner address")

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)

	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	getbalanceAddress:= getBalanceCmd.String("address","","the address to get balance of")

	sendCmd := flag.NewFlagSet("send",flag.ExitOnError)
	sendFrom:=sendCmd.String("from","","source wallet address")
	sendTO:=sendCmd.String("to","","DDestination wallet address")
	sendAmount:=sendCmd.Int("amount",0,"Amount to send")

	createWalletCmd := flag.NewFlagSet("creatwallet", flag.ExitOnError)
	listAddressCmd := flag.NewFlagSet("listaddress", flag.ExitOnError)
	getBastHeightCmd := flag.NewFlagSet("getBestHeight", flag.ExitOnError)
	switch os.Args[1] {
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getBestHeight":
		err := getBastHeightCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "creatwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddress":
		err := listAddressCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printChain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		cli.addBlock()
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
	if getBalanceCmd.Parsed() {
		if *getbalanceAddress ==""{
			os.Exit(1)
		}
		cli.getBalance(*getbalanceAddress)
	}
	if sendCmd.Parsed() {
		if *sendFrom ==""||*sendTO==""||*sendAmount <=0{
			os.Exit(1)
		}
		cli.send(*sendFrom,*sendTO,*sendAmount)
	}

	if createWalletCmd.Parsed() {
		cli.creatWallet()
	}
	if listAddressCmd.Parsed() {
		cli.listAddress()
	}
	if getBastHeightCmd.Parsed() {
		cli.getbestHeght()
	}

	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_I")
		if nodeID == ""{
			startNodeCmd.Usage()
			os.Exit(1)
		}
		cli.startNode(nodeID,*startNodeminner)
	}
}

func (cli *CLI) startNode(nodeID string,minnerAddress string) {
    fmt.Printf("Starting node%s\n",nodeID)

    if len(minnerAddress) >0{
    	if ValidateAddress([]byte(minnerAddress)){
    		fmt.Println("%minner is no",minnerAddress)
		}else{
			log.Panic("error minner Address")}
	}

	StartServer(nodeID,minnerAddress,cli.bc)
}



