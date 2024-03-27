package transaction

import (
	"github.com/mateeullahmalik/eh_parser/parser/domain"
)

type Repository interface {
	ReadRepository
	WriteRepository
}

type ReadRepository interface {
	GetAllByAddress(address string) (domain.Transactions, error)
}

type WriteRepository interface {
	Save(tx domain.Transaction) error
	SaveAll(tx domain.Transactions) error
}
