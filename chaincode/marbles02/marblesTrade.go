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

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke marbles ====
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble1","blue","35","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble2","red","50","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble3","blue","70","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarble","marble2","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarblesBasedOnColor","blue","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["delete","marble1"]}'

// ==== Query marbles ====
// peer chaincode query -C myc1 -n marbles -c '{"Args":["readMarble","marble1"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getMarblesByRange","marble1","marble3"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getHistoryForMarble","marble1"]}'

// Rich Query (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarblesByOwner","tom"]}'
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"owner\":\"tom\"}}"]}'

//The following examples demonstrate creating indexes on CouchDB
//Example hostname:port configurations
//
//Docker or vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for chaincodeid, docType, owner.
// Note that docType and owner fields must be prefixed with the "data" wrapper
// chaincodeid must be added for all queries
//
// Definition for use with Fauxton interface
// {"index":{"fields":["chaincodeid","data.docType","data.owner"]},"ddoc":"indexOwnerDoc", "name":"indexOwner","type":"json"}
//
// example curl definition for use with command line
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"chaincodeid\",\"data.docType\",\"data.owner\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1/_index
//

// Index for chaincodeid, docType, owner, size (descending order).
// Note that docType, owner and size fields must be prefixed with the "data" wrapper
// chaincodeid must be added for all queries
//
// Definition for use with Fauxton interface
// {"index":{"fields":[{"data.size":"desc"},{"chaincodeid":"desc"},{"data.docType":"desc"},{"data.owner":"desc"}]},"ddoc":"indexSizeSortDoc", "name":"indexSizeSortDesc","type":"json"}
//
// example curl definition for use with command line
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"data.size\":\"desc\"},{\"chaincodeid\":\"desc\"},{\"data.docType\":\"desc\"},{\"data.owner\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1/_index

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":\"marble\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":{\"$eq\":\"marble\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"reflect"
	
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type marble struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Color      string `json:"color"`
	Size       int    `json:"size"`
	Owner      string `json:"owner"`
}

type Description struct{
	Color string `json:"color"`
	Size int `json:"size"`
}

type AnOpenTrade struct{
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	User string `json:"user"`					//user who created the open trade order
	Timestamp int64 `json:"timestamp"`	
	Want Description  `json:"want"`				//description of desired marble
	Willing Description `json:"willing"`		//marbles willing to trade away
}

type AllOpenTrades struct{
	OpenTrades []AnOpenTrade `json:"open_trades"`
}

var openTradesStr = "_opentrades"				//name for the key/value that will store all open trades

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
	if function == "initMarble" { //create a new marble
		return t.initMarble(stub, args)
	} else if function == "transferMarble" { //change owner of a specific marble
		return t.transferMarble(stub, args)
	} else if function == "transferMarblesBasedOnColor" { //transfer all marbles of a certain color
		return t.transferMarblesBasedOnColor(stub, args)
	} else if function == "delete" { //delete a marble
		return t.delete(stub, args)
	} else if function == "readMarble" { //read a marble
		return t.readMarble(stub, args)
	} else if function == "queryMarblesByOwner" { //find marbles for owner X using rich query
		return t.queryMarblesByOwner(stub, args)
	} else if function == "queryMarbles" { //find marbles based on an ad hoc rich query
		return t.queryMarbles(stub, args)
	} else if function == "getHistoryForMarble" { //get history of values for a marble
		return t.getHistoryForMarble(stub, args)
	} else if function == "getMarblesByRange" { //get marbles based on range query
		return t.getMarblesByRange(stub, args)
	} else if function == "openTrade" { //open a new marble trade
		return t.openTrade(stub, args)
	} else if function == "initOpenTrade" { //open a new marble trade
		return t.initOpenTrade(stub, args)
	} else if function == "getOpenTradesByRange" { //open a new marble trade
		return t.getOpenTradesByRange(stub, args)
	} else if function == "readOpenTrade" { //read marble trades
		return t.readOpenTrade(stub, args)
	} else if function == "removeOpenTrade" { //remove marble trade
		return t.removeOpenTrade(stub, args)
	} else if function == "swapMarble" { // swap two marbles between owner
		return t.swapMarble(stub, args)
	} else if function == "swapMarbleTri" { // swap two marbles between owner
		return t.swapMarbleTri(stub, args)
	} else if function == "matchTrade" { // match the open trades
		return t.matchTrade(stub, args)
	} else if function == "matchTrade2" { // match the open trades
		return t.matchTrade2(stub, args)
	} else if function == "matchTriTrade" { // match the open trades
		return t.matchTriTrade(stub, args)
	} else if function == "clearOpenTrades" { // match the open trades
		return t.clearOpenTrades(stub, args)
	}
	
	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initMarble - create a new marble, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("testing init Marble2")
	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init marble")
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
	marbleName := args[0]
	color := strings.ToLower(args[1])
	owner := strings.ToLower(args[3])
	size, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	// ==== Check if marble already exists ====
	marbleAsBytes, err := stub.GetState(marbleName)
	if err != nil {
		return shim.Error("Failed to get marble: " + err.Error())
	} else if marbleAsBytes != nil {
		fmt.Println("This marble already exists: " + marbleName)
		return shim.Error("This marble already exists: " + marbleName)
	}

	// ==== Create marble object and marshal to JSON ====
	objectType := "marble"
	marble := &marble{objectType, marbleName, color, size, owner}
	marbleJSONasBytes, err := json.Marshal(marble)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the marble json string manually if you don't want to use struct marshalling
	//marbleJSONasString := `{"docType":"Marble",  "name": "` + marbleName + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//marbleJSONasBytes := []byte(str)

	// === Save marble to state ===
	err = stub.PutState(marbleName, marbleJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the marble to enable color-based range queries, e.g. return all blue marbles ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marble.Color, marble.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(colorNameIndexKey, value)

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("- end init marble")
	return shim.Success(nil)
}

