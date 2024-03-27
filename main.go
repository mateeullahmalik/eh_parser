package main

import (
	"context"
	"fmt"

	"github.com/mateeullahmalik/eh_parser/ethereum"
	"github.com/mateeullahmalik/eh_parser/parser"
	infraEth "github.com/mateeullahmalik/eh_parser/parser/infrastructure/ethereum"
	"github.com/mateeullahmalik/eh_parser/parser/infrastructure/store/memory"
)

func main() {
	// example of how to use the parser
	ethereumClient := ethereum.NewClient(ethereum.NewConfig())

	txnsParser := parser.NewClient(
		infraEth.NewEthereumBlockchain(ethereumClient),
		memory.NewTransactionMemoryStore(),
	)

	if err := txnsParser.Run(context.Background()); err != nil {
		panic(err) // To Do: use a better error handling mechanism
	}

	txnsParser.Subscribe("0x1234567890abcdef1234567890abcdef12345678")
	fmt.Println("Subscribed to address 0x1234567890abcdef1234567890abcdef12345678")

	// Get transactions for the address won't work withouth connecting to the Ethereum blockchain
	// txns, err := txnsParser.GetTransactions("0x1234567890abcdef1234567890abcdef12345678")
	// if err != nil {
	// 	panic(err)
	// }
}
