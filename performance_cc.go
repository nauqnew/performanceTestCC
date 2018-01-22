package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"

	"encoding/hex"
	"github.com/bitly/go-simplejson"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	msp "github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	separatorColon                   = ":"
	separatorUnderscore              = "_"
	separatorComma                   = ","
	businessFlowParameter            = "t"
	businessFlowKey                  = "BUSINESS_FLOW"
	categoryCollectionKey            = "CATEGORY_COLLECTION"
	orgCodeJD                        = "cloudFactory"
	orgCodeJY                        = "jyzb"
	errorCodeIncorrectArgumentNumber = "CF1401"
	errorCodePutStateError           = "CF1402"
	errorCodeGetStateError           = "CF1403"
	errorCodeOperationError          = "CF1404"
	errorCodeNoStageError            = "CF1405"
	errorCodeIncorrectCategory       = "CF1406"
	errorCodeNoWriteAccess           = "CF1407"
	errorCodeIncorrectSequence       = "CF1408"
)

// Stage is used to describe a stage in business flow
type Stage struct {
	Name           string   `json:"name"`           // Stage name
	PreviousStages []string `json:"previousStages"` // Possilbe previous stage collection
	Operator       string   `json:"operator"`       // Legal operator orgCode
	Category       string   `json:"category"`       // Category the stage belongs to
}

// BusinessFlow is used to describe the busienss flow containing all the stages
type BusinessFlow struct {
	Stages     []*Stage `json:"stages"`     // All possible stages
	Categories []string `json:"categories"` // All possible categories
}

// AbsChaincode implements chaincode
type AbsChaincode struct {
}

func (abs *AbsChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response{
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 0 {
		return shim.Error(errorCodeIncorrectArgumentNumber + " Incorrect number of arguments. Expecting 0")
	}

	businessFlowBytes := []byte(businessFlowJSON)

	err := stub.PutState(businessFlowKey, businessFlowBytes)
	if err != nil {
		return shim.Error(errorCodePutStateError + "Put state (business flow) failed. " + err.Error())
	}

	pubKeyBytesJD, err := hex.DecodeString(pubKeyHexJD)
	if err != nil {
		return shim.Error(errorCodeOperationError + "Decode hex string (JD public key) failed. " + err.Error())
	}
	err = stub.PutState(orgCodeJD, pubKeyBytesJD)
	if err != nil {
		return shim.Error(errorCodePutStateError + "Put state (JD public key) failed. " + err.Error())
	}

	pubKeyBytesJY, err := hex.DecodeString(pubKeyHexJY)
	if err != nil {
		return shim.Error(errorCodeOperationError + "Decode hex string (JY public key) failed. " + err.Error())
	}
	err = stub.PutState(orgCodeJY, pubKeyBytesJY)
	if err != nil {
		return shim.Error(errorCodePutStateError + "Put state (JY public key) failed. " + err.Error())
	}

	return shim.Success(nil)

}

type BizContent struct{
	assets []AssetDetails `json:"assets"`
}

type AssetDetails struct{
	assetUid string `json:"assetUid"`
	assetDetails string `json:"assetDetails"`
}

func (abs *AbsChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response{
	fmt.Println("Invoke start!")
	// 1. get args
	function, args := stub.GetFunctionAndParameters()
	if function == "query" {
		return abs.query(stub, args)
	}
	if len(args) != 7 {
		return shim.Error(errorCodeIncorrectArgumentNumber + " Incorrect number of arguments. Expecting 7")
	}
	org := args[0]
	//category := args[3]
	bizContent := args[6]

	fmt.Println("parameters ok !")
	fmt.Println("bizContent " + bizContent)


	// 2. check write access
	//err := checkValidity(stub, function, org, category)
	//if err != nil {
	//	return shim.Error(err.Error())
	//}

	// TODO  3. extract assetDetails from bizContent
	//keyToPut := assetUID + separatorUnderscore + category

	bizContentJson,_ := simplejson.NewJson([]byte(bizContent))
	assets, _ := bizContentJson.Get("assets").Array()
	for i,_ := range assets {
		asset := bizContentJson.Get("assets").GetIndex(i)
		assetUid := asset.Get("assetUid")
		assetDetail := asset.Get("assetDetails")
		keyToPut, _ := assetUid.String()
		valueToput, _ := assetDetail.String()
		valueToPutBytes := []byte(valueToput)
		err := stub.PutState(keyToPut, valueToPutBytes)
		if err != nil {
			return shim.Error(errorCodePutStateError + " Put state (current stage and txID) failed. " + err.Error())
		}
		fmt.Println(keyToPut + " : " + valueToput)
	}
	fmt.Println("Invoke chaincode succeed. " + "Type: " + function + ". Operator: " + org)

	return shim.Success(nil)
}

func (abs *AbsChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var worldStateKey string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 to query")
	}

	worldStateKey = args[0]

	worldStateValueBytes, err := stub.GetState(worldStateKey)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + worldStateKey + "\"}"
		return shim.Error(jsonResp)
	}

	if worldStateValueBytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + worldStateKey + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Key\":\"" + worldStateKey + "\",\"Value\":\"" + string(worldStateValueBytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success([]byte(jsonResp))
}

// NewBusinessFlow new a BusinessFlow struct given a JSON string
func NewBusinessFlow(businessFlowString string) (*BusinessFlow, error) {
	businessFlow := BusinessFlow{}
	err := json.Unmarshal([]byte(businessFlowString), &businessFlow)
	if err != nil {
		return nil, err
	}
	return &businessFlow, nil
}

func (b *BusinessFlow) GetStageDefinitionByName(stageName string) (*Stage, error) {
	for _, stage := range b.Stages {
		if stageName == stage.Name {
			return stage, nil
		}
	}
	return nil, nil
}

func getPubKeyFromIdentity(identityBytes []byte) (string, error) {
	identity := &msp.SerializedIdentity{}
	err := proto.Unmarshal(identityBytes, identity)
	if err != nil {
		return "", err
	}
	bl, _ := pem.Decode(identity.IdBytes)
	if bl == nil {
		return "", fmt.Errorf("Could not decode the PEM structure")
	}
	cert, err := x509.ParseCertificate(bl.Bytes)
	if err != nil {
		return "", fmt.Errorf("ParseCertificate failed %s", err)
	}
	pubKey := cert.PublicKey.(*ecdsa.PublicKey)
	pubKeyRaw, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", fmt.Errorf("Marshal public key failed [%s]", err)
	}
	return hex.EncodeToString(pubKeyRaw), nil
}

