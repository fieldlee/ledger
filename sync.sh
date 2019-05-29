
for i in github.com/golang/protobuf/proto github.com/hyperledger/fabric/core/chaincode/lib/cid github.com/hyperledger/fabric/common/attrmgr github.com/hyperledger/fabric/protos/msp github.com/hyperledger/fabric/vendor/github.com/pkg/errors
do
  mkdir -p vendor/$i/
  rsync -avzP --delete $GOPATH/src/$i/ vendor/$i/
done