// ===============================================
// readMarble - read a marble from chaincode state
// ===============================================
func (t *SimpleChaincode) readMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ==================================================
// delete - remove a marble key/value pair from state
// ==================================================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var marbleJSON marble
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	marbleName := args[0]

	// to maintain the color~name index, we need to read the marble first and get its color
	valAsbytes, err := stub.GetState(marbleName) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + marbleName + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + marbleName + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &marbleJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + marbleName + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(marbleName) //remove the marble from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// maintain the index
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marbleJSON.Color, marbleJSON.Name})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(colorNameIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}

// ===========================================================
// transfer a marble by setting a new owner name on the marble
// ===========================================================
func (t *SimpleChaincode) transferMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	marbleName := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start transferMarble ", marbleName, newOwner)

	marbleAsBytes, err := stub.GetState(marbleName)
	if err != nil {
		return shim.Error("Failed to get marble:" + err.Error())
	} else if marbleAsBytes == nil {
		return shim.Error("Marble does not exist")
	}

	marbleToTransfer := marble{}
	err = json.Unmarshal(marbleAsBytes, &marbleToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	marbleToTransfer.Owner = newOwner //change the owner

	marbleJSONasBytes, _ := json.Marshal(marbleToTransfer)
	err = stub.PutState(marbleName, marbleJSONasBytes) //rewrite the marble
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferMarble (success)")
	return shim.Success(nil)
}

// ===========================================================================================
// getMarblesByRange performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) getMarblesByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
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

	fmt.Printf("- getMarblesByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ==== Example: GetStateByPartialCompositeKey/RangeQuery =========================================
// transferMarblesBasedOnColor will transfer marbles of a given color to a certain new owner.
// Uses a GetStateByPartialCompositeKey (range query) against color~name 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *SimpleChaincode) transferMarblesBasedOnColor(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "color", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	color := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start transferMarblesBasedOnColor ", color, newOwner)

	// Query the color~name index by color
	// This will execute a key range query on all keys starting with 'color'
	coloredMarbleResultsIterator, err := stub.GetStateByPartialCompositeKey("color~name", []string{color})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer coloredMarbleResultsIterator.Close()

	// Iterate through result set and for each marble found, transfer to newOwner
	var i int
	for i = 0; coloredMarbleResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the marble name from the composite key
		responseRange, err := coloredMarbleResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the color and name from color~name composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedColor := compositeKeyParts[0]
		returnedMarbleName := compositeKeyParts[1]
		fmt.Printf("- found a marble from index:%s color:%s name:%s\n", objectType, returnedColor, returnedMarbleName)

		// Now call the transfer function for the found marble.
		// Re-use the same function that is used to transfer individual marbles
		response := t.transferMarble(stub, []string{returnedMarbleName, newOwner})
		// if the transfer failed break out of loop and return error
		if response.Status != shim.OK {
			return shim.Error("Transfer failed: " + response.Message)
		}
	}

	responsePayload := fmt.Sprintf("Transferred %d %s marbles to %s", i, color, newOwner)
	fmt.Println("- end transferMarblesBasedOnColor: " + responsePayload)
	return shim.Success([]byte(responsePayload))
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// queryMarblesByOwner queries for marbles based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryMarblesByOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	owner := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"marble\",\"owner\":\"%s\"}}", owner)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===== Example: Ad hoc rich query ========================================================
// queryMarbles uses a query string to perform a query for marbles.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryMarblesForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *SimpleChaincode) queryMarbles(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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

