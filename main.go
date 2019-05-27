package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"ledger/common"
	"ledger/services"
	"strings"
)
type LedgerChainCode struct {

}

func(lcc *LedgerChainCode)Init(stub shim.ChaincodeStubInterface)pb.Response{
	return shim.Success(nil)
}

func(lcc *LedgerChainCode)Invoke(stub shim.ChaincodeStubInterface)pb.Response{
	funcname,_ := stub.GetFunctionAndParameters()

	switch  strings.ToLower(funcname)  {
	case "version":
		return shim.Success([]byte(common.VERSION))
	case "account_register":
		return services.AccountRegister()

	}

	return shim.Success(nil)
}

func main(){
	err := shim.Start(new(LedgerChainCode))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err.Error())
	}
}