// 校验有效性，权限校验
func checkValidity(stub shim.ChaincodeStubInterface, stageToPut, org string) error {
	// 1. get business flow
	businessFlowBytes, err := stub.GetState(businessFlowKey)
	if err != nil {
		return errors.New(errorCodeGetStateError + " Get state (business flow) failed. " + err.Error())
	} else if businessFlowBytes == nil {
		return errors.New(errorCodeGetStateError + " No business flow found")
	}
	businessFlow, err := NewBusinessFlow(string(businessFlowBytes))
	if err != nil {
		return errors.New(errorCodeOperationError + err.Error())
	}

	// 2. get stage instance
	stageInstance, err := businessFlow.GetStageDefinitionByName(stageToPut)
	if err != nil {
		return errors.New(errorCodeOperationError + err.Error())
	} else if stageInstance == nil {
		return errors.New(errorCodeNoStageError + " No stage instance found for " + stageToPut)
	}

	// 3. check category
	//if stageInstance.Category != category {
	//	return errors.New(errorCodeIncorrectCategory + " Incorrect category received")
	//}

	// 4. check write access
	if stageInstance.Operator != org {
		fmt.Println("Wanted " + stageInstance.Operator + ", but got " + org)
		return errors.New(errorCodeNoWriteAccess + " Check write access failed")
	}
	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return errors.New(errorCodeOperationError + " Get creator failed. " + err.Error())
	}
	creatorPubKey, err := getPubKeyFromIdentity(creatorBytes)
	if err != nil {
		return errors.New(errorCodeOperationError + " Get creator public key failed. " + err.Error())
	}
	operatorPubKeyBytes, err := stub.GetState(org)
	if err != nil {
		return errors.New(errorCodeGetStateError + " Get state (public key) failed. " + err.Error())
	}
	if creatorPubKey != hex.EncodeToString(operatorPubKeyBytes) {
		// fmt.Println("Wanted " + hex.EncodeToString(operatorPubKeyBytes) + ", but got " + creatorPubKey)
		return errors.New(errorCodeNoWriteAccess + " Check write access failed")
	}

	fmt.Println("access check passed. " + stageToPut + org  )
	return nil

}

func main() {
	err := shim.Start(new(AbsChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}

}

const pubKeyHexJD = "3059301306072a8648ce3d020106082a8648ce3d0301070342000469bed65ebfb770313329e770b97d6d298cce48e16da3db24423cdea84d73eb4b300f564efad0137052e7bf89665b5c3397d4f448c8cf711d5b227d40b419c6fb"
const pubKeyHexJY = "3059301306072a8648ce3d020106082a8648ce3d03010703420004c282028c74a018889a0223e8f81a1b91d8337ee8a9ceedcf0c8f586cee9bd5e56eeee7e553ab97639f0860a702e8bee4b534abb673631dfa09ff6d77082f3bf6"

const businessFlowJSON = `
{
	"stages":
	[
		{
			"name": "ASSET_UPLOAD",
			"previousStages": [],
			"operator": "jyzb",
			"category": "CAT_ASSET_UPLOAD"
		}

	],

	"categories": ["CAT_ASSET_UPLOAD"]
}
`