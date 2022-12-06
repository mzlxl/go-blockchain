package core

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// 提供一个与程序交互的接口，通过命令行方式控制NewBlockchain 和 bc.AddBlock
type CLI struct {
	bc *Blockchain
}

func NewCli(bc *Blockchain) *CLI {
	return &CLI{bc}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	// 使用标准库里面的 flag 包来解析命令行参数
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	// 给 addblock命令 添加 -data 标志
	addBlockData := addBlockCmd.String("data", "", "Block data")

	// 命令解析
	switch os.Args[1] {
	case "addblock":
		_ = addBlockCmd.Parse(os.Args[2:])
	case "printchain":
		_ = printChainCmd.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
	}

	// 解析相关并执行命令
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			// 没有-data参数时，直接退出
			addBlockCmd.Usage()
			os.Exit(1)
		}
		transactions := []*Transaction{
			&Transaction{nil, []TXInput{}, []TXOutput{}},
		}
		cli.addBlock(transactions)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
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
	log.Println("	addblock -data data - add block to blockchain")
	log.Println("	printchain - print all blocks of the blockchain")
}

// 添加区块
func (cli *CLI) addBlock(transactions []*Transaction) {
	cli.bc.AddBlock(transactions)
}

// 通过迭代方式打印链条
func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev. hash: %x\n", block.PreHash)
		fmt.Printf("Transactions: %s\n", block.Transactions)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PreHash) == 0 {
			break
		}
	}
}
