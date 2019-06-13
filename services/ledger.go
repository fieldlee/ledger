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
		return common.SendError(common.Param_ERR,"Parameters error ,please check Parameters")
	}
	issueJson := args[0]
	log.Logger.Info(issueJson)
    issueParam   :=	model.LedgerIssueParam{}
    err := json.Unmarshal([]byte(issueJson),&issueParam)
    if err != nil {
		log.Logger.Error("Unmarshal:",err)
    	return common.SendError(common.MARSH_ERR,err.Error())
	}

    token, err := TokenGet(stub,issueParam.Token)
	if err != nil {
		log.Logger.Error("TokenGet:",err)
		return common.SendError(common.MARSH_ERR,err.Error())
	}

    curUserName,err  := common.GetCommonName(stub)
	if err != nil {
		log.Logger.Error("GetCommonName:",err)
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}
    //// super admin
    if ! common.IsSuperAdmin(stub) {
		return common.SendError(common.Right_ERR,"only super admin  can issue the token")
	}
	///// token enable
	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s token not enable",token.Name))
	}
	////// issue to super admin
	issueParam.Holder = strings.ToUpper(strings.TrimSpace(curUserName))

	accout, err := AccountGetByName(stub,issueParam.Holder)
	if err != nil {
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}

	if accout.Status == false {
		return  common.SendError(common.ACCOUNT_NOT_EXIST,"the holder is not exist or the holder is disable")
	}

	holder := accout.DidName

	log.Logger.Info("holder:",holder)
	leder := model.Ledger{}
	leder.Type = common.LEDGER

	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(holder)})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, holder, err.Error()))
	}

	issueByte,_ := stub.GetState(key)
	////////////////////////===========================多次ｉｓｓｕｅ
	if issueByte != nil {
		err = json.Unmarshal(issueByte,&leder)
		if err != nil {
			return common.SendError(common.MARSH_ERR,err.Error())
		}
		leder.Amount = leder.Amount + issueParam.Amount
		leder.Desc = fmt.Sprintf("%s issue %s token amount:%s",curUserName,token.Name,strconv.FormatFloat(issueParam.Amount,'f',2,64))
	}else{
		leder.Holder = holder
		leder.Token = token.Name
		leder.Amount = issueParam.Amount
		leder.Desc = fmt.Sprintf("%s issue %s token amount:%s",curUserName,token.Name,strconv.FormatFloat(issueParam.Amount,'f',2,64))
	}

	ledgerByte, err := json.Marshal(leder)
	if err != nil {
		log.Logger.Error("Marshal:",err)
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	err = stub.PutState(key,ledgerByte)

	if err != nil {
		log.Logger.Error("PutState:",err)
		return common.SendError(common.PUTSTAT_ERR,err.Error())
	}

	///////////////================修改token
	if token.Amount > 0 {
		token.Amount = token.Amount + issueParam.Amount
	}else{
		token.Amount =  issueParam.Amount
	}

	token.Desc = fmt.Sprintf("%s issue amount :%s",token.Name,strconv.FormatFloat(issueParam.Amount,'f',2,64))
	tokenByted , err := json.Marshal(token)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	err = stub.PutState(common.TOKEN_PRE + token.Name,tokenByted)
	if err != nil {
		return common.SendError(common.PUTSTAT_ERR,err.Error())
	}

	return common.SendScuess(fmt.Sprintf("%s token had issue",token.Name))
}
// get balance
func LedgerGetBalance(stub shim.ChaincodeStubInterface)pb.Response  {

	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return common.SendError(common.Param_ERR,"Parameters error ,please check Parameters")
	}

	balancejson := args[0]
	log.Logger.Info(balancejson)

	balance := model.LedgerBalanceParam{}

	err  := json.Unmarshal([]byte(balancejson),&balance)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	/////////////////////////////// ==============account accountFrom, err := AccountGetByName(stub,transfer.From)
	account,err := AccountGetByName(stub,balance.Holder)
	if err != nil {
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}
	////////////////////////////// =============token
	token , err := TokenGet(stub,balance.Token)
	if err != nil {
		log.Logger.Error("TokenGet:",err)
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	/////////////////////////////// =================== token and account check
	if account.Status == false {
		return  common.SendError(common.ACCOUNT_NOT_EXIST,"the holder is not exist or the holder is disable")
	}
	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s token not enable",token.Name))
	}

	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(account.DidName)})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, account.DidName, err.Error()))
	}

	ledgerByte,err := stub.GetState(key)

	if err != nil {
		return common.SendError(common.GETSTAT_ERR,err.Error())
	}
	return common.SendScuess(string(ledgerByte))
}
// get history
func LedgerGetHistory(stub shim.ChaincodeStubInterface)pb.Response{

	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return common.SendError(common.Param_ERR,"Parameters error ,please check Parameters")
	}

	balancejson := args[0]

	balance := model.LedgerBalanceParam{}

	err  := json.Unmarshal([]byte(balancejson),&balance)

	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	account, err := AccountGetByName(stub,balance.Holder)
	if err != nil {
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}

	token , err := TokenGet(stub,balance.Token)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	if account.Status == false {
		return  common.SendError(common.ACCOUNT_NOT_EXIST,"the holder is not exist or the holder is disable")
	}

	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s token not enable",token.Name))
	}

	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(account.DidName)})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, account.DidName, err.Error()))
	}

	history, err := stub.GetHistoryForKey(key)
	if err != nil {
		return common.SendError(common.GETSTAT_ERR,err.Error())
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

	return common.SendScuess(buffer.String())
}

