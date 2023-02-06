package subscan

import "time"

type List struct {
	Symbol string `json:"symbol"`
	From   string `json:"from"`
	To     string `json:"to"`
	Value  int64  `json:"value"`
	Failed bool   `json:"failed"`
}

type Transaction struct {
	Height    int64     `json:"height"`
	Hash      string    `json:"hash"`
	IsSuccess bool      `json:"isSuccess"`
	Fee       int64     `json:"fee"`
	List      []List    `json:"list"`
	Timestamp time.Time `json:"timestamp"`
	Executor  string    `json:"executor"`
}

type Block struct {
	Data []Transaction `json:"data"`
}
