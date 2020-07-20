package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

const dbFile = "blockchain.db"
const blockBucket = "blocks"
const genesisDate = "liqiang BlockChain"

type Blockchain struct {
	tip []byte //最近一个区块的哈希值
	db  *bolt.DB
}
type BlockChainIterateor struct {
	currenthash []byte
	db          *bolt.DB
}

func (bc *Blockchain) AddBlock(block *Block) {
	err :=bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		blockIndb := b.Get(block.Hash)
		if blockIndb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}
		lastHash := b.Get([]byte("l"))
		lastBlockdata := b.Get(lastHash)
		lastblok := DeserializeBlock(lastBlockdata)
		if block.Height > lastblok.Height {
		}
		err = b.Put([]byte("l"), block.Hash)
		{
			if err != nil {

				log.Panic(err)
			}
			bc.tip = block.Hash
		}
return  nil
	})
	if err!= nil{
		log.Panic(err)
	}
}

func (bc *Blockchain) MineBlock(transaction []*Transation) *Block {
	for _, tx := range transaction {
		if bc.VerifyTransation(tx) != true {
			log.Panic("ERROR:无效的交易")
		} else {
			fmt.Println("有效的交易")
		}
	}

	var lasthash []byte
	var lastheight int32
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		lasthash = b.Get([]byte("l"))
		blockdata := b.Get(lasthash)
		block := DeserializeBlock(blockdata)
		lastheight = block.Height
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	newBlock := NewBlock(transaction, lasthash, lastheight+1)
	bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		err = b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = newBlock.Hash
		return nil
	})
	return newBlock
}

func NewBlocchain(address string) *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		if b == nil {
			fmt.Printf("区块链不存在，创建一个新的区块链\n")
			transactions := NewCoinbaseTX(address, genesisDate)
			genesis := NewGensisBlock([]*Transation{transactions})
			b, err := tx.CreateBucket([]byte(blockBucket))
			if err != nil {
				log.Panic(err)
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash

		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err != nil {
		log.Panic()
	}
	bc := Blockchain{tip, db}
	set := UTXOSet{&bc}
	set.Reindex()
	return &bc
}
func (bc *Blockchain) iterator() *BlockChainIterateor {
	bci := &BlockChainIterateor{bc.tip, bc.db}
	return bci
}
func (i *BlockChainIterateor) Next() *Block {
	var block *Block
	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		deblock := b.Get(i.currenthash)
		block = DeserializeBlock(deblock)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	i.currenthash = block.PreBlockHash
	return block
}
func (bc *Blockchain) printBlockchain() {
	bci := bc.iterator()
	for {
		block := bci.Next()
		block.String()

		if len(block.PreBlockHash) == 0 {
			break
		}
	}

}

func (bc *Blockchain) FindUnspentTransaction(pubkeyhash []byte) []Transation {
	var unspentTXs []Transation         //所有未花费的交易
	spendTXOs := make(map[string][]int) //存储已经花费的交易  string代表交易的哈希值 ->【】int 代表输出的序号
	bci := bc.iterator()
	//遍历区块链，获取每一个区块
	for {
		block := bci.Next()
		//遍历区块上的所有交易
		for _, tx := range block.Transations {
			//txid交易的哈希值
			txID := hex.EncodeToString(tx.ID)
		output:
			//遍历一笔交易的输出 outIdx代表输出序号
			for outIdx, out := range tx.Vout {
				if spendTXOs[txID] != nil {
					for _, spentOut := range spendTXOs[txID] {
						if spentOut == outIdx {
							continue output
						}
					}
				}
				if out.CanBeUnlockedWith(pubkeyhash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					if in.BanBeUnlockedWith(pubkeyhash) {
						inTxId := hex.EncodeToString(in.TXid)
						spendTXOs[inTxId] = append(spendTXOs[inTxId], in.Voutindex)
					}
				}
			}
		}

		if len(block.PreBlockHash) == 0 {
			break
		}
	}
	fmt.Println(unspentTXs)
	return unspentTXs

}

func (bc *Blockchain) FindUTXO(pubkeyhash []byte) []TXOutput {
	var UTXOs []TXOutput
	unspendTransations := bc.FindUnspentTransaction(pubkeyhash)
	for _, tx := range unspendTransations {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(pubkeyhash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(pubkeyhash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransaction(pubkeyhash)

	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(pubkeyhash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
}

func (bc *Blockchain) SignTransation(tx *Transation, prikey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transation)
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransationByID(vin.TXid)
		if err != nil {
			log.Panic()
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(prikey, prevTXs)
}

func (bc *Blockchain) FindTransationByID(ID []byte) (Transation, error) {
	bci := bc.iterator()
	for {
		block := bci.Next()
		for _, tx := range block.Transations {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PreBlockHash) == 0 {
			break
		}
	}
	return Transation{}, errors.New("transation is not found")
}

func (bc *Blockchain) VerifyTransation(tx *Transation) bool {
	prevTXs := make(map[string]Transation)
	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransationByID(vin.TXid)
		if err != nil {
			log.Panic(err)
		}

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return tx.Verify(prevTXs)
}

///这个函数的作用就是获取整个区块链所有未花费的交易  用map[string]TXOutputs来存储  string代表交易的哈希值，TXOutputs代表这个哈希值的输出（可能会存在多个）
///函数逻辑：
///①定义连个map，分别用来存储未花费的交易输出（make(map[string]TXOutputs)）和已经花费的交易输出（make(map[string][]int) 其中string代表交易的哈希值，【】int切片存储的是输入的索引号）
///②然后从后往前遍历每一个区块，每一个区块中遍历每一笔交易，每一笔交易遍历所有输出，最后判断输出是否在已经花费的交易里面（先判断交易哈希是否在里面，如果在里面的话，再判断输出的索引号是否在里面）
func (bc *Blockchain) FindALLUTXO() map[string]TXOutputs {
	//未花费的交易
	UTXO := make(map[string]TXOutputs)
	//已经花费的交易
	spentTXs := make(map[string][]int)
	bci := bc.iterator()
	for {
		block := bci.Next()

		for _, tx := range block.Transations {
			//txID 代表每一笔交易的哈希值
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXs[txID] != nil {
					for _, spendOutIds := range spentTXs[txID] {
						if spendOutIds == outIdx {
							continue Outputs
						}
					}

				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinBase() == false {
				for _, in := range tx.Vin {
					inTXID := hex.EncodeToString(in.TXid)
					spentTXs[inTXID] = append(spentTXs[inTXID], in.Voutindex)
				}
			}

		}
		if len(block.PreBlockHash) == 0 {
			break
		}
	}
	return UTXO
}

func (bc *Blockchain) GetBaseHeight() int32 {
	var lastBlock Block
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		lashHash := b.Get([]byte("l"))
		blockdate := b.Get(lashHash)
		lastBlock = *DeserializeBlock(blockdate)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return lastBlock.Height
}

func (bc *Blockchain) Getblockhash() [][]byte {
	var blocks [][]byte
	bci := bc.iterator()
	for {
		block := bci.Next()
		blocks = append(blocks, block.Hash)

		if len(block.PreBlockHash) == 0 {
			break
		}
	}
	return blocks
}

func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockBucket))
		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not Fund")
		}

		block = *DeserializeBlock(blockData)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return block, nil
}