func LedgerBurnBalance(stub shim.ChaincodeStubInterface)pb.Response{
	_,args := stub.GetFunctionAndParameters()
	if len(args) != 1{
		return common.SendError(common.Param_ERR,"Parameters error ,please check Parameters")
	}
	burnJson := args[0]
	burnParam   :=	model.LedgerBurnParam{}
	err := json.Unmarshal([]byte(burnJson),&burnParam)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	token, err := TokenGet(stub,burnParam.Token)
	if err != nil {
		log.Logger.Error("TokenGet:",err)
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	curUserName,err  := common.GetCommonName(stub)
	if err != nil {
		log.Logger.Error("GetCommonName:",err)
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}
	//// super admin
	if ! common.IsSuperAdmin(stub) {
		return common.SendError(common.Right_ERR,"only super admin  can issue the token")
	}
	///// token enable
	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s token not enable",token.Name))
	}
	//////////// holder
	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(curUserName)})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, curUserName, err.Error()))
	}

	ledgerByte,err := stub.GetState(key)
	if err != nil {
		return common.SendError(common.GETSTAT_ERR,err.Error())
	}

	ledger := model.Ledger{}
	err = json.Unmarshal(ledgerByte,&ledger)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	if ledger.Amount < burnParam.Amount {
		return common.SendError(common.Balance_NOT_ENOUGH,fmt.Sprintf("the %s token balance not enough",token.Name))
	}
	ledger.Amount = ledger.Amount - burnParam.Amount
	ledger.Desc = fmt.Sprintf("%s token burned amount :%s ",token.Name,strconv.FormatFloat(burnParam.Amount,'f',2,64))

	ledgerByted , err := json.Marshal(ledger)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	err = stub.PutState(key,ledgerByted)
	if err != nil {
		return common.SendError(common.PUTSTAT_ERR,err.Error())
	}
	///////////////================修改token
	token.Amount = token.Amount - burnParam.Amount
	token.Desc = fmt.Sprintf("%s Token burned amount :%s",token.Name,strconv.FormatFloat(burnParam.Amount,'f',2,64))
	tokenByted , err := json.Marshal(token)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	err = stub.PutState(common.TOKEN_PRE + token.Name,tokenByted)
	if err != nil {
		return common.SendError(common.PUTSTAT_ERR,err.Error())
	}
	return common.SendScuess(fmt.Sprintf("had burned amount %s  %s token",strconv.FormatFloat(burnParam.Amount,'f',2,64),token.Name))
}

