package model

type Account struct {
	ObjectType string `json:"type"` //docType is used to distinguish the various types of objects in state database
	DidName    string `json:"name"`
	CommonName string `json:"cn"`
	MspID      string `json:"mspid"`
	Status 		bool `json:"status"`
}

type Token struct {
	ObjectType string `json:"type"`
	Name       string `json:"name"`    // stock name, eg: "AAPL"
	Desc       string `json:"desc"`    // description
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

type SignRequest struct {
	ObjectType string `json:"type"` //docType is used to distinguish the various types of objects in state database
	Desc       string `json:"desc"`
	Token      string `json:"token"`
	Sender     string `json:"sender"`
	Receiver   string `json:"receiver"`
	Amount     string `json:"amount"`
}