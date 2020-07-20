package main

func main() {
	//________________________________________________
	//target:= big.NewInt(1)
	//target.Lsh(target,uint(256-targetBits))
	//fmt.Printf("%x",target.Bytes())
	//—————————————————————————————————————————————————
	//TestPow()
	//_________________________________________________
	//NewGensisBlock()
	//------------------------------------------------
	//TestBoltDB()
	//------------------------------------------------
	bc:= NewBlocchain("2UMvqNgxrn1KrFWyhM3twkoKx9EJpnCDc")
	cli := CLI{bc}
	cli.Run()
	//-------------------------------------------------
	//wallet:= NewWallet()
	//fmt.Printf("\n私钥：%x\n",wallet.PrivateKey.D.Bytes())
	//fmt.Printf("公钥：%x\n",wallet.PublicKey)
	//fmt.Printf("地址：%x\n",wallet.GetAddress())
	//address,_ := hex.DecodeString("41755634694741776e37334762634c393665624b6f43437652324d347553377141")
	//fmt.Printf("%v\n",ValidateAddress(address))
	//-----------------------------------------------------
	//bc:= NewBlocchain("2UMvqNgxrn1KrFWyhM3twkoKx9EJpnCDc")
	//tx:= NewUTXOTransaction("2UMvqNgxrn1KrFWyhM3twkoKx9EJpnCDc","9pMu8Gx6JKMNdkEfBTpTzyhP7GiaLtTCx",20,bc)
	//bc.MineBlock([]*Transation{tx})
	//fmt.Printf("Success")

}