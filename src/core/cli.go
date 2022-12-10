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
	cleanCmd := flag.NewFlagSet("clean", flag.ExitOnError)
	createChainCmd := flag.NewFlagSet("createchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddrCmd := flag.NewFlagSet("listaddr", flag.ExitOnError)
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
	case "clean":
		_ = cleanCmd.Parse(os.Args[2:])
	case "createchain":
		_ = createChainCmd.Parse(os.Args[2:])
	case "printchain":
		_ = printChainCmd.Parse(os.Args[2:])
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaddr":
		err := listAddrCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "transfer":
		_ = transferCmd.Parse(os.Args[2:])
	case "balance":
		_ = balanceCmd.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
	}

	// 解析相关并执行命令
	if cleanCmd.Parsed() {
		cli.cleanEnv()
	}
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

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddrCmd.Parsed() {
		cli.listaddr()
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
	log.Println("	clean - clean env")
	log.Println("	createchain -address address - init block chain")
	log.Println("	printchain - print all blocks of the blockchain")
	log.Println("	createwallet - generates a new key-pair and saves it into the wallet file")
	log.Println("	listaddr - lists all addresses from the wallet file")
	log.Println("	transfer -form tom -to jerry -amount 1 - tom transfers 1 coin to jerry")
	log.Println("	balance -address address - print balance of address")
}

func (cli *CLI) cleanEnv() {
	os.Remove(dbFile)
	os.Remove(dbFile + ".lock")
	os.Remove(walletFile)
	fmt.Println("Clean Done!")
}

// 创建并获取链
func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	bc := NewBlockchain(address)
	defer bc.Db.Close()

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()
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

// 创建钱包
func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}

// 打印地址
func (cli *CLI) listaddr() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

// 转账
func (cli *CLI) transfer(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := GetBlockchain()
	UTXOSet := UTXOSet{bc}
	defer bc.Db.Close()

	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}

	wallet := wallets.GetWallet(from)
	tx := NewTransaction(&wallet, to, amount, &UTXOSet)

	cbTx := NewRewardTX(from, "")
	txs := []*Transaction{cbTx, tx}

	newBlock := bc.AddBlock(txs)
	UTXOSet.Update(newBlock)
	fmt.Printf("%s transfers %d coin to %s\n", from, amount, to)
}

// 获取余额
func (cli *CLI) getBalance(address string) {

	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := GetBlockchain()
	UTXOSet := UTXOSet{bc}
	defer bc.Db.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
