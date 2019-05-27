package common

import (
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"ledger/log"
)


/** 获取交易发起方的MSPID **/
func GetMspid(stub shim.ChaincodeStubInterface) (string) {
	createrbyte, err := stub.GetCreator() //获得创建者
	if err != nil {
		log.Logger.Error("shim GetCreater error", err.Error())
		return ""
	}
	//解析MSPID
	newbytes := []byte{}
	headFlg := true
	for i := 0; i < len(createrbyte); i++ {
		if createrbyte[i] >= 33 && createrbyte[i] <= 126 {
			headFlg = false
			newbytes = append(newbytes, createrbyte[i])
		}
		if createrbyte[i] < 33 || createrbyte[i] > 126 {
			if !headFlg {
				break
			}
		}
	}
	return string(newbytes)
}


func GetMsp(stub shim.ChaincodeStubInterface)(string){
	id, err := cid.New(stub)
	if err != nil {
		log.Logger.Error("shim getMsp error", err.Error())
	}
	mspid, err := id.GetMSPID()
	if err != nil {
		log.Logger.Error("shim getMsp error", err.Error())
	}
	return mspid
}

func GetRight(stub shim.ChaincodeStubInterface){
	id, err := cid.New(stub)
	if err != nil {
		log.Logger.Error("shim GetRight error", err.Error())
	}

	cert, err := id.GetX509Certificate()
	if err != nil {
		log.Logger.Error("shim GetRight error", err.Error())
	}


}

func SendError(errno int32, msg string) pb.Response {
	return pb.Response{
		Status:  errno,
		Message: msg,
	}
}

func GetCommonName(stub shim.ChaincodeStubInterface)( string, error){
	cert,err := cid.New(stub)
	if err != nil {
		return "",err
	}
	certfiaction,err := cert.GetX509Certificate()
	if err != nil {
		return "",err
	}
	return certfiaction.Subject.CommonName,nil
}

func GetIsAdmin(stub shim.ChaincodeStubInterface)( bool, error){
	cert,err := cid.New(stub)
	if err != nil {
		return false,err
	}
	certfiaction,err := cert.GetX509Certificate()
	if err != nil {
		return false,err
	}
	if certfiaction.Subject.CommonName == ADMIN_Name {
		return true, nil
	}else{
		return false, nil
	}
}
