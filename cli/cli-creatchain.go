package cli

import (
	"fmt"
	"github.com/azd1997/golang-MimbleWimble-try/blockchain"
	"github.com/azd1997/golang-MimbleWimble-try/utils"
	"github.com/azd1997/golang-MimbleWimble-try/wallet"
	"log"
)

/*创建区块链，其创世区块coinbase交易地址给定*/
func (cli *CommandLine) createBlockChain(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not valid")
	}

	chain := blockchain.InitBlockChain(address, nodeID)
	//defer chain.Db.Close()

	//UTXO
	UTXOSet := blockchain.UTXOSet{chain}
	UTXOSet.Reindex()

	err := chain.Db.Close()
	utils.Handle(err)

	fmt.Println("Finished!")
}


//createBlockChain <-
//1.验证地址合法性（增加：若地址被使用过应从钱包文件删除）
//2.以address, nodeid为参数初始化区块链
//3.根据区块链对象新建UTXO集
//4.对UTXO集重索引
//5.关闭数据库