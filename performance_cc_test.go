package main

import (
	"encoding/hex"
	"fmt"
	"testing"

	//"github.com/bitly/go-simplejson"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.FailNow()
	}
}

func checkState(t *testing.T, stub *shim.MockStub, name string, value string) {
	bytes := stub.State[name]
	if bytes == nil {
		fmt.Println("State", name, "failed to get value")
		t.FailNow()
	}
	if string(bytes) != value && hex.EncodeToString(bytes) != value {
		fmt.Println("State value", name, "is", string(bytes), "not", value, "as expected")
		t.FailNow()
	}

}

func checkQuery(t *testing.T, stub *shim.MockStub, name string, value string) {
	res := stub.MockInvoke("1", [][]byte{[]byte("query"), []byte(name)})
	if res.Status != shim.OK {
		fmt.Println("Query", name, "failed", string(res.Message))
		t.FailNow()
	}
	if res.Payload == nil {
		fmt.Println("Query", name, "failed to get value")
		t.FailNow()
	}
	if string(res.Payload) != value {
		fmt.Println("Query value", name, "was not", value, "as expected")
		t.FailNow()
	}
}

func checkInvokeAssetUpload(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args) //"1" is the txid
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.FailNow()
	}
	keyToCheck := "160815609421112015"
	valueToCheck := "160815609421112015,***n66_m,2016-08-15 20:41:30,2017-08-15 23:59:59,12,2399,199.92,0,24,0,3,3,1,1,HT201606300001"
	valueToCheckBytes := []byte(valueToCheck)
	checkState(t, stub, keyToCheck, string(valueToCheckBytes))
}

func TestAbs_Init(t *testing.T) {
	absChaincode := new(AbsChaincode)
	stub := shim.NewMockStub("abs", absChaincode)

	checkInit(t, stub, [][]byte{[]byte("init")})
	checkState(t, stub, businessFlowKey, businessFlowJSON)
	checkState(t, stub, orgCodeJD, pubKeyHexJD)
	checkState(t, stub, orgCodeJY, pubKeyHexJY)
}

func TestAbs_Invoke(t *testing.T) {

	absChaincode := new(AbsChaincode)
	stub := shim.NewMockStub("abs", absChaincode)

	//init must be invoked first
	checkInit(t, stub, [][]byte{[]byte("init")})

	//1. invoke ASSET_UPLOAD {function, orgCode, assetUID, outTradeNo, category, previousTxID, businesshash, bizcontent}
	//bizContent := "{\"assetDetails\":\"160815609421112015,***n66_m,2016-08-15 20:41:30,2017-08-15 23:59:59,12,2399,199.92,0,24,0,3,3,1,1,HT201606300001\"}"

	bizContent := `{
        "assets":[
            {
                "assetUid":"160815609421112015",
                "assetDetails":"160815609421112015,***n66_m,2016-08-15 20:41:30,2017-08-15 23:59:59,12,2399,199.92,0,24,0,3,3,1,1,HT201606300001"
            },
            {
                "assetUid":"160815609421112017",
                "assetDetails":"160815609421112017,***n66_m,2016-08-15 20:41:30,2017-08-15 23:59:59,12,2399,199.92,0,24,0,3,3,1,1,HT201606300001"
            }
        ]
    }`
	assetUploadArgs := [][]byte{[]byte("ASSET_UPLOAD"), []byte("jyzb"), []byte("160815609421112015"), []byte("jyzb0001"), []byte("CAT_ASSET_UPLOAD"), []byte(""), []byte(""), []byte(bizContent)}
	checkInvokeAssetUpload(t, stub, assetUploadArgs)

}

/*func TestStringJson(t *testing.T){
	bizContent := `{
        "assets":[
            {
                "assetUid":"160815609421112015",
                "assetDetails":"160815609421112015,***n66_m,2016-08-15 20:41:30,2017-08-15 23:59:59,12,2399,199.92,0,24,0,3,3,1,1,HT201606300001"
            },
            {
                "assetUid":"160815609421112017",
                "assetDetails":"160815609421112017,***n66_m,2016-08-15 20:41:30,2017-08-15 23:59:59,12,2399,199.92,0,24,0,3,3,1,1,HT201606300001"
            }
        ]
    }`
	bizContentJson,_ := simplejson.NewJson([]byte(bizContent))
	assets, _ := bizContentJson.Get("assets").Array()
	for i,_ := range assets {
		asset := bizContentJson.Get("assets").GetIndex(i)
		assetUid := asset.Get("assetUid")
		assetDetail := asset.Get("assetDetails")
		fmt.Println(assetUid.String())
		fmt.Println(assetDetail.String())
	}


	}*/

 func TestAbs_Query(t *testing.T){
	 absChaincode := new(AbsChaincode)
	 stub := shim.NewMockStub("abs", absChaincode)

	 //init must be invoked first
	 checkInit(t, stub, [][]byte{[]byte("init")})
	 //args:=[][]byte{[]byte("QUERY"), []byte("160815609421112015_CAT_ASSET_UPLOAD")}
	 //checkQuery(t, stub, "160815609421112015_CAT_ASSET_UPLOAD", "")

	 bizContent := "{\"assetDetails\":\"160815609421112015,***n66_m,2016-08-15 20:41:30,2017-08-15 23:59:59,12,2399,199.92,0,24,0,3,3,1,1,HT201606300001\"}"
	 assetUploadArgs := [][]byte{[]byte("ASSET_UPLOAD"), []byte("jyzb"), []byte("160815609421112015"), []byte("jyzb0001"), []byte("CAT_ASSET_UPLOAD"), []byte(""), []byte(""), []byte(bizContent)}
	 checkInvokeAssetUpload(t, stub, assetUploadArgs)

	response := "{\"Key\":\"160815609421112015_CAT_ASSET_UPLOAD\",\"Value\":\"160815609421112015,***n66_m,2016-08-15 20:41:30,2017-08-15 23:59:59,12,2399,199.92,0,24,0,3,3,1,1,HT201606300001\"}"
	 checkQuery(t, stub, "160815609421112015_CAT_ASSET_UPLOAD", response)
 }
