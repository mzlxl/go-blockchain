package main

import "go-blockchain/src/core"

// 参考实现：https://learnblockchain.cn/article/577
func main() {
	// 通过命令方式运行（首先创建区块链，给tom奖励10个币）
	// go run .\src\main\main.go createchain -address tom
	// go run .\src\main\main.go printchain
	// go run .\src\main\main.go transfer -from tom -to jerry -amount 3
	// go run .\src\main\main.go balance -address tom
	// go run .\src\main\main.go balance -address jerry
	// go run .\src\main\main.go printchain
	cli := core.CLI{}
	cli.Run()
}
