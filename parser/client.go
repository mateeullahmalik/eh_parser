package parser

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mateeullahmalik/eh_parser/parser/domain"
	"github.com/mateeullahmalik/eh_parser/parser/domain/ethereum"
	"github.com/mateeullahmalik/eh_parser/parser/domain/transaction"
)

type client struct {
	ethClient   ethereum.EthClient
	txnStore    transaction.Repository
	latestBlock int32
	subscribers sync.Map
	isRunning   int32 // atomic; 0 means not running, 1 means running
}

func NewClient(eth ethereum.EthClient, store transaction.Repository) *client {
	return &client{
		ethClient: eth,
		txnStore:  store,
	}
}

func (c *client) Subscribe(address string) (loaded bool, err error) {
	if atomic.LoadInt32(&c.isRunning) == 0 {
		return false, fmt.Errorf("cannot subscribe while parser is not running")
	}

	_, loaded = c.subscribers.LoadOrStore(address, struct{}{})
	return !loaded, nil // If the address was already present, loaded is true, and we return false; otherwise, return true
}

func (c *client) GetTransactions(address string) (txns domain.Transactions, err error) {
	if atomic.LoadInt32(&c.isRunning) == 0 {
		return txns, fmt.Errorf("cannot subscribe while parser is not running")
	}

	txns, err = c.txnStore.GetAllByAddress(address)
	if err != nil {
		// To Do: Use a better logging library with structured logging support
		log.Printf("Error retrieving transactions for address %s: %v", address, err)
		return
	}

	return
}

func (c *client) GetCurrentBlock() int {
	return int(atomic.LoadInt32(&c.latestBlock))
}

func (c *client) Run(ctx context.Context) error {
	// Ensure Run is only executed once
	if !atomic.CompareAndSwapInt32(&c.isRunning, 0, 1) {
		return fmt.Errorf("parser is already running")
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				atomic.StoreInt32(&c.isRunning, 0)
				return
			case <-ticker.C:
				c.processNewBlocks(ctx)
			}
		}
	}()

	return nil
}

func (c *client) processNewBlocks(ctx context.Context) {
	blockCount, err := c.ethClient.GetBlockCount(ctx)
	if err != nil {
		log.Printf("Error getting block count: %v", err)
		return
	}

	lastProcessedBlock := atomic.LoadInt32(&c.latestBlock)
	if blockCount <= lastProcessedBlock {
		return
	}

	// Process transactions for each subscribed address
	c.subscribers.Range(func(key, value interface{}) bool {
		address, ok := key.(string)
		if !ok {
			log.Printf("Invalid type for subscriber address")
			return true
		}

		txns, err := c.ethClient.GetTransactionsByAddress(ctx, address, lastProcessedBlock+1)
		if err != nil {
			log.Printf("Error fetching transactions for address %s from block %d: %v", address, lastProcessedBlock+1, err)
			return true
		}

		if err := c.txnStore.SaveAll(txns); err != nil {
			log.Printf("Error storing transaction for address %s: %v", address, err)
		}

		return true
	})

	atomic.StoreInt32(&c.latestBlock, blockCount)
}
