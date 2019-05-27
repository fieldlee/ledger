package log

import "github.com/hyperledger/fabric/core/chaincode/shim"

var Logger *shim.ChaincodeLogger

func init() {
	Logger = shim.NewLogger("LedgerTrace")
	Logger.SetLevel(shim.LogCritical)
	Logger.IsEnabledFor(shim.LogCritical)
}