func LedgerTransfer(stub shim.ChaincodeStubInterface)pb.Response{
	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return common.SendError(common.Param_ERR,"Parameters error ,please check Parameters")
	}

	transferjson := args[0]
	log.Logger.Info(transferjson)
	transfer := model.LedgerTransferParam{}
	err := json.Unmarshal([]byte(transferjson),&transfer)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	curUserName,err  := common.GetCommonName(stub)
	if err != nil {
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}

	if strings.ToUpper(strings.TrimSpace(curUserName)) != strings.ToUpper(strings.TrimSpace(transfer.From)){
		return common.SendError(common.ACCOUNT_PREMISSION,fmt.Sprintf("%s is not current login user :%s",transfer.From,curUserName))
	}

	accountFrom, err := AccountGetByName(stub,transfer.From)
	if err != nil {
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}
	accountTo, err := AccountGetByName(stub,transfer.To)
	if err != nil {
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}

	if accountFrom.DidName == accountTo.DidName {
		return common.SendError(common.ACCOUNT_PREMISSION,fmt.Sprintf("from : %s can not equal to user :%s",transfer.From,transfer.To))
	}

	if accountFrom.Status == false || accountTo.Status == false {
		return common.SendError(common.ACCOUNT_LOCK,fmt.Sprintf("from : %s OR to  :%s is locked",transfer.From,transfer.To))
	}

	token , err := TokenGet(stub,transfer.Token)
	if err != nil {
		log.Logger.Error("TokenGet:",err)
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s Token is disable",token.Name))
	}

	//////////// from
	key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(accountFrom.DidName)})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, accountFrom.DidName, err.Error()))
	}

	ledgerByte,err := stub.GetState(key)
	if err != nil {
		log.Logger.Error("GetState:",err)
		return common.SendError(common.GETSTAT_ERR,err.Error())
	}


	///////////////////////////from
	ledger := model.Ledger{}

	err = json.Unmarshal(ledgerByte,&ledger)
	if err != nil {
		log.Logger.Error("Unmarshal:",err)
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	if ledger.Amount < transfer.Amount {
		return common.SendError(common.Balance_NOT_ENOUGH,fmt.Sprintf("the %s token balance not enough",token.Name))
	}

	ledger.Amount = ledger.Amount - transfer.Amount

	ledger.Desc = fmt.Sprintf("From : %s transfer To : %s , value : %s ",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(transfer.Amount,'f',2,64))

	ledgerByted , err := json.Marshal(ledger)
	if err != nil {
		log.Logger.Error("Marshal:",err)
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	err = stub.PutState(key,ledgerByted)
	if err != nil {
		log.Logger.Error("PutState:",err)
		return common.SendError(common.PUTSTAT_ERR,err.Error())
	}
	//////////////////////to
	tokey, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(accountTo.DidName)})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, accountTo.DidName, err.Error()))
	}

	toledgerByte,err := stub.GetState(tokey)
	if err != nil{
		return common.SendError(common.PUTSTAT_ERR,err.Error())
	}

	toledger := model.Ledger{}
	toledger.Type = common.LEDGER
	if toledgerByte == nil {
		toledger.Holder = strings.ToUpper(accountTo.DidName)
		toledger.Token = strings.ToUpper(token.Name)
		toledger.Desc = fmt.Sprintf("From : %s transfer To : %s , value : %s ",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(transfer.Amount,'f',2,64))
		toledger.Amount = toledger.Amount + transfer.Amount
	}else {
		err = json.Unmarshal(toledgerByte,&toledger)
		if err != nil {
			return common.SendError(common.MARSH_ERR,err.Error())
		}
		toledger.Desc = fmt.Sprintf("From : %s transfer To : %s , value : %s ",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(transfer.Amount,'f',2,64))
		toledger.Amount = toledger.Amount + transfer.Amount
	}
	toTransferedByted , err := json.Marshal(toledger)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	err = stub.PutState(tokey,toTransferedByted)
	if err != nil {
		return common.SendError(common.PUTSTAT_ERR,err.Error())
	}

	////////////////// send event
	ts, err := stub.GetTxTimestamp()
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
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
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	err = stub.SetEvent(fmt.Sprintf(common.TOPIC, transfer.To), eventJSONasBytes)
	if err != nil {
		return common.SendError(common.EVENT_ERR,err.Error())
	}

	return common.SendScuess(fmt.Sprintf("From : %s transfer To : %s , value : %s ",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(transfer.Amount,'f',2,64)))
}