func (t *SimpleChaincode) getHistoryForMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	marbleName := args[0]

	fmt.Printf("- start getHistoryForMarble: %s\n", marbleName)

	resultsIterator, err := stub.GetHistoryForKey(marbleName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
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
		//as-is (as the Value itself a JSON marble)
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

	fmt.Printf("- getHistoryForMarble returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

//=================================================================================
// marbles trading
//=================================================================================


// ============================================================
// initOpenTrade - create a new marble, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initOpenTrade(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
	if len(args) < 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	
	size1, err := strconv.Atoi(args[2])
	size2, err := strconv.Atoi(args[4])
	
	open := AnOpenTrade{}
	open.ObjectType = "openTrade"
	open.Timestamp = makeTimestamp()
	open.User = args[0]
	open.Want.Color = args[1]
	open.Want.Size =  size1
	open.Willing.Color = args[3]
	open.Willing.Size =  size2

	openTradeKey := "openTrade" + strconv.FormatInt(open.Timestamp, 10)

	// ==== Check if opentrade already exists ====
	openTradeAsBytes, err := stub.GetState(openTradeKey)
	if err != nil {
		return shim.Error("Failed to get opentrade: " + err.Error())
	} else if openTradeAsBytes != nil {
		fmt.Println("This opentrade already exists: " )
		return shim.Error("This opentrade already exists: " )
	}

	// ==== Create AnOpenTrade object and marshal to JSON ====

	openTradeJSONasBytes, err := json.Marshal(open)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save marble to state ===
	err = stub.PutState(openTradeKey, openTradeJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("- end init openTrade: " + strconv.FormatInt(open.Timestamp, 10))
	return shim.Success(nil)
}


func (t *SimpleChaincode) getOpenTradesByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
		if len(args) < 2 {
			return shim.Error("Incorrect number of arguments. Expecting 2")
		}
	
		startKey := args[0]
		endKey := args[1]
	
		resultsIterator, err := stub.GetStateByRange(startKey, endKey)
		if err != nil {
			return shim.Error(err.Error())
		}
		defer resultsIterator.Close()
	
		// buffer is a JSON array containing QueryResults
		var buffer bytes.Buffer
		buffer.WriteString("[")
	
		bArrayMemberAlreadyWritten := false
		for resultsIterator.HasNext() {
			queryResponse, err := resultsIterator.Next()
			if err != nil {
				return shim.Error(err.Error())
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
	
		fmt.Printf("- getOpenTradesByRange queryResult:\n%s\n", buffer.String())
	
		return shim.Success(buffer.Bytes())
	}

func (t *SimpleChaincode) openTrade(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
		if len(args) < 5 {
			return shim.Error("Incorrect number of arguments. Expecting 5")
		}
		
		size1, err := strconv.Atoi(args[2])
		size2, err := strconv.Atoi(args[4])
		
		open := AnOpenTrade{}
		open.ObjectType = "openTrade"
		open.Timestamp = makeTimestamp()
		open.User = args[0]
		open.Want.Color = args[1]
		open.Want.Size =  size1
		open.Willing.Color = args[3]
		open.Willing.Size =  size2
		
		//get the open trade struct
		tradesAsBytes, err := stub.GetState(openTradesStr)
		if err != nil {
			return shim.Error(err.Error())
		}
		var trades AllOpenTrades
		json.Unmarshal(tradesAsBytes, &trades)

		fmt.Printf("- Finished getting current open trades \n")
		fmt.Printf("- Adding new trade to opentrade \n")
		
		trades.OpenTrades = append(trades.OpenTrades, open)						//append to open trades
		fmt.Println("! appended open to trades")
		tradeJSONasBytes, _ := json.Marshal(trades)
		err = stub.PutState(openTradesStr, tradeJSONasBytes)								//rewrite open orders
		if err != nil {
			return shim.Error(err.Error())
		}
		fmt.Println("- end open trade")
		return shim.Success(nil)
	}

// ============================================================================================================================
// makeTimestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Second))
    // return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}	

// ===============================================
// readOpenTrade - read a readOpenTrade from chaincode state
// ===============================================
func (t *SimpleChaincode) readOpenTrade(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string

	valAsbytes, err := stub.GetState(openTradesStr) //get the openTrades from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + openTradesStr + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + openTradesStr + "\"}"
		return shim.Error(jsonResp)
	}
	
	var trades AllOpenTrades
	json.Unmarshal(valAsbytes, &trades)
	fmt.Println(trades)
	return shim.Success(valAsbytes)
}

// ===============================================
// removeOpenTrade - removeOpenTrade from chaincode state
// ===============================================
func (t *SimpleChaincode) removeOpenTrade(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting timestamp of the open trade")
	}

	fmt.Println("- start remove trade")
	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("1st argument must be a numeric string")
	}

	//get the open trade struct
	tradesAsBytes, err := stub.GetState(openTradesStr)
	if err != nil {
		return shim.Error("Failed to get opentrades")
	}
	var trades AllOpenTrades
	json.Unmarshal(tradesAsBytes, &trades)		

	for i := range trades.OpenTrades{																	//look for the trade
		//fmt.Println("looking at " + strconv.FormatInt(trades.OpenTrades[i].Timestamp, 10) + " for " + strconv.FormatInt(timestamp, 10))
		if trades.OpenTrades[i].Timestamp == timestamp{
			fmt.Println("found the trade");
			trades.OpenTrades = append(trades.OpenTrades[:i], trades.OpenTrades[i+1:]...)				//remove this trade
			tradesAsBytes, _ := json.Marshal(trades)
			err = stub.PutState(openTradesStr, tradesAsBytes)												//rewrite open orders
			if err != nil {
				return shim.Error(err.Error())
			}
			break
		}
	}
	
	fmt.Println("- end remove trade")
	fmt.Println(trades.OpenTrades)
	return shim.Success(nil)
}

