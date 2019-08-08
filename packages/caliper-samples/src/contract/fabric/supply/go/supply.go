/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type product struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Puid       string `json:"puid"`
	Pname      string `json:"pname"` //the fieldtags are needed to keep case from bouncing around
	Ptype      string `json:"ptype"`
	Owner      string `json:"owner"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initProduct" { //create a new product
		return t.initProduct(stub, args)
	} else if function == "transferProduct" { //change owner of a specific product
		return t.transferProduct(stub, args)
	} else if function == "readProduct" { //read a product
		return t.readProduct(stub, args)
	} else if function == "queryProduct" { //find product based on an ad hoc rich query
		return t.queryProduct(stub, args)
	} else if function == "getHistoryForProduct" { //get history of values for a product
		return t.getHistoryForProduct(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initProduct - create a new product, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init product")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}

	productUID := args[0]
	pname := strings.ToLower(args[1])
	ptype := strings.ToLower(args[2])
	owner := strings.ToLower(args[3])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	// ==== Check if product already exists ====
	productAsBytes, err := stub.GetState(productUID)
	if err != nil {
		return shim.Error("Failed to get product: " + err.Error())
	} else if productAsBytes != nil {
		fmt.Println("This product already exists: " + productUID)
		return shim.Error("This product already exists: " + productUID)
	}

	// ==== Create product object and marshal to JSON ====
	objectType := "product"
	product := &product{objectType, productUID, pname, ptype, owner}
	productJSONasBytes, err := json.Marshal(product)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save product to state ===
	err = stub.PutState(productUID, productJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	indexName := "type~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{product.Ptype, product.Pname})
	if err != nil {
		return shim.Error(err.Error())
	}

	value := []byte{0x00}
	stub.PutState(colorNameIndexKey, value)

	fmt.Println("- end init product")
	return shim.Success(nil)
}

func (t *SimpleChaincode) readProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var Puid, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting product ID of the product to query")
	}

	Puid = args[0]
	valAsbytes, err := stub.GetState(Puid) //get the product from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + Puid + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + Puid + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

func (t *SimpleChaincode) transferProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	puid := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start product transfer ", puid, newOwner)

	productAsBytes, err := stub.GetState(puid)
	if err != nil {
		return shim.Error("Failed to get product:" + err.Error())
	} else if productAsBytes == nil {
		return shim.Error("Product does not exist")
	}

	productToTransfer := product{}
	err = json.Unmarshal(productAsBytes, &productToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	productToTransfer.Owner = newOwner //change the owner

	productJSONasBytes, _ := json.Marshal(productToTransfer)
	err = stub.PutState(puid, productJSONasBytes) //rewrite the product
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end of product transfer (success)")
	return shim.Success(nil)
}

//getHistoryForProdcut

func (t *SimpleChaincode) queryProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func (t *SimpleChaincode) getHistoryForProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	Puid := args[0]

	fmt.Printf("- start getHistoryForProduct: %s\n", Puid)

	resultsIterator, err := stub.GetHistoryForKey(Puid)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the product
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON product)
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

	fmt.Printf("- getHistoryForProduct returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
