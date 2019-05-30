package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"ledger/common"
	"ledger/log"
	"ledger/model"
	"strconv"
	"strings"
	"time"
)

/*token 發行*/
func LedgerIssue(stub shim.ChaincodeStubInterface)pb.Response{

	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	issueJson := args[0]
	log.Logger.Info(issueJson)
    issueParam   :=	model.LedgerIssueParam{}
    err := json.Unmarshal([]byte(issueJson),&issueParam)
    if err != nil {
    	return shim.Error(err.Error())
	}

    token, err := TokenGet(stub,issueParam.Token)
	if err != nil {
		return shim.Error(err.Error())
	}

    curUserName,err  := common.GetCommonName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
    //// super admin
    if ! common.IsSuperAdmin(stub) {
		/// 發行人
		if strings.ToUpper(token.Issuer) != strings.ToUpper(curUserName) {
			return common.SendError(common.Right_ERR,"only super admin and token issuer can issue the token")
		}
	}

	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s token not enable",token.Name))
	}

	holder, err  := common.GetCommonName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if common.IsSuperAdmin(stub){
		if issueParam.Holder == "" {
			return  common.SendError(common.Param_ERR,"the token holder not allownce empty")
		}

		accout, err := AccountGetByName(stub,issueParam.Holder)
		if err != nil {
			return shim.Error(err.Error())
		}

		if accout.Status == false {
			return  common.SendError(common.ACCOUNT_NOT_EXIST,"the holder is not exist or the holder is disable")
		}
		holder = accout.CommonName
	}


	log.Logger.Info("holder:",holder)

	leder := model.Ledger{}

	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(holder)})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, holder, err.Error()))
	}

	leder.Holder = holder
	leder.Token = token.Name
	leder.Amount = issueParam.Amount
	leder.Desc = fmt.Sprintf("%s issue %s token amount:%s",curUserName,token.Name,strconv.FormatFloat(issueParam.Amount,'f',2,64))

	ledgerByte, err := json.Marshal(leder)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key,ledgerByte)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