//====================================================================================
// query the hyperleger with a queryString and convert the results(key value structure) into map(dict)
//====================================================================================
func getQueryResultForQueryStringtoMap(stub shim.ChaincodeStubInterface, queryString string) (map[string]interface{}, error) {
		objectMap := make(map[string]interface{})
		
		fmt.Printf("- getQueryResultForQueryStringtoMap queryString:\n%s\n", queryString)
	
		resultsIterator, err := stub.GetQueryResult(queryString)
		if err != nil {
			return nil, err
		}
		defer resultsIterator.Close()
	
		// buffer is a JSON array containing QueryRecords
		var buffer bytes.Buffer
		buffer.WriteString("[")

		for resultsIterator.HasNext() {
			queryResponse, err := resultsIterator.Next()
			if err != nil {
				return nil, err
			}
			// Record is a JSON object, so we write as-is
			var data map[string]interface{}
			err = json.Unmarshal([]byte(queryResponse.Value), &data)
			fmt.Println(data)
			objectMap[queryResponse.Key] = data
			// marbles[queryResponse.Key] = string(queryResponse.Value)
		}
		
		fmt.Printf("- getQueryResultForQueryStringtoMap queryResult:\n%s\n", objectMap)
		return objectMap, nil
	}

// ===============================================
// swapMarble - swap marble between two owners base on color and size ( without knowing marbleName)
// ===============================================

