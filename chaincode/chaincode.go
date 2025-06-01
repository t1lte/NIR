package chaincode


import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)


type SmartContract struct {
	contractapi.Contract
}


type StateMemory struct {
    Error bool `json:"Error"`
}

type ElementState int

const (
	DISABLED = iota
	ENABLED
	WAITINGFORCONFIRMATION
	COMPLETED
)

type Participant struct {
	MSP        string            `json:"msp"`
	Attributes map[string]string `json:"attributes"`
}

type Message struct {
	MessageID            string       `json:"messageID"`
	SendParticipantID    string       `json:"sendParticipantID"`
	ReceiveParticipantID string       `json:"receiveParticipantID"`
	MsgState             ElementState `json:"msgState"`
	Format               string       `json:"format"`
}

type Gateway struct {
	GatewayID    string       `json:"gatewayID"`
	GatewayState ElementState `json:"gatewayState"`
}

type ActionEvent struct {
	EventID    string       `json:"eventID"`
	EventState ElementState `json:"eventState"`
}

func (cc *SmartContract) CreateParticipant(ctx contractapi.TransactionContextInterface, participantID string, msp string, attributes map[string]string) (*Participant, error) {
	stub := ctx.GetStub()

	existingData, err := stub.GetState(participantID)
	if err != nil {
		return nil, fmt.Errorf("Error getting status data: %v", err)
	}
	if existingData != nil {
		return nil, fmt.Errorf("Participant %s already exists", participantID)
	}

	
	participant := &Participant{
		MSP:        msp,
		Attributes: attributes,
	}

	
	participantJSON, err := json.Marshal(participant)
	if err != nil {
		return nil, fmt.Errorf("Error serializing participant data: %v", err)
	}
	err = stub.PutState(participantID, participantJSON)
	if err != nil {
		return nil, fmt.Errorf("Error saving participant data: %v", err)
	}

	return participant, nil
}

func (cc *SmartContract) CreateMessage(ctx contractapi.TransactionContextInterface, messageID string, sendParticipantID string, receiveParticipantID string, msgState ElementState, format string) (*Message, error) {
	stub := ctx.GetStub()

	
	existingData, err := stub.GetState(messageID)
	if err != nil {
		return nil, fmt.Errorf("Error getting status data: %v", err)
	}
	if existingData != nil {
		return nil, fmt.Errorf("Message %s already exists", messageID)
	}

	
	msg := &Message{
		MessageID:            messageID,
		SendParticipantID:    sendParticipantID,
		ReceiveParticipantID: receiveParticipantID,
		MsgState:             msgState,
		Format:               format,
	}

	
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("Error serializing message data: %v", err)
	}
	err = stub.PutState(messageID, msgJSON)
	if err != nil {
		return nil, fmt.Errorf("Error saving message data: %v", err)
	}

	return msg, nil
}

func (cc *SmartContract) CreateGateway(ctx contractapi.TransactionContextInterface, gatewayID string, gatewayState ElementState) (*Gateway, error) {
	stub := ctx.GetStub()

	
	existingData, err := stub.GetState(gatewayID)
	if err != nil {
		return nil, fmt.Errorf("Error getting status data: %v", err)
	}
	if existingData != nil {
		return nil, fmt.Errorf("Gateway %s already exists", gatewayID)
	}

	
	gtw := &Gateway{
		GatewayID:    gatewayID,
		GatewayState: gatewayState,
	}

	
	gtwJSON, err := json.Marshal(gtw)
	if err != nil {
		return nil, fmt.Errorf("Error serializing gateway data: %v", err)
	}
	err = stub.PutState(gatewayID, gtwJSON)
	if err != nil {
		return nil, fmt.Errorf("Error saving gateway data: %v", err)
	}

	return gtw, nil
}

func (cc *SmartContract) CreateActionEvent(ctx contractapi.TransactionContextInterface, eventID string, eventState ElementState) (*ActionEvent, error) {
	stub := ctx.GetStub()

	
	actionEvent := &ActionEvent{
		EventID:    eventID,
		EventState: eventState,
	}

	
	actionEventJSON, err := json.Marshal(actionEvent)
	if err != nil {
		return nil, fmt.Errorf("Error serializing event data: %v", err)
	}
	err = stub.PutState(eventID, actionEventJSON)
	if err != nil {
		return nil, fmt.Errorf("Error saving event data: %v", err)
	}

	return actionEvent, nil
}


