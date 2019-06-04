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

func SignRequest(stub shim.ChaincodeStubInterface)pb.Response{
	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return common.SendError(common.Param_ERR,"Parameters error ,please check Parameters")
	}

	signRequestStr := args[0]
	log.Logger.Info(signRequestStr)
	request := model.LedgerRequestParam{}

	err := json.Unmarshal([]byte(signRequestStr),&request)

	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}


	currentName,err := common.GetCommonName(stub)
	if err != nil {
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}
	if strings.ToUpper(strings.TrimSpace(currentName)) != strings.ToUpper(strings.TrimSpace(request.Receiver)){
		return common.SendError(common.ACCOUNT_PREMISSION,fmt.Sprintf("current login user:%s not equal receiver: %s",currentName,request.Receiver))
	}

	//// account receiver
	account,err := AccountGetByName(stub,currentName)

	if err != nil {
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}

	if account.Status == false{
		return  common.SendError(common.ACCOUNT_NOT_EXIST,"the receiver is not exist or the receiver is disable")
	}

	//// account sender

	accountFrom,err := AccountGetByName(stub,request.Sender)

	if err != nil {
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}

	if accountFrom.Status == false{
		return  common.SendError(common.ACCOUNT_NOT_EXIST,"the sender is not exist or the sender is disable")
	}


	token , err := TokenGet(stub,request.Token)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	if token.Status== false{
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s token not enable",token.Name))
	}

	txid := stub.GetTxID()
	from := strings.ToUpper( strings.TrimSpace(request.Sender))
	key, err := stub.CreateCompositeKey(common.CompositeRequestIndexName, []string{common.SIGN_PRE,token.Name, from,  txid})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", from, txid, err.Error()))
	}

	sign := model.SignRequest{}
	sign.Token = request.Token
	sign.Sender = request.Sender
	sign.Desc = request.Desc
	sign.Amount = request.Amount
	sign.Receiver = request.Receiver
	sign.TxID = txid
	sign.Status = common.PENDING_SIGN

	signBYte , err := json.Marshal(sign)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	err = stub.PutState(key,signBYte)
	if err != nil {
		return common.SendError(common.PUTSTAT_ERR,err.Error())
	}
	return common.SendScuess(fmt.Sprintf("the %s token from %s to %s had request",request.Token,request.Sender,request.Receiver))
}

func SignGetRequest(stub shim.ChaincodeStubInterface) pb.Response{
	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return common.SendError(common.Param_ERR,"Parameters error ,please check Parameters")
	}
	signGet := args[0]
	log.Logger.Info(signGet)
	//// account sender

	sign := model.LedgerSignGetParam{}
	err := json.Unmarshal([]byte(signGet),&sign)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	account,err := AccountGetByName(stub,sign.Sender)
	if err != nil {
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}
	if account.Status == false{
		return  common.SendError(common.ACCOUNT_NOT_EXIST,"the sender is not exist or the sender is disable")
	}

	token , err := TokenGet(stub,sign.Token)
	if err != nil {
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s Token is disable",token.Name))
	}

	iter, err := stub.GetStateByPartialCompositeKey(common.CompositeRequestIndexName, []string{common.SIGN_PRE,token.Name,account.DidName})
	if err != nil {
		return common.SendError(common.GETSTAT_ERR,err.Error())
	}
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false

	for iter.HasNext(){
		response ,err := iter.Next()

		if err != nil {
			continue
		}

		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		buffer.WriteString(string(response.Value))

		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return common.SendScuess(buffer.String())
}

func SignHistory(stub shim.ChaincodeStubInterface)pb.Response {
	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return common.SendError(common.Param_ERR,"Parameters error ,please check Parameters")
	}
	signRespJson := args[0]
	log.Logger.Info(signRespJson)
	response := model.LedgerResponseParam{}

	err := json.Unmarshal([]byte(signRespJson),&response)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	token , err := TokenGet(stub,response.Token)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	key, err := stub.CreateCompositeKey(common.CompositeRequestIndexName, []string{common.SIGN_PRE,token.Name,strings.ToUpper(strings.TrimSpace(response.Sender)),  response.Txid})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", response.Sender, response.Txid, err.Error()))
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

