package main

import "go-blockchain/src/core"

// 参考实现：https://learnblockchain.cn/article/577
func main() {
	// 通过命令方式运行（首先创建区块链，给tom奖励10个币）
	// go run .\src\main\main.go clean
	// go run .\src\main\main.go createwallet
	// go run .\src\main\main.go listaddr
	// go run .\src\main\main.go createchain -address 1PfpqyEvx7R1551YE75pzc8jCajfnPkLK1
	// go run .\src\main\main.go balance -address  1PfpqyEvx7R1551YE75pzc8jCajfnPkLK1
	// go run .\src\main\main.go createwallet
	// go run .\src\main\main.go transfer -from 1PfpqyEvx7R1551YE75pzc8jCajfnPkLK1 -to 1GG73iS7GhRxPyksADckHLZtQdNuUpAN8o -amount 3
	// go run .\src\main\main.go balance -address 1GG73iS7GhRxPyksADckHLZtQdNuUpAN8o
	// go run .\src\main\main.go printchain
	cli := core.CLI{}
	cli.Run()
}
