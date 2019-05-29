package model

type Account struct {
	DidName    string `json:"name"`
	CommonName string `json:"cn"`
	MspID      string `json:"mspid"`
	Status 		bool `json:"status"`
}

type Token struct {
	Name       string `json:"name"`    // stock name, eg: "AAPL"
	Desc       string `json:"desc"`    // description
	Issuer		string `json:"issuer"`
	Status     bool   `json:"status"`
}

type LedgerEvent struct {
	Type   uint8  `json:"type"`
	Txid   string `json:"txid"`
	Time   int64  `json:"time"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	Token  string `json:"token"`
}

type Ledger struct {
	Token string `json:"token"`
	Holder string `json:"holder"`
	Amount float64 `json:"amount"`
	Desc string `json:"desc"`
}

type SignRequest struct {
	TxID 	   string `json:"txid"`
	Desc       string `json:"desc"`
	Token      string `json:"token"`
	Sender     string `json:"sender"`
	Receiver   string `json:"receiver"`
	Amount     float64 `json:"amount"`
	Status 	   string `json:"status"`
}

type LedgerIssueParam struct {
	Holder  string `json:"holder"`
	Token 	string `json:"token"`
	Amount  float64 `json:"amount"`
}

type LedgerBalanceParam struct {
	Holder  string `json:"holder"`
	Token 	string `json:"token"`
}

type LedgerTransferParam struct {
	From  	string `json:"from"`
	To 		string `json:"to"`
	Amount  float64 `json:"amount"`
	Token 	string `json:"token"`
}

type LedgerScaleParam struct {
	Token 	string `json:"token"`
	Scale 	float64 `json:"scale"`
}

type LedgerRequestParam struct {
	Desc       string `json:"desc"`
	Token      string `json:"token"`
	Sender     string `json:"sender"`
	Receiver   string `json:"receiver"`
	Amount     float64 `json:"amount"`
}

type LedgerResponseParam struct {
	Accept bool `json:"accept"`
	Desc       string `json:"desc"`
	Txid      string `json:"txid"`
	Sender     string `json:"sender"`
}