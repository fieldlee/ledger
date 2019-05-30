package common

import (
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"ledger/log"
	"strings"
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
	log.Logger.Info(id)
	if err != nil {
		log.Logger.Error("shim getMsp error", err.Error())
	}
	mspid, err := id.GetMSPID()
	if err != nil {
		log.Logger.Error("shim getMsp error", err.Error())
	}
	log.Logger.Info(mspid)
	return mspid
}

func GetRight(stub shim.ChaincodeStubInterface)(string){
	id, err := cid.New(stub)
	if err != nil {
		log.Logger.Error("shim GetRight error", err.Error())
	}

	cert, err := id.GetX509Certificate()
	if err != nil {
		log.Logger.Error("shim GetRight error", err.Error())
	}

	//id.GetAttributeValue()

	log.Logger.Info(id)
	if cert.IsCA {
		return "Admin"
	}else{
		return "Member"
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
		log.Logger.Error(err)
		return "",err
	}
	certfiaction,err := cert.GetX509Certificate()
	if err != nil {
		log.Logger.Error(err)
		return "",err
	}
	return certfiaction.Subject.CommonName,nil
}

func GetIsAdmin(stub shim.ChaincodeStubInterface)( bool, error){

	name,err := GetCommonName(stub)


	if err != nil {
		log.Logger.Error("current username err :", err.Error())
		return false,err
	}

	log.Logger.Error("current username :", name)

	if strings.ToUpper(strings.TrimSpace(name)) == strings.ToUpper(strings.TrimSpace(ADMIN_Name)){
		return true , nil
	}

	return false,nil
	//val, ok, err := cid.GetAttributeValue(stub, "Admin")
	//if err != nil {
	//	return false,err
	//}
	//
	//if ok {
	//	if val == "true" {
	//		return true,nil
	//	}
	//}
	//return false, nil
}

func IsSuperAdmin(stub shim.ChaincodeStubInterface)(bool){

	//orgid := GetMsp(stub)
	//
	//isAdmin, err:= GetIsAdmin(stub)
	//
	//if err != nil {
	//	log.Logger.Error("IsSuperAdmin error", err.Error())
	//}
	//comName , err := GetCommonName(stub)
	//if err != nil {
	//	log.Logger.Error("GetCommonName error", err.Error())
	//}
	//if strings.ToLower(orgid) == strings.ToLower(ADMIN_ORG) {
	//	if isAdmin == true {
	//		return  true
	//	}
	//	if strings.ToLower(comName) == strings.ToLower(ADMIN_Name) {
	//		return true
	//	}
	//}
	//return false

	name,err := GetCommonName(stub)
	if err != nil {
		log.Logger.Error("IsSuperAdmin error", err.Error())
		return false
	}

	if strings.ToUpper(strings.TrimSpace(name)) == strings.ToUpper(strings.TrimSpace(ADMIN_Name)){
		return true
	}

	return false
}