func (t *SimpleChaincode) swapMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//args = owner1, color1, size1, owner2, color2, size2 

	var owner1 = args[0]
	var color1 = args[1]
	// var size1 = args[2]
	var owner2 = args[3]
	var color2 = args[4]
	// var size2 = args[5]

	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6 args")
	}

	queryString1 := fmt.Sprintf("{\"selector\":{\"docType\":\"marble\",\"owner\":\"%s\"}}", owner1)
	
	queryString2 := fmt.Sprintf("{\"selector\":{\"docType\":\"marble\",\"owner\":\"%s\"}}", owner2)
	
	queryResults1, err := getQueryResultForQueryStringtoMap(stub, queryString1)
	queryResults2, err := getQueryResultForQueryStringtoMap(stub, queryString2)
	fmt.Printf("- swapMarble queryResults1:\n%s\n", queryResults1)
	fmt.Println(queryResults1)
	fmt.Println(queryResults2)

	marble1Name := ""
	marble2Name := ""
	for k, v := range queryResults1 {
		innermap, ok := v.(map[string]interface{})

		if !ok {
			panic("inner map is not a map!")
		}
		if (innermap["color"] == color1) {
			fmt.Println("marble1 bingo............")
			marble1Name = k
	
		}
		fmt.Printf("key[%s] value[%s]\n", k, innermap)
	}

	for k, v := range queryResults2 {
		innermap, ok := v.(map[string]interface{})

		if !ok {
			panic("inner map is not a map!")
		}
		if (innermap["color"] == color2) {
			fmt.Println("marble2 bingo............")
			marble2Name = k
		}
		fmt.Printf("key[%s] value[%s]\n", k, innermap)
	}

	if (marble1Name != "")  && (marble2Name != "") {
		fmt.Println("- swapMarble : start swapping marbles")

		response := t.transferMarble(stub, []string{marble1Name, owner2})

		// if the transfer failed break out of loop and return error
		if response.Status != shim.OK {
			return shim.Error("Transfer failed: " + response.Message)
		}
		response = t.transferMarble(stub, []string{marble2Name, owner1})
		// if the transfer failed break out of loop and return error
		if response.Status != shim.OK {
			return shim.Error("Transfer failed: " + response.Message)
		}
		fmt.Println("- swapMarble : finished swapping marbles")
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("success"))

}

// ===============================================
// swapMarbleTri - swap marble between two owners base on color and size ( without knowing marbleName)
// ===============================================

func (t *SimpleChaincode) swapMarbleTri(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
		//args = owner1, color1, size1, owner2, color2, size2 
	
		var owner1 = args[0]
		var color1 = args[1]
		// var size1 = args[2]
		var owner2 = args[3]
		var color2 = args[4]
		// var size2 = args[5]
		var owner3 = args[6]
		var color3 = args[7]
		// var size3 = args[8]

	
		if len(args) != 9 {
			return shim.Error("Incorrect number of arguments. Expecting 9 args")
		}
	
		queryString1 := fmt.Sprintf("{\"selector\":{\"docType\":\"marble\",\"owner\":\"%s\"}}", owner1)
		
		queryString2 := fmt.Sprintf("{\"selector\":{\"docType\":\"marble\",\"owner\":\"%s\"}}", owner2)

		queryString3 := fmt.Sprintf("{\"selector\":{\"docType\":\"marble\",\"owner\":\"%s\"}}", owner3)
		
		queryResults1, err := getQueryResultForQueryStringtoMap(stub, queryString1)
		queryResults2, err := getQueryResultForQueryStringtoMap(stub, queryString2)
		queryResults3, err := getQueryResultForQueryStringtoMap(stub, queryString3)
		fmt.Printf("- swapMarble queryResults1:\n%s\n", queryResults1)
		fmt.Printf("- swapMarble queryResults2:\n%s\n", queryResults2)
		fmt.Printf("- swapMarble queryResults3:\n%s\n", queryResults3)
		marble1Name := ""
		marble2Name := ""
		marble3Name := ""
		for k, v := range queryResults1 {
			innermap, ok := v.(map[string]interface{})
	
			if !ok {
				panic("inner map is not a map!")
			}
			if (innermap["color"] == color1) {
				fmt.Println("marble1 bingo............")
				marble1Name = k
		
			}
			fmt.Printf("key[%s] value[%s]\n", k, innermap)
		}
	
		for k, v := range queryResults2 {
			innermap, ok := v.(map[string]interface{})
	
			if !ok {
				panic("inner map is not a map!")
			}
			if (innermap["color"] == color2) {
				fmt.Println("marble2 bingo............")
				marble2Name = k
			}
			fmt.Printf("key[%s] value[%s]\n", k, innermap)
		}
		
		for k, v := range queryResults3 {
			innermap, ok := v.(map[string]interface{})
	
			if !ok {
				panic("inner map is not a map!")
			}
			if (innermap["color"] == color3) {
				fmt.Println("marble3 bingo............")
				marble3Name = k
			}
			fmt.Printf("key[%s] value[%s]\n", k, innermap)
		}

		if (marble1Name != "")  && (marble2Name != "")  && (marble3Name != ""){
			fmt.Println("- swapMarbleTri : start swapping marbles Tri")
	
			response := t.transferMarble(stub, []string{marble1Name, owner2})
	
			// if the transfer failed break out of loop and return error
			if response.Status != shim.OK {
				return shim.Error("Transfer failed: " + response.Message)
			}
			response = t.transferMarble(stub, []string{marble2Name, owner3})
			// if the transfer failed break out of loop and return error
			if response.Status != shim.OK {
				return shim.Error("Transfer failed: " + response.Message)
			}
			response = t.transferMarble(stub, []string{marble3Name, owner1})
			// if the transfer failed break out of loop and return error
			if response.Status != shim.OK {
				return shim.Error("Transfer failed: " + response.Message)
			}
			fmt.Println("- swapMarble : finished swapping marbles")
		}
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte("success"))
	
	}