func LedgerScale(stub shim.ChaincodeStubInterface)pb.Response  {
	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return common.SendError(common.Param_ERR,"Parameters error ,please check Parameters")
	}

	scaleJson := args[0]
	log.Logger.Info(scaleJson)
	scaleParam := model.LedgerScaleParam{}

	err := json.Unmarshal([]byte(scaleJson),&scaleParam)

	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	token, err := TokenGet(stub,scaleParam.Token)

	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	if token.Status == true {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s Token is enable,must lock first",token.Name))
	}

	if common.IsSuperAdmin(stub) == false {
		return common.SendError(common.ACCOUNT_PREMISSION,"only super admin can break up or merge operation")
	}

	if scaleParam.Deno == 0 {
		return common.SendError(common.Param_ERR,"deno must Integers greater than 0 ")
	}

	///////////////////////////////////////////////////////================================修改持有token賬戶
	resultIterator, err := stub.GetStateByPartialCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE,strings.ToUpper(token.Name)})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s: %s", token.Name, err.Error()))
	}
	defer resultIterator.Close()

	var i int
	for i=0; resultIterator.HasNext();i++ {
		iterObj,err := resultIterator.Next()
		if err != nil {
			return common.SendError(common.MARSH_ERR,err.Error())
		}
		key := iterObj.Key
		log.Logger.Info(key)
		ledger := model.Ledger{}
		err = json.Unmarshal(iterObj.Value,&ledger)
		if err != nil {
			return  common.SendError(common.MARSH_ERR,err.Error())
		}
		ledger.Amount = common.ComputeForMD(ledger.Amount,scaleParam.Mole,scaleParam.Deno)
		if common.ComMD(scaleParam.Mole,scaleParam.Deno) > float64(1.0) {
			ledger.Desc =  fmt.Sprintf("%s token break up , breake up scale %s",scaleParam.Token, strconv.FormatFloat(common.ComMD(scaleParam.Mole,scaleParam.Deno),'f',2,64) )
		}else{
			ledger.Desc =  fmt.Sprintf("%s token merge , merge scale %s",scaleParam.Token, strconv.FormatFloat(common.ComMD(scaleParam.Mole,scaleParam.Deno),'f',2,64) )
		}
		ledgerByte,err  := json.Marshal(ledger)
		if err != nil {
			return  common.SendError(common.MARSH_ERR,err.Error())
		}
		err = stub.PutState(key,ledgerByte)
		if err != nil {
			return  common.SendError(common.PUTSTAT_ERR,err.Error())
		}
	}
	///////////////////////////////////////////////////////===============================修改未簽名的轉賬
	signIter , err := stub.GetStateByPartialCompositeKey(common.CompositeRequestIndexName,[]string{common.SIGN_PRE,strings.ToUpper(token.Name)})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,err.Error())
	}
	defer signIter.Close()
	var j int
	for j=0;signIter.HasNext();j++{
		iterSignObj,err := signIter.Next()
		if err != nil {
			return common.SendError(common.MARSH_ERR,err.Error())
		}
		signkey := iterSignObj.Key

		log.Logger.Info(signkey)

		signReq := model.SignRequest{}
		err = json.Unmarshal(iterSignObj.Value,&signReq)
		if err != nil {
			return common.SendError(common.MARSH_ERR,err.Error())
		}
		if signReq.Status == common.PENDING_SIGN {
			signReq.Amount = common.ComputeForMD(signReq.Amount,scaleParam.Mole,scaleParam.Deno)
			////////// desc info
			if common.ComMD(scaleParam.Mole,scaleParam.Deno) > float64(1.0) {
				signReq.Desc =  fmt.Sprintf("%s token break up , breake up scale %s",scaleParam.Token, strconv.FormatFloat(common.ComMD(scaleParam.Mole,scaleParam.Deno),'f',2,64) )
			}else{
				signReq.Desc =  fmt.Sprintf("%s token merge , merge scale %s",scaleParam.Token, strconv.FormatFloat(common.ComMD(scaleParam.Mole,scaleParam.Deno),'f',2,64) )
			}
			ledgerByte,err  := json.Marshal(signReq)

			if err != nil {
				return common.SendError(common.MARSH_ERR,err.Error())
			}
			err = stub.PutState(signkey,ledgerByte)

			if err != nil {
				return common.SendError(common.PUTSTAT_ERR,err.Error())
			}
		}else{
			continue
		}
	}

	///////////////////////////////////////////////////================================修改token总发行量
	token.Amount = common.ComputeForMD(token.Amount,scaleParam.Mole,scaleParam.Deno)

	if common.ComMD(scaleParam.Mole,scaleParam.Deno) > float64(1.0) {
		token.Desc =  fmt.Sprintf("%s token break up , breake up scale %s",scaleParam.Token, strconv.FormatFloat(common.ComMD(scaleParam.Mole,scaleParam.Deno),'f',2,64) )
	}else{
		token.Desc =  fmt.Sprintf("%s token merge , merge scale %s",scaleParam.Token, strconv.FormatFloat(common.ComMD(scaleParam.Mole,scaleParam.Deno),'f',2,64) )
	}
	tokenBytes,err := json.Marshal(token)
	if err != nil {
		return  common.SendError(common.MARSH_ERR,err.Error())
	}
	err = stub.PutState(common.TOKEN_PRE+token.Name,tokenBytes)
	if err != nil {
		return  common.SendError(common.PUTSTAT_ERR,err.Error())
	}

	returnString := fmt.Sprintf("had scale %d token holders, had scale %d pending for sign tx",i,j)
	return common.SendScuess(returnString)

}

