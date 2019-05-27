package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"ledger/model"
	"ledger/common"
	"strconv"
	"strings"
	"time"
)
/// check 用户名是否可用
func AccountCheck(stub shim.ChaincodeStubInterface)pb.Response{
	_,args := stub.GetFunctionAndParameters()
	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	accountName := strings.ToUpper(strings.TrimSpace(args[0]))

	if accountName == ""{
		return shim.Error("The Name is Blank")
	}

	if len(accountName) < 3 && len(accountName) > 64 {
		return shim.Error("The Name must low 64 strings and higher 3 strings")
	}

	accByte,err := stub.GetState(common.ACCOUNT_PRE + accountName)
	if err != nil {
		return shim.Error(err.Error())
	}
	// 查询account 是否存在
	if accByte == nil {
		// 添加用户信息
		newAccount := model.Account{}
		newAccount.DidName = accountName

		newAccByte,err := json.Marshal(newAccount)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = stub.PutState(common.ACCOUNT_PRE + accountName,newAccByte)
		if err != nil {
			return shim.Error(err.Error())
		}
		// 返回check 状态
		return shim.Success(nil)
	}

	return common.SendError(common.ACCOUNT_EXIST,fmt.Sprintf("%s is exist",args[0]))
}
/// admin or 用户 confirm 用户并制为有效
func AccountConfirm(stub shim.ChaincodeStubInterface)pb.Response{

	isAdmin,err := common.GetIsAdmin(stub)

	if err != nil {
		return shim.Error(err.Error())
	}

	if isAdmin {

		_,args := stub.GetFunctionAndParameters()

		if len(args) != 1{
			return shim.Error("Parameters error ,please check Parameters")
		}

		accountName := strings.ToUpper(strings.TrimSpace(args[0]))

		if accountName == ""{
			return shim.Error("The Name is Blank")
		}
		if len(accountName) < 3 && len(accountName) > 64 {
			return shim.Error("The Name must low 64 strings and higher 3 strings")
		}

		accountByte,err := stub.GetState(common.ACCOUNT_PRE + accountName)

		if err != nil {
			return shim.Error(err.Error())
		}

		account := model.Account{}
		if accountByte != nil {
			err = json.Unmarshal(accountByte,&account)
			if err != nil {
				return shim.Error(err.Error())
			}
		}
		account.DidName = accountName
		account.CommonName = strings.TrimSpace(args[0])
		account.MspID = common.GetMspid(stub)
		account.ObjectType = common.ACCOUNT
		account.Status = true

		newaccountByte,err := json.Marshal(account)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = stub.PutState(common.ACCOUNT_PRE + accountName,newaccountByte)
		if err != nil {
			return shim.Error(err.Error())
		}

		return shim.Success(nil)

	} else { //////  普通用户

		commonName,err := common.GetCommonName(stub)

		if err != nil {
			return shim.Error(err.Error())
		}

		accountName := strings.ToUpper(strings.TrimSpace(commonName))

		accountByte,err := stub.GetState(common.ACCOUNT_PRE + accountName)

		if err != nil {
			return shim.Error(err.Error())
		}

		account := model.Account{}
		if accountByte == nil {
			return common.SendError(common.ACCOUNT_NOT_EXIST,"the common name not check, please first call check api")
		}else{
			err = json.Unmarshal(accountByte,&account)
			if err != nil {
				return shim.Error(err.Error())
			}
			account.DidName = accountName
			account.CommonName = commonName
			account.MspID = common.GetMspid(stub)
			account.ObjectType = common.ACCOUNT
			account.Status = true

			return shim.Success(nil)
		}
	}
}
// 查找用户
func AccountGetByName(stub shim.ChaincodeStubInterface,didName string)(model.Account,error){
	accountName := strings.ToUpper(strings.TrimSpace(didName))

	didByte,err := stub.GetState(common.ACCOUNT_PRE + accountName)
	if err != nil {
		return model.Account{},err
	}
	if didByte == nil {
		return model.Account{},nil
	}

	didAccount := model.Account{}
	err = json.Unmarshal(didByte,&didAccount)
	if err != nil {
		return model.Account{},err
	}
	return didAccount,nil
}
/// 锁定账号
func AccountLock(stub shim.ChaincodeStubInterface)pb.Response{

	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	isAdmin , err := common.GetIsAdmin(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	if isAdmin == false {
		return shim.Error("only admin can call this function")
	}
	// 超级管理员账号
	account := model.Account{}

	accountName := strings.ToUpper(strings.TrimSpace(args[0]))
	byteAccount,err := stub.GetState(common.ACCOUNT_PRE + accountName)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = json.Unmarshal(byteAccount,&account)
	if err != nil {
		return shim.Error(err.Error())
	}

	if common.GetMspid(stub) == common.ADMIN_ORG {

		account.Status = false

		accountByte,err := json.Marshal(account)
		if err != nil{
			return shim.Error(err.Error())
		}
		err = stub.PutState(common.ACCOUNT_PRE + accountName,accountByte)
		if err != nil{
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}else{
		//  比较mspid
		if common.GetMspid(stub) == account.MspID {
			account.Status = false

			accountByte,err := json.Marshal(account)
			if err != nil{
				return shim.Error(err.Error())
			}
			err = stub.PutState(common.ACCOUNT_PRE + accountName,accountByte)
			if err != nil{
				return shim.Error(err.Error())
			}
			return shim.Success(nil)
		}else{
			return shim.Error(fmt.Sprintf("only %s admin or super admin can call this function",account.MspID))
		}
	}
}
/// 解锁账号
func AccountUNLock(stub shim.ChaincodeStubInterface)pb.Response{

	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	isAdmin , err := common.GetIsAdmin(stub)
	if err != nil {
		return shim.Error(err.Error())
	}

	if isAdmin == false {
		return shim.Error("only admin can call this function")
	}
	// 超级管理员账号
	account := model.Account{}

	accountName := strings.ToUpper(strings.TrimSpace(args[0]))
	byteAccount,err := stub.GetState(common.ACCOUNT_PRE + accountName)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = json.Unmarshal(byteAccount,&account)
	if err != nil {
		return shim.Error(err.Error())
	}

	if common.GetMspid(stub) == common.ADMIN_ORG {

		account.Status = true

		accountByte,err := json.Marshal(account)
		if err != nil{
			return shim.Error(err.Error())
		}
		err = stub.PutState(common.ACCOUNT_PRE + accountName,accountByte)
		if err != nil{
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}else{
		//  比较mspid
		if common.GetMspid(stub) == account.MspID {
			account.Status = true
			accountByte,err := json.Marshal(account)
			if err != nil{
				return shim.Error(err.Error())
			}
			err = stub.PutState(common.ACCOUNT_PRE + accountName,accountByte)
			if err != nil{
				return shim.Error(err.Error())
			}
			return shim.Success(nil)
		}else{
			return shim.Error(fmt.Sprintf("only %s admin or super admin can call this function",account.MspID))
		}
	}
}
/// get history
func AccountGetHistory(stub shim.ChaincodeStubInterface)pb.Response{
	_,args := stub.GetFunctionAndParameters()

	if len(args) != 1{
		return shim.Error("Parameters error ,please check Parameters")
	}

	accountName := strings.ToUpper(strings.TrimSpace(args[0]))

	history, err := stub.GetHistoryForKey(common.ACCOUNT_PRE + accountName)

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