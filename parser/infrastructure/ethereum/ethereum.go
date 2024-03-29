package ethereum

import (
	"context"

	"github.com/mateeullahmalik/eh_parser/ethereum"
	"github.com/mateeullahmalik/eh_parser/parser/domain"
)

type EthereumBlockchain struct {
	client ethereum.Client
}

func NewEthereumBlockchain(client ethereum.Client) *EthereumBlockchain {
	return &EthereumBlockchain{
		client: client,
	}
}

func (e *EthereumBlockchain) GetBlockCount(ctx context.Context) (int32, error) {
	return e.client.GetLatestBlockNumber(ctx)
}

func (e *EthereumBlockchain) GetTransactionsWithAddressesFilter(ctx context.Context, block int32, addresses ...string) (txns domain.Transactions, err error) {
	transactions, err := e.client.GetBlockTransactions(ctx, block)
	if err != nil {
		return txns, err
	}

	txnsMap := make(map[string]domain.Transactions)
	for _, addr := range addresses {
		txnsMap[addr] = domain.Transactions{}
	}

	count := 0
	for _, tx := range transactions {
		_, fromExists := txnsMap[tx.From]
		_, toExists := txnsMap[tx.To]

		if fromExists || toExists {
			count++
			txnsMap[tx.From] = append(txnsMap[tx.From], domain.Transaction{
				TxID:     tx.Hash,
				Gas:      tx.Gas,
				From:     tx.From,
				To:       tx.To,
				GasPrice: tx.GasPrice,
				Value:    tx.Value,
				Block:    block,
			})
		}
	}

	txns = make(domain.Transactions, count)
	i := 0
	for _, tx := range txnsMap {
		for _, t := range tx {
			txns[i] = t
			i++
		}
	}

	return txns, nil
}