// ===============================================
// matchTrade - match trades from within openTrades in chaincode state, compatibale with AnOpenTrade as slice in AllOpenTrades
// ===============================================

func (t *SimpleChaincode) matchTrade(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	valAsbytes, err := stub.GetState(openTradesStr) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + openTradesStr + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"opentrades does not exist: " + openTradesStr + "\"}"
		return shim.Error(jsonResp)
	}
	
	var openTradesStruct AllOpenTrades
	json.Unmarshal(valAsbytes, &openTradesStruct)
	fmt.Println("matchTrade")
	fmt.Println(openTradesStruct.OpenTrades)
	openTrades := openTradesStruct.OpenTrades
	for i := 0; i < len(openTrades); i++ {
		for j := i + 1; j < len(openTrades); j++ {
			if reflect.DeepEqual(openTrades[i].Want, openTrades[j].Willing) && reflect.DeepEqual(openTrades[i].Willing, openTrades[j].Want) {
				fmt.Println("swapMarbles")
				// swapMarbles
				t.swapMarble(stub, []string{openTrades[i].User, openTrades[i].Willing.Color, strconv.Itoa(openTrades[i].Willing.Size), openTrades[j].User, openTrades[j].Willing.Color, strconv.Itoa(openTrades[j].Willing.Size)})
				fmt.Println(i)
				fmt.Println(openTrades[i])
				// delete openTrades after matching orders
				// delete from hyperledger blockchain
				// t.removeOpenTrade(stub,[]string{strconv.FormatInt(openTrades[i].Timestamp, 10)})
				// t.removeOpenTrade(stub,[]string{strconv.FormatInt(openTrades[j].Timestamp, 10)})
				// fmt.Println(strconv.FormatInt(openTrades[i].Timestamp, 10))

				// delete from cache so that no re-matching can happen
				openTrades = append(openTrades[:j], openTrades[j+1:]...)
				openTrades = append(openTrades[:i], openTrades[i+1:]...)
				fmt.Println(openTrades)
				i-- // redo index since the orignal has been deleted
				break
			}
		}
	}

	fmt.Printf(" Saving new state of open trades to hyperledger:")
	fmt.Println(openTrades)
	openTradesStruct.OpenTrades = openTrades
	tradesAsBytes, _ := json.Marshal(openTradesStruct)
	err = stub.PutState(openTradesStr, tradesAsBytes)												//rewrite open orders
	if err != nil {
		return shim.Error(err.Error())
	}


	return shim.Success(nil)

}


// ===============================================
// matchTrade2 - match trades from within openTrades in chaincode state, compatibale with AnOpenTrade as seperate states
// ===============================================

