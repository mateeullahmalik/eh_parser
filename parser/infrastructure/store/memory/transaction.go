package memory

import (
	"encoding/json"
	"fmt"

	"github.com/mateeullahmalik/eh_parser/common/storage"
	"github.com/mateeullahmalik/eh_parser/common/storage/memory"
	"github.com/mateeullahmalik/eh_parser/parser/domain"
)

type TransactionMemoryStore struct {
	db storage.KeyValue
}

func NewTransactionMemoryStore() *TransactionMemoryStore {
	return &TransactionMemoryStore{
		db: memory.NewKeyValue(),
	}
}

func (t *TransactionMemoryStore) GetAllByAddress(address string) (txns domain.Transactions, err error) {
	txns = make(domain.Transactions, 0)

	data, err := t.db.Get(address)
	if err != nil {
		if err == storage.ErrKeyValueNotFound {
			return txns, nil
		}

		return nil, fmt.Errorf("unable to get transactions for address %s: %w", address, err)
	}

	if err := json.Unmarshal(data, &txns); err != nil {
		return nil, fmt.Errorf("unable to unmarshal transactions for address %s: %w", address, err)
	}

	return
}

func (t *TransactionMemoryStore) insertTransactions(id string, txs domain.Transactions) error {
	txns, err := t.GetAllByAddress(id)
	if err != nil {
		return fmt.Errorf("unable to get transactions for address %s: %w", id, err)
	}

	txns = append(txns, txs...)

	data, err := json.Marshal(txns)
	if err != nil {
		return fmt.Errorf("unable to marshal transactions for address %s: %w", id, err)
	}

	if err := t.db.Set(id, data); err != nil {
		return fmt.Errorf("unable to insert transactions for address %s: %w", id, err)
	}

	return nil
}

// Save inserts the transaction for both the sender and the receiver
// Its understood that there's an overhead of inserting the same transaction twice
// but assuming that (a) we are concered about the overhead at this time
// and (b) the goal here is to keep the implementation simple and flexible for other storage implementations
// for example, the sqlite implmentaion can implement this method in a way that it only inserts the transaction once
func (t *TransactionMemoryStore) Save(tx domain.Transaction) error {
	txs := domain.Transactions{tx}
	if err := t.insertTransactions(tx.From, txs); err != nil {
		return fmt.Errorf("unable to insert transaction for address %s: %w", tx.From, err)
	}

	if err := t.insertTransactions(tx.To, txs); err != nil {
		return fmt.Errorf("unable to insert transaction for address %s: %w", tx.To, err)
	}

	return nil
}

// SaveAll inserts transactions for both the sender and the receiver
// while this is understood that there's an overhead of inserting the same transaction twice
// keeping the interface simple and flexible for other storage implementations where
// we can batch insert transactions for an effecient insert
func (t *TransactionMemoryStore) SaveAll(txs domain.Transactions) error {
	groupedTxs := make(map[string]domain.Transactions)

	for _, tx := range txs {
		groupedTxs[tx.From] = append(groupedTxs[tx.From], tx)
		groupedTxs[tx.To] = append(groupedTxs[tx.To], tx)
	}

	for address, txsForAddress := range groupedTxs {
		if err := t.insertTransactions(address, txsForAddress); err != nil {
			return fmt.Errorf("unable to insert transaction for address %s: %w", address, err)
		}
	}

	return nil
}
