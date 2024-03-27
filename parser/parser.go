package parser

import "github.com/mateeullahmalik/eh_parser/parser/domain"

type Parser interface {
	// GetCurrentBlock returns last parsed block
	GetCurrentBlock() int

	// Subscribe adds address to observer
	Subscribe(address string) (bool, error)

	// GetTransactions returns list of inbound or outbound transactions for an address
	GetTransactions(address string) (domain.Transactions, error)
}
