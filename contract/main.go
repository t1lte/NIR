package main

import (
    "log"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
    "mychaincode/chaincode"
)

func main() {
    cc, err := contractapi.NewChaincode(new(chaincode.SmartContract))
    if err != nil {
        log.Panicf("Error creating codechain: %v", err)
    }

    if err := cc.Start(); err != nil {
        log.Panicf("Error starting chaincode: %v", err)
    }
}