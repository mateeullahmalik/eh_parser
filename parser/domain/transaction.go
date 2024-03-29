package domain

type Transactions []Transaction

type Transaction struct {
	From     string
	To       string
	TxID     string
	Block    int32
	Fee      float64
	Gas      string
	GasPrice string
	Value    string
}
