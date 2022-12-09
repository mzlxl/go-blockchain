package core

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct {
}

func (cli *CLI) Run() {
	cli.validateArgs()

	// 使用标准库里面的 flag 包来解析命令行参数
	createChainCmd := flag.NewFlagSet("createchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	transferCmd := flag.NewFlagSet("transfer", flag.ExitOnError)
	balanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)

	// 给 createchain命令 添加 -address 标志
	createChainAddress := createChainCmd.String("address", "", "The address to send genesis block reward to")
	balanceAddress := balanceCmd.String("address", "", "The address to get balance for")
	transferFromAddress := transferCmd.String("from", "", "Source wallet address")
	transferToAddress := transferCmd.String("to", "", "Destination wallet address")
	transferAmount := transferCmd.Int("amount", 0, "Amount to send")

	// 命令解析
	switch os.Args[1] {
	case "createchain":
		_ = createChainCmd.Parse(os.Args[2:])
	case "printchain":
		_ = printChainCmd.Parse(os.Args[2:])
	case "transfer":
		_ = transferCmd.Parse(os.Args[2:])
	case "balance":
		_ = balanceCmd.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
	}

	// 解析相关并执行命令
	if createChainCmd.Parsed() {
		if *createChainAddress == "" {
			createChainCmd.Usage()
			os.Exit(1) // 没有-address参数时，直接退出
		}
		cli.createBlockchain(*createChainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if transferCmd.Parsed() {
		if *transferFromAddress == "" || *transferToAddress == "" || *transferAmount <= 0 {
			transferCmd.Usage()
			os.Exit(1)
		}
		cli.transfer(*transferFromAddress, *transferToAddress, *transferAmount)
	}

	if balanceCmd.Parsed() {
		if *balanceAddress == "" {
			balanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*balanceAddress)
	}
}

// 校验参数
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// 使用说明
func (cli *CLI) printUsage() {
	log.Println("Usage:")
	log.Println("	createchain -address address - init block chain")
	log.Println("	printchain - print all blocks of the blockchain")
	log.Println("	transfer -form tom -to jerry -amount 1 - tom transfers 1 coin to jerry")
	log.Println("	balance -address address - print balance of address")
}

// 创建并获取链
func (cli *CLI) createBlockchain(address string) {
	// 校验钱包
	//if !ValidateAddress(address) {
	//	log.Panic("ERROR: Address is not valid")
	//}
	bc := NewBlockchain(address)
	defer bc.Db.Close()
	fmt.Println("createBlockchain Done!")
}

// 通过迭代方式打印链条
func (cli *CLI) printChain() {
	bc := GetBlockchain()
	defer bc.Db.Close()
	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PreHash)
		fmt.Printf("Transactions: %x\n", block.HashTransactions())
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PreHash) == 0 {
			break
		}
	}
}

// 转账
func (cli *CLI) transfer(from, to string, amount int) {
	bc := GetBlockchain()
	defer bc.Db.Close()

	tx := NewTransaction(from, to, amount, bc)
	bc.AddBlock([]*Transaction{tx})
	fmt.Printf("%s transfers %d coin to %s\n", from, amount, to)
}

// 获取余额
func (cli *CLI) getBalance(address string) int {
	bc := GetBlockchain()
	defer bc.Db.Close()
	unspentTransactions := bc.FindUnspentTransactions(address)
	accumulated := 0

	// 迭代未护花费的，将属于我的余额（CanBeUnlockedWith）进行余额累加，
	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				accumulated += out.Value
			}
		}
	}
	fmt.Printf("balance of %s:%d\n", address, accumulated)
	return accumulated
}
