package common


const ADMIN_ORG = "Org1MSP"
const ADMIN_Name = "Admin"

const VERSION = "Ledger-Code v1.0"

const ACCOUNT = "account"
const TOKEN = "token"

const ACCOUNT_PRE = "ACCOUNT_"
const TOKEN_PRE = "TOKEN_"
const Ledger_PRE = "LEDGER"
const SIGN_PRE = "SIGN"

const CompositeIndexName = "pre~tkn~name"
const CompositeRequestIndexName = "pre~name~txid"

const (
	TKNERR_EXIST   = 501
	TKNERR_LOCKED  = 502
	TKNERR_PREMISSON	 = 503
)

const (
	ACCOUNT_EXIST 	  = 501
	ACCOUNT_NOT_EXIST = 502
	ACCOUNT_PREMISSION = 503
	ACCOUNT_LOCK = 504
)

const (
	Right_ERR = 504
	Param_ERR = 505
	Balance_NOT_ENOUGH = 506
)

const(
	PENDING_SIGN = "Pending"
	SENT_SIGN 	= "Sent"
	Failed_SIGN = "Refused"
)