func SignRepsonse(stub shim.ChaincodeStubInterface)pb.Response  {
	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return  common.SendError(common.Param_ERR,"Parameters error ,please check Parameters")
	}
	signRespJson := args[0]
	log.Logger.Info(signRespJson)
	response := model.LedgerResponseParam{}

	err := json.Unmarshal([]byte(signRespJson),&response)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	currentName ,err := common.GetCommonName(stub)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}

	if strings.ToUpper(strings.TrimSpace(currentName)) != strings.ToUpper(strings.TrimSpace(response.Sender)) {
		return common.SendError(common.ACCOUNT_PREMISSION,fmt.Sprintf("current login user:%s not equal sender: %s",currentName,response.Sender))
	}

	token , err := TokenGet(stub,response.Token)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	if token.Status == false {
		return common.SendError(common.TKNERR_LOCKED,fmt.Sprintf("%s Token is disable",token.Name))
	}

	key, err := stub.CreateCompositeKey(common.CompositeRequestIndexName, []string{common.SIGN_PRE,token.Name,strings.ToUpper(strings.TrimSpace(response.Sender)),  response.Txid})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", response.Sender, response.Txid, err.Error()))
	}

	respBYte,err := stub.GetState(key)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.GETSTAT_ERR,err.Error())
	}

	sign := model.SignRequest{}

	err = json.Unmarshal(respBYte,&sign)

	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	if sign.Status != common.PENDING_SIGN {
		return common.SendError(common.STATUS_ERR,fmt.Sprintf("from : %s can not transfer to user :%s , the request had signed",sign.Sender,sign.Receiver))
	}

	///////////////////////check token and from to
	accountFrom, err := AccountGetByName(stub,sign.Sender)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}
	accountTo, err := AccountGetByName(stub,sign.Receiver)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.ACCOUNT_COMMONNAME,err.Error())
	}

	if accountFrom.DidName == accountTo.DidName {
		return common.SendError(common.ACCOUNT_PREMISSION,fmt.Sprintf("from : %s can not equal to user :%s",sign.Sender,sign.Receiver))
	}

	if accountFrom.Status == false || accountTo.Status == false {
		return common.SendError(common.ACCOUNT_LOCK,fmt.Sprintf("from : %s OR to  :%s is locked",sign.Sender,sign.Receiver))
	}

	///////// composite key
	ledgerkey, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name),  strings.ToUpper(accountFrom.DidName)})
	if err != nil {
		return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, accountFrom.DidName, err.Error()))
	}

	ledgerByte,err := stub.GetState(ledgerkey)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.GETSTAT_ERR,err.Error())
	}

	//////// transfer
	ledger := model.Ledger{}
	err = json.Unmarshal(ledgerByte,&ledger)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.MARSH_ERR,err.Error())
	}

	if response.Accept == true{  ///// 同意支付

		if ledger.Amount < sign.Amount {
			return common.SendError(common.Balance_NOT_ENOUGH,fmt.Sprintf("the %s token balance not enough",token.Name))
		}
		ledger.Amount = ledger.Amount - sign.Amount

		ledger.Desc = fmt.Sprintf("From : %s transfer To : %s , value : %s ",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(sign.Amount,'f',2,64))

		ledgerByted , err := json.Marshal(ledger)
		if err != nil {
			log.Logger.Error(err.Error())
			return common.SendError(common.MARSH_ERR,err.Error())
		}
		err = stub.PutState(ledgerkey,ledgerByted)
		if err != nil {
			log.Logger.Error(err.Error())
			return common.SendError(common.PUTSTAT_ERR,err.Error())
		}
		////// update sign
		sign.Status = common.SENT_SIGN
		sign.Desc = fmt.Sprintf("the signature request had accepted !")
		//////////////////////to
		tokey, err := stub.CreateCompositeKey(common.CompositeIndexName, []string{common.Ledger_PRE, strings.ToUpper(token.Name), strings.ToUpper(accountTo.DidName)})
		if err != nil {
			return common.SendError(common.COMPOSTEKEY_ERR,fmt.Sprintf("Could not create a composite key for %s-%s: %s", token.Name, accountTo.DidName, err.Error()))
		}
		toledgerByte,err := stub.GetState(tokey)
		if err != nil{
			log.Logger.Error("TO GetState:",err)
			return common.SendError(common.GETSTAT_ERR,err.Error())
		}

		toledger := model.Ledger{}

		if toledgerByte == nil {
			toledger.Holder = strings.ToUpper(accountTo.DidName)
			toledger.Token = strings.ToUpper(token.Name)
			toledger.Desc = fmt.Sprintf("From : %s transfer To : %s , value : %s ",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(sign.Amount,'f',2,64))
			toledger.Amount = toledger.Amount + sign.Amount
		}else {
			err = json.Unmarshal(toledgerByte,&toledger)
			if err != nil {
				log.Logger.Error(err.Error())
				return common.SendError(common.MARSH_ERR,err.Error())
			}
			toledger.Desc = fmt.Sprintf("From : %s transfer To : %s , value : %s ",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(sign.Amount,'f',2,64))
			toledger.Amount = toledger.Amount + sign.Amount
		}
		toTransferedByted , err := json.Marshal(toledger)
		if err != nil {
			return common.SendError(common.MARSH_ERR,err.Error())
		}
		err = stub.PutState(tokey,toTransferedByted)
		if err != nil {
			return common.SendError(common.PUTSTAT_ERR,err.Error())
		}

		////////////////////==============================================///////////////////////
		////////////////// send event
		ts, err := stub.GetTxTimestamp()
		if err != nil {
			return common.SendError(common.MARSH_ERR,err.Error())
		}
		//////// set event
		evt := model.LedgerEvent{
			Type: common.Evt_payment,
			Txid:  stub.GetTxID(),
			Time:   ts.GetSeconds(),
			From:   sign.Sender,
			To:     sign.Receiver,
			Amount: strconv.FormatFloat(sign.Amount,'f',2,64) ,
			Token:  sign.Token,
		}

		eventJSONasBytes, err := json.Marshal(evt)
		if err != nil {
			return common.SendError(common.MARSH_ERR,err.Error())
		}

		err = stub.SetEvent(sign.TxID, eventJSONasBytes)
		if err != nil {
			return common.SendError(common.EVENT_ERR,err.Error())
		}

	}else{  /////// 不同意支付
		////// update sign
		sign.Desc = fmt.Sprintf("the signature request had refused!")
		sign.Status = common.Failed_SIGN
	}

	signBYte , err := json.Marshal(sign)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.MARSH_ERR,err.Error())
	}
	err = stub.PutState(key,signBYte)
	if err != nil {
		log.Logger.Error(err.Error())
		return common.SendError(common.PUTSTAT_ERR,err.Error())
	}
	if sign.Status == common.Failed_SIGN {
		return common.SendScuess(fmt.Sprintf("From : %s transfer To : %s , value : %s had refused",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(sign.Amount,'f',2,64)))
	}else{
		return common.SendScuess(fmt.Sprintf("From : %s transfer To : %s , value : %s had accepted",accountFrom.DidName,accountTo.DidName,strconv.FormatFloat(sign.Amount,'f',2,64)))
	}

}