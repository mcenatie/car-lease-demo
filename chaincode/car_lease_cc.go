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
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"strings"

	"github.com/openblockchain/obc-peer/openchain/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var v5cIndexStr = "_v5cindex"				//name for the key/value that will store a list of all known v5c

type V5C struct{
	Id string `json:"id"`
	Vin string `json:"vin"`					
	Make string `json:"make"`
	Model string `json:"model"`
	Rego string `json:"rego"`
	Owner string `json:"owner"`
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) init(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				
	if err != nil {
		return nil, err
	}
	
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(v5cIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.init(stub, args)
	} else if function == "delete" {										//deletes an entity from its state
		return t.Delete(stub, args)
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "init_v5c" {										//create a new v5c
		return t.init_v5c(stub, args)
	} else if function == "set_owner" {										//change owner of a marble
		return t.set_owner(stub, args)
	} else if function == "update_v5c" {										//change owner of a marble
		return t.update_v5c(stub, args)
	}
	fmt.Println("run did not find func: " + function)						//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Delete - remove a key/value pair from state
// ============================================================================================================================
func (t *SimpleChaincode) Delete(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	
	id := args[0]
	err := stub.DelState(id)													//remove the key from chaincode state
	if err != nil {
		return nil, errors.New("Failed to delete state")
	}

	//get the v5c index
	v5csAsBytes, err := stub.GetState(v5cIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get v5c index")
	}
	var v5cIndex []string
	fmt.Println(v5csAsBytes)
	json.Unmarshal(v5csAsBytes, &v5cIndex)								//un stringify it aka JSON.parse()
	
	//remove v5c from index
	for i,val := range v5cIndex{
		fmt.Println(strconv.Itoa(i) + " - looking at " + val + " for " + id)
		if val == id{															//find the correct v5c
			fmt.Println("found v5c")
			v5cIndex = append(v5cIndex[:i], v5cIndex[i+1:]...)			//remove it
			for x:= range v5cIndex{											//debug prints...
				fmt.Println(string(x) + " - " + v5cIndex[x])
			}
			break
		}
	}
	jsonAsBytes, _ := json.Marshal(v5cIndex)									//save new index
	err = stub.PutState(v5cIndexStr, jsonAsBytes)
	return nil, nil
}

// ============================================================================================================================
// Query - read a variable from chaincode state - (aka read)
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if function != "query" {
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	}
	var id, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting id of the v5c to query")
	}

	id = args[0]
	valAsbytes, err := stub.GetState(id)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + id + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) Write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var id, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. id of the variable and value to set")
	}

	id = args[0]														
	value = args[1]
	err = stub.PutState(id, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init V5C - create a new v5c, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_v5c(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error

	//   0       1       2       3      4       5
	// "id",    "VIN", "MAKE","MODEL","REGO","OWNER"
	if len(args) != 6 {
		return nil, errors.New("Incorrect number of arguments. Expecting 6")
	}

	fmt.Println("- start init v5c")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return nil, errors.New("6th argument must be a non-empty string")
	}
		
	vin := strings.ToLower(args[1])
	mk := strings.ToLower(args[2])
	model := strings.ToLower(args[3])
	rego := strings.ToLower(args[4])
	owner := strings.ToLower(args[5])

	str := `{"id": "` + args[0] + `", "vin": "` + vin + `", "make": "` + mk + `", "model": "` + model + `", "rego": "` + rego + `", "owner": "` + owner + `"}`
	err = stub.PutState(args[0], []byte(str))								//store v5c with id as key
	if err != nil {
		return nil, err
	}
		
	//get the v5c index
	v5csAsBytes, err := stub.GetState(v5cIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get v5c index")
	}
	var v5cIndex []string
	json.Unmarshal(v5csAsBytes, &v5cIndex)							//un stringify it aka JSON.parse()
	
	//append
	v5cIndex = append(v5cIndex, args[0])								//add v5c name to index list
	fmt.Println("! v5c index: ", v5cIndex)
	jsonAsBytes, _ := json.Marshal(v5cIndex)
	err = stub.PutState(v5cIndexStr, jsonAsBytes)						//store name of v5c

	fmt.Println("- end init v5c")
	return nil, nil
}

// ============================================================================================================================
// Transfer Ownership on V5C
// ============================================================================================================================
func (t *SimpleChaincode) set_owner(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error
	
	//   0       1
	// "id", 	"owner"
	if len(args) < 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}
	
	fmt.Println("- start set owner")
	fmt.Println(args[0] + " - " + args[1])
	v5cAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := V5C{}
	json.Unmarshal(v5cAsBytes, &res)										//un stringify it aka JSON.parse()
	res.Owner = args[1]														//change the owner
	
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the v5c with id as key
	if err != nil {
		return nil, err
	}
	
	fmt.Println("- end set owner")
	return nil, nil
}

// ============================================================================================================================
// Update V5C
// ============================================================================================================================
func (t *SimpleChaincode) update_v5c(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error
	
	//   0       1       2       3      4       
	// "id",    "VIN", "MAKE","MODEL","REGO"
	if len(args) < 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting 5")
	}
	
	fmt.Println("- start update V5C")
	fmt.Println(args[0] + " - " + args[1] + " - " + args[2] + " - " + args[3] + " - " + args[4])
	v5cAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return nil, errors.New("Failed to get thing")
	}
	res := V5C{}
	json.Unmarshal(v5cAsBytes, &res)										//un stringify it aka JSON.parse()
	res.Vin = args[1]	
	res.Make = args[2]
	res.Model = args[3]
	res.Rego = args[4]
	
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(args[0], jsonAsBytes)								//rewrite the v5c with id as key
	if err != nil {
		return nil, err
	}
	
	fmt.Println("- end update V5C")
	return nil, nil
}