func (c *SmartContract) ReadMsg(ctx contractapi.TransactionContextInterface, messageID string) (*Message, error) {
	msgJSON, err := ctx.GetStub().GetState(messageID)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if msgJSON == nil {
		errorMessage := fmt.Sprintf("Message %s does not exist", messageID)
		fmt.Println(errorMessage)
		return nil, errors.New(errorMessage)
	}

	var msg Message
	err = json.Unmarshal(msgJSON, &msg)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &msg, nil
}

func (c *SmartContract) ReadGtw(ctx contractapi.TransactionContextInterface, gatewayID string) (*Gateway, error) {
	gtwJSON, err := ctx.GetStub().GetState(gatewayID)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if gtwJSON == nil {
		errorMessage := fmt.Sprintf("Gateway %s does not exist", gatewayID)
		fmt.Println(errorMessage)
		return nil, errors.New(errorMessage)
	}

	var gtw Gateway
	err = json.Unmarshal(gtwJSON, &gtw)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &gtw, nil
}

func (c *SmartContract) ReadEvent(ctx contractapi.TransactionContextInterface, eventID string) (*ActionEvent, error) {
	eventJSON, err := ctx.GetStub().GetState(eventID)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if eventJSON == nil {
		errorMessage := fmt.Sprintf("Event state %s does not exist", eventID)
		fmt.Println(errorMessage)
		return nil, errors.New(errorMessage)
	}

	var event ActionEvent
	err = json.Unmarshal(eventJSON, &event)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &event, nil
}


func (c *SmartContract) ChangeMsgState(ctx contractapi.TransactionContextInterface, messageID string, msgState ElementState) error {
	stub := ctx.GetStub()

	msg, err := c.ReadMsg(ctx, messageID)
	if err != nil {
		return err
	}

	msg.MsgState = msgState

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	err = stub.PutState(messageID, msgJSON)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func (c *SmartContract) ChangeGtwState(ctx contractapi.TransactionContextInterface, gatewayID string, gtwState ElementState) error {
	stub := ctx.GetStub()

	gtw, err := c.ReadGtw(ctx, gatewayID)
	if err != nil {
		return err
	}

	gtw.GatewayState = gtwState

	gtwJSON, err := json.Marshal(gtw)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	err = stub.PutState(gatewayID, gtwJSON)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func (c *SmartContract) ChangeEventState(ctx contractapi.TransactionContextInterface, eventID string, eventState ElementState) error {
	stub := ctx.GetStub()

	actionEvent, err := c.ReadEvent(ctx, eventID)
	if err != nil {
		return err
	}

	actionEvent.EventState = eventState

	actionEventJSON, err := json.Marshal(actionEvent)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	err = stub.PutState(eventID, actionEventJSON)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}



func (cc *SmartContract) GetAllMessages(ctx contractapi.TransactionContextInterface) ([]*Message, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("Error getting status data: %v", err) 
	}
	defer resultsIterator.Close()

	var messages []*Message
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("Error while iterating state data: %v", err)
		}

		var message Message
		err = json.Unmarshal(queryResponse.Value, &message)
		if strings.HasPrefix(message.MessageID, "Message") {
			if err != nil {
				return nil, fmt.Errorf("Error deserializing message data: %v", err)
			}

			messages = append(messages, &message)
		}
	}

	return messages, nil
}

func (cc *SmartContract) GetAllGateways(ctx contractapi.TransactionContextInterface) ([]*Gateway, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("Error getting status data: %v", err)
	}
	defer resultsIterator.Close()

	var gateways []*Gateway
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("Error while iterating state data: %v", err)
		}

		var gateway Gateway
		err = json.Unmarshal(queryResponse.Value, &gateway)
		if strings.HasPrefix(gateway.GatewayID, "ExclusiveGateway") ||
			strings.HasPrefix(gateway.GatewayID, "EventBasedGateway") ||
			strings.HasPrefix(gateway.GatewayID, "Gateway") ||
			strings.HasPrefix(gateway.GatewayID, "ParallelGateway") {
			if err != nil {
				return nil, fmt.Errorf("Error deserializing gateway data: %v", err)
			}

			gateways = append(gateways, &gateway)
		}
	}

	return gateways, nil
}