func (t *SimpleChaincode) matchTrade2(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	
	
	
	queryString1 := fmt.Sprintf("{\"selector\":{\"docType\":\"openTrade\"}}")
		
	queryResults1, err := getQueryResultForQueryStringtoMap(stub, queryString1)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf("- matchTrade2 queryResults1:\n%s\n", queryResults1)
	fmt.Println(queryResults1)
	// convert the queryResults in to a slice of AnOpenTrades

	var openTrades []AnOpenTrade
	for  _, value := range queryResults1 {
		innermap, ok := value.(map[string]interface{})
		if !ok {
			panic("inner map is not a map!")
		}
		open := AnOpenTrade{}
		open.ObjectType = "openTrade"
		open.Timestamp = innermap["TimeStamp"].(int64)
		open.User = innermap["User"].(string)
		open.Want.Color = innermap["Want"].(map[string]interface{})["Color"].(string)
		open.Want.Size =  innermap["Want"].(map[string]interface{})["Size"].(int)
		open.Willing.Color = innermap["Willing"].(map[string]interface{})["Color"].(string)
		open.Willing.Size =  innermap["Willing"].(map[string]interface{})["Size"].(int)
		openTrades = append (openTrades, open )
	}
	fmt.Println("matchTrade2..................")
	fmt.Println(openTrades)

	// var openTradesStruct AllOpenTrades
	// json.Unmarshal(valAsbytes, &openTradesStruct)
	// fmt.Println(openTradesStruct.OpenTrades)
	// openTrades := openTradesStruct.OpenTrades
	for i := 0; i < len(openTrades); i++ {
		for j := i + 1; j < len(openTrades); j++ {
			if reflect.DeepEqual(openTrades[i].Want, openTrades[j].Willing) && reflect.DeepEqual(openTrades[i].Willing, openTrades[j].Want) {
				fmt.Println("swapMarbles")
				// swapMarbles
				t.swapMarble(stub, []string{openTrades[i].User, openTrades[i].Willing.Color, strconv.Itoa(openTrades[i].Willing.Size), openTrades[j].User, openTrades[j].Willing.Color, strconv.Itoa(openTrades[j].Willing.Size)})
				fmt.Println(i)
				fmt.Println(openTrades[i])
				// delete openTrades after matching orders
				// delete from hyperledger blockchain
				// t.removeOpenTrade(stub,[]string{strconv.FormatInt(openTrades[i].Timestamp, 10)})
				// t.removeOpenTrade(stub,[]string{strconv.FormatInt(openTrades[j].Timestamp, 10)})
				// fmt.Println(strconv.FormatInt(openTrades[i].Timestamp, 10))

				// delete from cache so that no re-matching can happen
				openTrades = append(openTrades[:j], openTrades[j+1:]...)
				openTrades = append(openTrades[:i], openTrades[i+1:]...)
				fmt.Println(openTrades)
				i-- // redo index since the orignal has been deleted
				break
			}
		}
	}

	// fmt.Printf(" Saving new state of open trades to hyperledger:")
	// fmt.Println(openTrades)
	// openTradesStruct.OpenTrades = openTrades
	// tradesAsBytes, _ := json.Marshal(openTradesStruct)
	// err = stub.PutState(openTradesStr, tradesAsBytes)												//rewrite open orders
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }


	return shim.Success(nil)

}

