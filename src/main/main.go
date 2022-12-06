package main

import "go-blockchain/src/core"

// 参考实现： https://learnblockchain.cn/article/586
func main() {

	// 手动维护链条数据
	//bc := core.NewBlockchain()
	//bc.AddBlock("send 1 to jerry")
	//bc.AddBlock("send 2 to tom")

	//for _, block := range bc.Blocks {
	//	//fmt.Printf("PreHash: %x\nHash: %x\nData: %s\n", block.PreHash, block.Hash, block.Data)
	//
	//	isValid := core.NewProofOfWork(block).Validate()
	//	fmt.Println("validate result：", isValid)
	//	fmt.Println("")
	//}

	// 通过命令行方式添加区块
	bc := core.NewBlockchain("start")

	// defer用来声明一个延迟函数，把这个函数放入到一个栈上，当外部的包含方法return之前调用
	defer bc.Db.Close()

	// 创建cli并运行
	cli := core.NewCli(bc)
	cli.Run()
}