func (cc *SmartContract) GetAllActionEvents(ctx contractapi.TransactionContextInterface) ([]*ActionEvent, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("Error getting status data: %v", err)
	}
	defer resultsIterator.Close()

	var events []*ActionEvent
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("Error while iterating state data: %v", err)
		}

		var event ActionEvent
		err = json.Unmarshal(queryResponse.Value, &event)
		if strings.HasPrefix(event.EventID, "StartEvent") ||
			strings.HasPrefix(event.EventID, "Event") ||
			strings.HasPrefix(event.EventID, "EndEvent") {
			if err != nil {
				return nil, fmt.Errorf("Error deserializing event data: %v", err)
			}

			events = append(events, &event)
		}
	}

	return events, nil
}

func (cc *SmartContract) ReadParticipant(ctx contractapi.TransactionContextInterface, participantID string) (*Participant, error) {
	participantJSON, err := ctx.GetStub().GetState(participantID)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if participantJSON == nil {
		errorMessage := fmt.Sprintf("Participant %s does not exist", participantID)
		fmt.Println(errorMessage)
		return nil, errors.New(errorMessage)
	}

	var participant Participant
	err = json.Unmarshal(participantJSON, &participant)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &participant, nil
}

func (cc *SmartContract) check_msp(ctx contractapi.TransactionContextInterface, target_participant string) bool {
	
	targetParticipant, err := cc.ReadParticipant(ctx, target_participant)
	if err != nil {
		return false
	}
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return false
	}
	return mspID == targetParticipant.MSP
}

func (cc *SmartContract) check_attribute(ctx contractapi.TransactionContextInterface, target_participant string, attributeName string) bool {
	targetParticipant, err := cc.ReadParticipant(ctx, target_participant)
	if err != nil {
		return false
	}
	if ctx.GetClientIdentity().AssertAttributeValue(attributeName, targetParticipant.Attributes[attributeName]) != nil {
		return false
	}

	return true
}

func (cc *SmartContract) check_participant(ctx contractapi.TransactionContextInterface, target_participant string) bool {
	
	targetParticipant, err := cc.ReadParticipant(ctx, target_participant)
	if err != nil {
		return false
	}
	
	if targetParticipant.MSP != "" && cc.check_msp(ctx, target_participant) == false {
		return false
	}

	
	for key, _ := range targetParticipant.Attributes {
		if cc.check_attribute(ctx, target_participant, key) == false {
			return false
		}
	}

	return true
}


func (cc *SmartContract) ReadGlobalVariable(ctx contractapi.TransactionContextInterface) (*StateMemory, error) {
	stateJSON, err := ctx.GetStub().GetState("currentMemory")
	if err != nil {
		return nil, err
	}

	if stateJSON == nil {
		
		return &StateMemory{}, nil
	}

	var stateMemory StateMemory
	err = json.Unmarshal(stateJSON, &stateMemory)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &stateMemory, nil
}