// ===============================================
// matchTriTrade - match trades from within openTrades in chaincode state, compatibale with AnOpenTrade as slice in AllOpenTrades
// ===============================================
func (t *SimpleChaincode) matchTriTrade(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	valAsbytes, err := stub.GetState(openTradesStr) //get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + openTradesStr + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"opentrades does not exist: " + openTradesStr + "\"}"
		return shim.Error(jsonResp)
	}
	
	var openTradesStruct AllOpenTrades
	json.Unmarshal(valAsbytes, &openTradesStruct)
	fmt.Println("matchTrade")
	fmt.Println(openTradesStruct.OpenTrades)
	openTrades := openTradesStruct.OpenTrades
	for i := 0; i < len(openTrades); i++ {
		for j := i + 1; j < len(openTrades); j++ {
			for k := j + 1; k < len(openTrades); k++{
				fmt.Println("matchTriTrade : compare opentrades")
				if (reflect.DeepEqual(openTrades[i].Want, openTrades[j].Willing) && reflect.DeepEqual(openTrades[j].Want, openTrades[k].Willing) && reflect.DeepEqual(openTrades[k].Want, openTrades[i].Willing)) ||
				(reflect.DeepEqual(openTrades[i].Want, openTrades[k].Willing) && reflect.DeepEqual(openTrades[k].Want, openTrades[j].Willing) && reflect.DeepEqual(openTrades[j].Want, openTrades[i].Willing)){
					fmt.Println("matchTriTrade - swapMarbles")
					// swapMarbles
					if (reflect.DeepEqual(openTrades[i].Want, openTrades[j].Willing) && reflect.DeepEqual(openTrades[j].Want, openTrades[k].Willing) && reflect.DeepEqual(openTrades[k].Want, openTrades[i].Willing)){
						fmt.Println("matchTriTrade - first case - swapMarbleTri")
						t.swapMarbleTri(stub, []string{openTrades[i].User, openTrades[i].Willing.Color, strconv.Itoa(openTrades[i].Willing.Size), openTrades[j].User, openTrades[j].Willing.Color, strconv.Itoa(openTrades[j].Willing.Size), openTrades[k].User, openTrades[k].Willing.Color, strconv.Itoa(openTrades[k].Willing.Size)})
						// following doesnt work......
						// t.swapMarble(stub, []string{openTrades[i].User, openTrades[i].Willing.Color, strconv.Itoa(openTrades[i].Willing.Size), openTrades[j].User, openTrades[j].Willing.Color, strconv.Itoa(openTrades[j].Willing.Size)})
						// t.swapMarble(stub, []string{openTrades[j].User, openTrades[i].Willing.Color, strconv.Itoa(openTrades[i].Willing.Size), openTrades[k].User, openTrades[k].Willing.Color, strconv.Itoa(openTrades[k].Willing.Size)})
						// fmt.Println(openTrades[i])
						
					}else {
						fmt.Println("matchTriTrade - second case - swapMarbleTri")
						t.swapMarbleTri(stub, []string{openTrades[i].User, openTrades[i].Willing.Color, strconv.Itoa(openTrades[i].Willing.Size), openTrades[k].User, openTrades[k].Willing.Color, strconv.Itoa(openTrades[k].Willing.Size), openTrades[j].User, openTrades[j].Willing.Color, strconv.Itoa(openTrades[j].Willing.Size)})
						
						// fmt.Println("matchTriTrade - second case - step 1")
						// t.swapMarble(stub, []string{openTrades[i].User, openTrades[i].Willing.Color, strconv.Itoa(openTrades[i].Willing.Size), openTrades[k].User, openTrades[k].Willing.Color, strconv.Itoa(openTrades[k].Willing.Size)})
						// fmt.Println("matchTriTrade - second case - step 2")
						// fmt.Println(openTrades[k].User)
						// fmt.Println(openTrades[k].Willing.Color)
						// fmt.Println(openTrades[i].User)
						// fmt.Println(openTrades[i].Willing.Color)
						// time.Sleep(2000 * time.Millisecond)
						// t.swapMarble(stub, []string{openTrades[k].User, openTrades[i].Willing.Color, strconv.Itoa(openTrades[i].Willing.Size), openTrades[j].User, openTrades[j].Willing.Color, strconv.Itoa(openTrades[j].Willing.Size)})
						// fmt.Println(openTrades[i])
					}
					// delete openTrades after matching orders
					// delete from hyperledger blockchain
					// t.removeOpenTrade(stub,[]string{strconv.FormatInt(openTrades[i].Timestamp, 10)})
					// t.removeOpenTrade(stub,[]string{strconv.FormatInt(openTrades[j].Timestamp, 10)})
					// fmt.Println(strconv.FormatInt(openTrades[i].Timestamp, 10))

					// delete from cache so that no re-matching can happen
					openTrades = append(openTrades[:k], openTrades[k+1:]...)
					openTrades = append(openTrades[:j], openTrades[j+1:]...)
					openTrades = append(openTrades[:i], openTrades[i+1:]...)
					fmt.Println(openTrades)
					i-- // redo index since the orignal has been deleted
					break
				}
			}
		}
	}

	fmt.Printf(" Saving new state of open trades to hyperledger:")
	fmt.Println(openTrades)
	openTradesStruct.OpenTrades = openTrades
	tradesAsBytes, _ := json.Marshal(openTradesStruct)
	err = stub.PutState(openTradesStr, tradesAsBytes)												//rewrite open orders
	if err != nil {
		return shim.Error(err.Error())
	}


	return shim.Success(nil)

}

// ===============================================
// clearOpenTrades - delete the slice in AllOpenTrades
// ===============================================

func (t *SimpleChaincode) clearOpenTrades(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	valAsbytes, err := stub.GetState(openTradesStr) //get the marble from chaincode state

	var trades AllOpenTrades
	json.Unmarshal(valAsbytes, &trades)		

	trades.OpenTrades = []AnOpenTrade{} 		//remove all trades
	tradesAsBytes, _ := json.Marshal(trades)
	err = stub.PutState(openTradesStr, tradesAsBytes)												//rewrite open orders
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("success"))
}
