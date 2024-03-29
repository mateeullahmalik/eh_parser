package ethereum

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/mateeullahmalik/eh_parser/ethereum/jsonrpc"
)

type TransactionResults []TransactionResult

// TransactionResult struct to hold individual transaction details
type TransactionResult struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"` // You might want to convert this to an integer
	From             string `json:"from"`
	Gas              string `json:"gas"`      // You might want to convert this to an integer
	GasPrice         string `json:"gasPrice"` // You might want to convert this to an integer
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"` // You might want to convert this to an integer
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"` // You might want to convert this to an integer
	Value            string `json:"value"`            // You might want to convert this to an integer or big.Int for handling large values
	V                string `json:"v"`
	R                string `json:"r"`
	S                string `json:"s"`
}

// Block struct to hold block details and an array of Transactions
type Block struct {
	Difficulty       string             `json:"difficulty"`
	ExtraData        string             `json:"extraData"`
	GasLimit         string             `json:"gasLimit"`
	GasUsed          string             `json:"gasUsed"`
	Hash             string             `json:"hash"`
	LogsBloom        string             `json:"logsBloom"`
	Miner            string             `json:"miner"`
	MixHash          string             `json:"mixHash"`
	Nonce            string             `json:"nonce"`
	Number           string             `json:"number"`
	ParentHash       string             `json:"parentHash"`
	ReceiptsRoot     string             `json:"receiptsRoot"`
	Sha3Uncles       string             `json:"sha3Uncles"`
	Size             string             `json:"size"`
	StateRoot        string             `json:"stateRoot"`
	Timestamp        string             `json:"timestamp"`
	TotalDifficulty  string             `json:"totalDifficulty"`
	Transactions     TransactionResults `json:"transactions"`
	TransactionsRoot string             `json:"transactionsRoot"`
}

type client struct {
	jsonrpc.RPCClient
}

func (client *client) GetLatestBlockNumber(ctx context.Context) (int32, error) {
	res, err := client.CallWithContext(ctx, "eth_blockNumber")
	if err != nil {
		return 0, fmt.Errorf("failed to call eth_blockNumber: %w", err)
	}

	if res.Error != nil {
		return 0, fmt.Errorf("failed to get block number: %w", res.Error)
	}

	cnt, err := res.GetInt()

	return int32(cnt), err
}

func (client *client) GetBlockTransactions(ctx context.Context, block int32) (TransactionResults, error) {
	result := Block{}
	if err := client.callFor(ctx, &result, "eth_getBlockByNumber", block, true); err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	return result.Transactions, nil
}

func (client *client) callFor(ctx context.Context, object interface{}, method string, params ...interface{}) error {
	return client.CallForWithContext(ctx, object, method, params)
}

// NewClient returns a new Client instance.
func NewClient(config *Config) *client {
	//Configure network addressing
	endpoint := net.JoinHostPort(config.Hostname, strconv.Itoa(config.Port))
	if !strings.Contains(endpoint, "//") {
		endpoint = "http://" + endpoint
	}

	//Parse and configure RPC authorization headers
	opts := &jsonrpc.RPCClientOpts{
		CustomHeaders: map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(config.Username+":"+config.Password)),
		},
	}

	//Return a Client interface with the proper RPCClient configurations
	return &client{
		RPCClient: jsonrpc.NewClientWithOpts(endpoint, opts),
	}
}