func (cc *SmartContract) SetGlobalVariable(ctx contractapi.TransactionContextInterface, globalVariable *StateMemory) error {
	stub := ctx.GetStub()
	globaleMemoryJson, err := json.Marshal(globalVariable)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	err = stub.PutState("currentMemory", globaleMemoryJson)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func (cc *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()

	
	isInitedBytes, err := stub.GetState("isInited")
	if err != nil {
		return fmt.Errorf("Failed to get isInited: %v", err)
	}
	if isInitedBytes != nil {
		errorMessage := "Chaincode has already been initialized"
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	cc.CreateParticipant(ctx, "Participant_1pasf6v", "", map[string]string{})
	cc.CreateParticipant(ctx, "Participant_1tddbk5", "", map[string]string{})
	cc.CreateActionEvent(ctx, "StartEvent_1v2ab61", ENABLED)

	cc.CreateActionEvent(ctx, "EndEvent_17h95ah", DISABLED)
	cc.CreateMessage(ctx, "Message_1j4s0qh", "Participant_1tddbk5", "Participant_1pasf6v", DISABLED, `
		{"properties":{"error":{"type":"boolean","description":""}},"required":[],"files":{},"file required":[]}
	`)
	cc.CreateMessage(ctx, "Message_0gg08bf", "Participant_1tddbk5", "Participant_1pasf6v", DISABLED, `{}`)
	cc.CreateMessage(ctx, "Message_0i0xp6a", "Participant_1pasf6v", "Participant_1tddbk5", DISABLED, `{}`)
	cc.CreateMessage(ctx, "Message_1uiozoi", "Participant_1tddbk5", "Participant_1pasf6v", DISABLED, `{}`)
	cc.CreateMessage(ctx, "Message_1e90tfn", "Participant_1pasf6v", "Participant_1tddbk5", DISABLED, `{}`)
	cc.CreateGateway(ctx, "ExclusiveGateway_0c8hy9b", DISABLED)

	cc.CreateGateway(ctx, "ExclusiveGateway_1sp1v7s", DISABLED)

	stub.PutState("isInited", []byte("true"))

	stub.SetEvent("initContractEvent", []byte("Contract has been initialized successfully"))
	return nil
}


func (cc *SmartContract) StartEvent_1v2ab61(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()
	actionEvent, err := cc.ReadEvent(ctx, "StartEvent_1v2ab61")
	if err != nil {
		return err
	}

	if actionEvent.EventState != ENABLED {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", actionEvent.EventID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	cc.ChangeEventState(ctx, "StartEvent_1v2ab61", COMPLETED)
	stub.SetEvent("StartEvent_1v2ab61", []byte("Contract has been started successfully"))
	
	    cc.ChangeGtwState(ctx, "ExclusiveGateway_1sp1v7s", ENABLED)
	
	return nil
}

func (cc *SmartContract) Message_1e90tfn_Send(ctx contractapi.TransactionContextInterface ) error {
	stub := ctx.GetStub()
	msg, err := cc.ReadMsg(ctx, "Message_1e90tfn")
	if err != nil {
		return err
	}

	if cc.check_participant(ctx,msg.SendParticipantID) == false{
		errorMessage := fmt.Sprintf("Participant %s is not allowed to send the message", msg.SendParticipantID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	if msg.MsgState != ENABLED {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", msg.MessageID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	msg.MsgState = WAITINGFORCONFIRMATION
	msgJSON, _ := json.Marshal(msg)
	stub.PutState("Message_1e90tfn", msgJSON)
	
	stub.SetEvent("Message_1e90tfn", []byte("Message is waiting for confirmation"))

	
	return nil
}

func (cc *SmartContract) Message_1e90tfn_Complete(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()
	msg, err := cc.ReadMsg(ctx, "Message_1e90tfn")
	if err != nil {
		return err
	}

	if cc.check_participant(ctx,msg.ReceiveParticipantID) == false{
		errorMessage := fmt.Sprintf("Participant %s is not allowed to send the message", msg.ReceiveParticipantID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	if msg.MsgState != WAITINGFORCONFIRMATION {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", msg.MessageID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	cc.ChangeMsgState(ctx, "Message_1e90tfn", COMPLETED)
	stub.SetEvent("Message_1e90tfn", []byte("Message has been done"))

	
	    cc.ChangeMsgState(ctx, "Message_1uiozoi", ENABLED)

	
	return nil
}

func (cc *SmartContract) Message_1uiozoi_Send(ctx contractapi.TransactionContextInterface ) error {
	stub := ctx.GetStub()
	msg, err := cc.ReadMsg(ctx, "Message_1uiozoi")
	if err != nil {
		return err
	}

	if cc.check_participant(ctx,msg.SendParticipantID) == false{
		errorMessage := fmt.Sprintf("Participant %s is not allowed to send the message", msg.SendParticipantID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	if msg.MsgState != ENABLED {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", msg.MessageID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	msg.MsgState = WAITINGFORCONFIRMATION
	msgJSON, _ := json.Marshal(msg)
	stub.PutState("Message_1uiozoi", msgJSON)
	
	stub.SetEvent("Message_1uiozoi", []byte("Message is waiting for confirmation"))

	
	return nil
}

func (cc *SmartContract) Message_1uiozoi_Complete(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()
	msg, err := cc.ReadMsg(ctx, "Message_1uiozoi")
	if err != nil {
		return err
	}

	if cc.check_participant(ctx,msg.ReceiveParticipantID) == false{
		errorMessage := fmt.Sprintf("Participant %s is not allowed to send the message", msg.ReceiveParticipantID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	if msg.MsgState != WAITINGFORCONFIRMATION {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", msg.MessageID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	cc.ChangeMsgState(ctx, "Message_1uiozoi", COMPLETED)
	stub.SetEvent("Message_1uiozoi", []byte("Message has been done"))

	
	    cc.ChangeMsgState(ctx, "Message_0i0xp6a", ENABLED)

	
	return nil
}

func (cc *SmartContract) Message_0i0xp6a_Send(ctx contractapi.TransactionContextInterface ) error {
	stub := ctx.GetStub()
	msg, err := cc.ReadMsg(ctx, "Message_0i0xp6a")
	if err != nil {
		return err
	}

	if cc.check_participant(ctx,msg.SendParticipantID) == false{
		errorMessage := fmt.Sprintf("Participant %s is not allowed to send the message", msg.SendParticipantID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	if msg.MsgState != ENABLED {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", msg.MessageID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	msg.MsgState = WAITINGFORCONFIRMATION
	msgJSON, _ := json.Marshal(msg)
	stub.PutState("Message_0i0xp6a", msgJSON)
	
	stub.SetEvent("Message_0i0xp6a", []byte("Message is waiting for confirmation"))

	
	return nil
}

func (cc *SmartContract) Message_0i0xp6a_Complete(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()
	msg, err := cc.ReadMsg(ctx, "Message_0i0xp6a")
	if err != nil {
		return err
	}

	if cc.check_participant(ctx,msg.ReceiveParticipantID) == false{
		errorMessage := fmt.Sprintf("Participant %s is not allowed to send the message", msg.ReceiveParticipantID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	if msg.MsgState != WAITINGFORCONFIRMATION {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", msg.MessageID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	cc.ChangeMsgState(ctx, "Message_0i0xp6a", COMPLETED)
	stub.SetEvent("Message_0i0xp6a", []byte("Message has been done"))

	
	    cc.ChangeMsgState(ctx, "Message_0gg08bf", ENABLED)

	
	return nil
}

func (cc *SmartContract) Message_0gg08bf_Send(ctx contractapi.TransactionContextInterface ) error {
	stub := ctx.GetStub()
	msg, err := cc.ReadMsg(ctx, "Message_0gg08bf")
	if err != nil {
		return err
	}

	if cc.check_participant(ctx,msg.SendParticipantID) == false{
		errorMessage := fmt.Sprintf("Participant %s is not allowed to send the message", msg.SendParticipantID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	if msg.MsgState != ENABLED {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", msg.MessageID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	msg.MsgState = WAITINGFORCONFIRMATION
	msgJSON, _ := json.Marshal(msg)
	stub.PutState("Message_0gg08bf", msgJSON)
	
	stub.SetEvent("Message_0gg08bf", []byte("Message is waiting for confirmation"))

	
	return nil
}

func (cc *SmartContract) Message_0gg08bf_Complete(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()
	msg, err := cc.ReadMsg(ctx, "Message_0gg08bf")
	if err != nil {
		return err
	}

	if cc.check_participant(ctx,msg.ReceiveParticipantID) == false{
		errorMessage := fmt.Sprintf("Participant %s is not allowed to send the message", msg.ReceiveParticipantID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	if msg.MsgState != WAITINGFORCONFIRMATION {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", msg.MessageID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	cc.ChangeMsgState(ctx, "Message_0gg08bf", COMPLETED)
	stub.SetEvent("Message_0gg08bf", []byte("Message has been done"))

	
	    cc.ChangeMsgState(ctx, "Message_1j4s0qh", ENABLED)

	
	return nil
}

func (cc *SmartContract) Message_1j4s0qh_Send(ctx contractapi.TransactionContextInterface , Error bool) error {
	stub := ctx.GetStub()
	msg, err := cc.ReadMsg(ctx, "Message_1j4s0qh")
	if err != nil {
		return err
	}

	if cc.check_participant(ctx,msg.SendParticipantID) == false{
		errorMessage := fmt.Sprintf("Participant %s is not allowed to send the message", msg.SendParticipantID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	if msg.MsgState != ENABLED {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", msg.MessageID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	msg.MsgState = WAITINGFORCONFIRMATION
	msgJSON, _ := json.Marshal(msg)
	stub.PutState("Message_1j4s0qh", msgJSON)
		globalMemory,readGlobalError := cc.ReadGlobalVariable(ctx)
	if readGlobalError != nil {
		fmt.Println(readGlobalError.Error())
		return readGlobalError
	}
	globalMemory.Error = Error
	setGlobalError :=cc.SetGlobalVariable(ctx, globalMemory)
	if setGlobalError != nil {
		fmt.Println(setGlobalError.Error())
		return setGlobalError
	}
	stub.SetEvent("Message_1j4s0qh", []byte("Message is waiting for confirmation"))

	
	return nil
}

func (cc *SmartContract) Message_1j4s0qh_Complete(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()
	msg, err := cc.ReadMsg(ctx, "Message_1j4s0qh")
	if err != nil {
		return err
	}

	if cc.check_participant(ctx,msg.ReceiveParticipantID) == false{
		errorMessage := fmt.Sprintf("Participant %s is not allowed to send the message", msg.ReceiveParticipantID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	if msg.MsgState != WAITINGFORCONFIRMATION {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", msg.MessageID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	cc.ChangeMsgState(ctx, "Message_1j4s0qh", COMPLETED)
	stub.SetEvent("Message_1j4s0qh", []byte("Message has been done"))

	
	    cc.ChangeGtwState(ctx, "ExclusiveGateway_0c8hy9b", ENABLED)

	
	return nil
}

func (cc *SmartContract) ExclusiveGateway_0c8hy9b(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()
	gtw, err := cc.ReadGtw(ctx, "ExclusiveGateway_0c8hy9b")
	if err != nil {
		return err
	}

	if gtw.GatewayState != ENABLED {
		errorMessage := fmt.Sprintf("Gateway state %s is not allowed", gtw.GatewayID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	cc.ChangeGtwState(ctx, "ExclusiveGateway_0c8hy9b", COMPLETED)
	stub.SetEvent("ExclusiveGateway_0c8hy9b", []byte("ExclusiveGateway has been done"))

    
    	currentMemory, err := cc.ReadGlobalVariable(ctx)
	if err != nil {
		return err
	}

    Error:=currentMemory.Error

if Error==true {
	    cc.ChangeGtwState(ctx, "ExclusiveGateway_1sp1v7s", ENABLED)
}
if Error==false {
	    cc.ChangeEventState(ctx, "EndEvent_17h95ah", ENABLED)
}
    

	return nil
}

func (cc *SmartContract) ExclusiveGateway_1sp1v7s(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()
	gtw, err := cc.ReadGtw(ctx, "ExclusiveGateway_1sp1v7s")
	if err != nil {
		return err
	}

	if gtw.GatewayState != ENABLED {
		errorMessage := fmt.Sprintf("Gateway state %s is not allowed", gtw.GatewayID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	cc.ChangeGtwState(ctx, "ExclusiveGateway_1sp1v7s", COMPLETED)
	stub.SetEvent("ExclusiveGateway_1sp1v7s", []byte("ExclusiveGateway has been done"))

    
        cc.ChangeMsgState(ctx, "Message_1e90tfn", ENABLED)
    

	return nil
}

func (cc *SmartContract) EndEvent_17h95ah(ctx contractapi.TransactionContextInterface) error {
	stub := ctx.GetStub()
	event, err := cc.ReadEvent(ctx, "EndEvent_17h95ah")
	if err != nil {
		return err
	}

	if event.EventState != ENABLED {
		errorMessage := fmt.Sprintf("Event state %s is not allowed", event.EventID)
		fmt.Println(errorMessage)
		return fmt.Errorf(errorMessage)
	}

	cc.ChangeEventState(ctx, "EndEvent_17h95ah", COMPLETED)
	stub.SetEvent("EndEvent_17h95ah", []byte("EndEvent has been done"))
	
	return nil
}