func LedgerGetListbyAccount(stub shim.ChaincodeStubInterface) pb.Response{
	_,args := stub.GetFunctionAndParameters()
	if len(args) != 1{
		return common.SendError(common.Param_ERR,"Parameters error ,please check Parameters")
	}
	accountName := args[0]
	log.Logger.Info(accountName)

	accout, err := AccountGetByName(stub,accountName)
	if err != nil {
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}
	if accout.Status == false {
		return  common.SendError(common.ACCOUNT_NOT_EXIST,"the holder is not exist or the holder is disable")
	}

	listLedger := []model.Ledger{}

	////////===============================get token list
	queryString := "{\"selector\":{\"type\":\"token\"}}"
	resultsIterator, err := stub.GetQueryResult(queryString)
	defer resultsIterator.Close()
	if err != nil {
		return common.SendError(common.GETSTAT_ERR,err.Error())
	}

	for resultsIterator.HasNext() {
		queryResponse,err := resultsIterator.Next()
		if err != nil {
			return common.SendError(common.MARSH_ERR,err.Error())
		}
		tmpToken := model.Token{}
		err = json.Unmarshal(queryResponse.Value,&tmpToken)
		if err != nil {
			return common.SendError(common.MARSH_ERR,err.Error())
		}

		key, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(tmpToken.Name),  strings.ToUpper(accout.DidName)})
		if err != nil {
			return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", tmpToken.Name, accout.DidName, err.Error()))
		}
		issueByte,_ := stub.GetState(key)
		if issueByte != nil {
			tmpLedger := model.Ledger{}
			err = json.Unmarshal(issueByte,&tmpLedger)
			if err != nil{
				return common.SendError(common.MARSH_ERR,err.Error())
			}
			listLedger = append(listLedger,tmpLedger)
		}
	}

	listLedgerBYte,err := json.Marshal(listLedger)

	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	return common.SendScuess(string(listLedgerBYte))
}