// get balance
func LedgerGetBalance(stub shim.ChaincodeStubInterface)pb.Response  {

	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	balancejson := args[0]
	log.Logger.Info(balancejson)
	balance := model.LedgerBalanceParam{}

	err  := json.Unmarshal([]byte(balancejson),&balance)

	if err != nil {
		return shim.Error(err.Error())
	}

	account, err := AccountGetByName(stub,balance.Holder)
	if err != nil {
		return shim.Error(err.Error())
	}
	token , err := TokenGet(stub,balance.Token)
	if err != nil {
		return shim.Error(err.Error())
	}
	if account.Status == false {
		return  common.SendError(common.ACCOUNT_NOT_EXIST,"the holder is not exist or the holder is disable")
	}
	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s token not enable",token.Name))
	}

	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(account.CommonName)})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, account.CommonName, err.Error()))
	}

	ledgerByte,err := stub.GetState(key)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(ledgerByte)
}
// get history
func LedgerGetHistory(stub shim.ChaincodeStubInterface)pb.Response{

	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	balancejson := args[0]

	balance := model.LedgerBalanceParam{}

	err  := json.Unmarshal([]byte(balancejson),&balance)

	if err != nil {
		return shim.Error(err.Error())
	}

	account, err := AccountGetByName(stub,balance.Holder)
	if err != nil {
		return shim.Error(err.Error())
	}
	token , err := TokenGet(stub,balance.Token)
	if err != nil {
		return shim.Error(err.Error())
	}
	if account.Status == false {
		return  common.SendError(common.ACCOUNT_NOT_EXIST,"the holder is not exist or the holder is disable")
	}
	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s token not enable",token.Name))
	}

	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(account.CommonName)})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, account.CommonName, err.Error()))
	}

	history, err := stub.GetHistoryForKey(key)

	if err != nil {
		return shim.Error(err.Error())
	}

	defer  history.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false

	for history.HasNext(){
		response ,err := history.Next()
		if err != nil {
			continue
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}
		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

func LedgerTransfer(stub shim.ChaincodeStubInterface)pb.Response{
	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	transferjson := args[0]
	log.Logger.Info(transferjson)
	transfer := model.LedgerTransferParam{}
	err := json.Unmarshal([]byte(transferjson),&transfer)
	if err != nil {
		return shim.Error(err.Error())
	}

	curUserName,err  := common.GetCommonName(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	if strings.ToUpper(strings.TrimSpace(curUserName)) != strings.ToUpper(strings.TrimSpace(transfer.From)){
		return common.SendError(common.ACCOUNT_PREMISSION,fmt.Sprintf("%s is not current login user :%s",transfer.From,curUserName))
	}

	accountFrom, err := AccountGetByName(stub,transfer.From)
	if err != nil {
		return shim.Error(err.Error())
	}
	accountTo, err := AccountGetByName(stub,transfer.To)
	if err != nil {
		return shim.Error(err.Error())
	}

	if accountFrom.CommonName == accountTo.CommonName {
		return common.SendError(common.ACCOUNT_PREMISSION,fmt.Sprintf("from : %s can not equal to user :%s",transfer.From,transfer.To))
	}

	if accountFrom.Status == false || accountTo.Status == false {
		return common.SendError(common.ACCOUNT_LOCK,fmt.Sprintf("from : %s OR to  :%s is locked",transfer.From,transfer.To))
	}

	token , err := TokenGet(stub,transfer.Token)
	if err != nil {
		return shim.Error(err.Error())
	}

	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s Token is disable",token.Name))
	}

	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(accountFrom.CommonName)})
	if err != nil {
		return shim.Error(fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, accountFrom.CommonName, err.Error()))
	}

	ledgerByte,err := stub.GetState(key)
	if err != nil {
		return shim.Error(err.Error())
	}

	ledger := model.Ledger{}

	err = json.Unmarshal(ledgerByte,&ledger)
	if err != nil {
		return shim.Error(err.Error())
	}
	if ledger.Amount < transfer.Amount {
		return common.SendError(common.Balance_NOT_ENOUGH,fmt.Sprintf("the %s token balance not enough",token.Name))
	}
	ledger.Amount = ledger.Amount - transfer.Amount
	ledger.Desc = fmt.Sprintf("From : %s transfer To : %s , value : %s ",accountFrom.CommonName,accountTo.CommonName,strconv.FormatFloat(transfer.Amount,'f',2,64))

	ledgerByted , err := json.Marshal(ledger)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(key,ledgerByted)
	if err != nil {
		return shim.Error(err.Error())
	}



	////////////////// send event
	ts, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error(err.Error())
	}
	//////// set event
	evt := model.LedgerEvent{
		Type: common.Evt_payment,
		Txid:   stub.GetTxID(),
		Time:   ts.GetSeconds(),
		From:   transfer.From,
		To:     transfer.To,
		Amount: strconv.FormatFloat(transfer.Amount,'f',2,64) ,
		Token:  transfer.Token,
	}

	eventJSONasBytes, err := json.Marshal(evt)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.SetEvent(fmt.Sprintf(common.TOPIC, transfer.To), eventJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func LedgerScale(stub shim.ChaincodeStubInterface)pb.Response  {
	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	scaleJson := args[0]
	log.Logger.Info(scaleJson)
	scaleParam := model.LedgerScaleParam{}

	err := json.Unmarshal([]byte(scaleJson),&scaleParam)

	if err != nil {
		return shim.Error(err.Error())
	}

	token, err := TokenGet(stub,scaleParam.Token)

	if err != nil {
		return shim.Error(err.Error())
	}

	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s Token is disable",token.Name))
	}

	if common.IsSuperAdmin(stub) == false {
		return common.SendError(common.ACCOUNT_PREMISSION,"only super admin can break up or merge operation")
	}

	resultIterator, err := stub.GetStateByPartialCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE,strings.ToUpper(token.Name)})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultIterator.Close()

	var i int
	for i=0; resultIterator.HasNext();i++ {
		iterObj,err := resultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		key := iterObj.Key
		ledger := model.Ledger{}
		err = json.Unmarshal(iterObj.Value,&ledger)
		if err != nil {
			return shim.Error(err.Error())
		}

		ledger.Amount = ledger.Amount * scaleParam.Scale

		if scaleParam.Scale > float64(1.0) {
			ledger.Desc =  fmt.Sprintf("%s token break up , breake up scale %s",scaleParam.Token, strconv.FormatFloat(scaleParam.Scale,'f',2,64) )
		}else{
			ledger.Desc =  fmt.Sprintf("%s token merge , merge scale %s",scaleParam.Token, strconv.FormatFloat(scaleParam.Scale,'f',2,64) )
		}

		ledgerByte,err  := json.Marshal(ledger)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(key,ledgerByte)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	returnString := fmt.Sprintf("had scale %d token holders",i)
	return shim.Success([]byte(returnString))

}
