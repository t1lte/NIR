from typing import List, Optional, Tuple, Any, Protocol, Union
from elements import (
    Participant,
    StartEvent,
    EndEvent,
    Message,
    NodeType,
    EdgeType,
    ChoreographyTask,
    ExclusiveGateway,
    ParallelGateway,
    EventBasedGateway,
    MessageFlow,
    SequenceFlow,
    Element,
)
from parser import Choreography
import chaincode_snippet as snippet
import json


def type_change_from_bpmn_to_go(type: str) -> str:
    if type == "string":
        return "string"
    if type == "integer":
        return "int"
    if type == "boolean":
        return "bool"
    if type == "float":
        return "float64"
    return type


def public_the_name(name: str) -> str:
    return name.capitalize()


class GoChaincodeTranslator:
    def __init__(self, bpmnContent: str, bpmn_file: str = None):
        self._choreography: Optional[Choreography] = None
        self._global_variabels: Optional[dict] = None
        self._judge_parameters: Optional[dict] = None
        self._hook_codes: Optional[dict] = None
        choreography: Choreography = Choreography()
        if bpmnContent:
            choreography.load_diagram_from_string(bpmnContent)
        elif bpmn_file:
            choreography.load_diagram_from_xml_file(bpmn_file)
        else:
            pass
        self._choreography = choreography
        self._global_parameters, self._judge_parameters = (
            self._extract_global_parameters()
        )

    def _extract_global_parameters(self) -> dict:
        choreography = self._choreography
        global_parameters = {}
        judge_parameters = (
            {}
        )
        message_properties = {}
        for message in choreography.query_element_with_type(NodeType.MESSAGE):
            if message.documentation == "{}":
                continue
            document_dict = json.loads(message.documentation)
            for name, attri in document_dict["properties"].items():
                message_properties[name] = {
                    **{
                        "message_id": (
                            [message.id] + message_properties[name]["message_id"]
                            if name in message_properties
                            else [message.id]
                        )
                    },
                    **attri,
                }

        for sequence_flow in choreography.query_element_with_type(
                EdgeType.SEQUENCE_FLOW
        ):
            name = sequence_flow.name
            if name == "":
                continue
            match name:
                case x if "==" in x:
                    prop, value = x.split("==")
                    prop_defination = message_properties.get(prop)
                    if prop_defination is None:
                        continue
                    global_parameters[prop] = {
                        "definition": prop_defination,
                    }
                    judge_parameters[sequence_flow.id] = {
                        "name": prop,
                        "value": value,
                        "type": prop_defination["type"],
                        "relation": "==",
                    }
                case x if "!=" in x:
                    prop, value = x.split("!=")
                    prop_defination = message_properties.get(prop)
                    if prop_defination is None:
                        continue
                    global_parameters[prop] = {
                        "definition": prop_defination,
                    }
                    judge_parameters[sequence_flow.id] = {
                        "name": prop,
                        "value": value,
                        "type": prop_defination["type"],
                        "relation": "!=",
                    }
                case x if ">" in x:
                    prop, value = x.split(">")
                    prop_defination = message_properties.get(prop)
                    if prop_defination is None:
                        continue
                    global_parameters[prop] = {
                        "definition": prop_defination,
                    }
                    judge_parameters[sequence_flow.id] = {
                        "name": prop,
                        "value": value,
                        "type": prop_defination["type"],
                        "relation": ">",
                    }
                case x if "<" in x:
                    prop, value = x.split("<")
                    prop_defination = message_properties.get(prop)
                    if prop_defination is None:
                        continue
                    global_parameters[prop] = {
                        "definition": prop_defination,
                    }
                    judge_parameters[sequence_flow.id] = {
                        "name": prop,
                        "value": value,
                        "type": prop_defination["type"],
                        "relation": "<",
                    }
                case x if ">=" in x:
                    prop, value = x.split(">=")
                    prop_defination = message_properties.get(prop)
                    if prop_defination is None:
                        continue
                    global_parameters[prop] = {
                        "definition": prop_defination,
                    }
                    judge_parameters[sequence_flow.id] = {
                        "name": prop,
                        "value": value,
                        "type": prop_defination["type"],
                        "relation": ">=",
                    }
                case x if "<=" in x:
                    prop, value = x.split("<=")
                    prop_defination = message_properties.get(prop)
                    if prop_defination is None:
                        continue
                    global_parameters[prop] = {
                        "definition": prop_defination,
                    }
                    judge_parameters[sequence_flow.id] = {
                        "name": prop,
                        "value": value,
                        "type": prop_defination["type"],
                        "relation": "<=",
                    }

        return global_parameters, judge_parameters

    def _generate_parameters_code(self) -> str:
        global_parameters = self._global_parameters
        temp_list = []
        for name, prop in global_parameters.items():
            type = prop["definition"]["type"]
            temp_list.append(
                snippet.StateMemoryParameterDefinition_code(
                    public_the_name(name), type_change_from_bpmn_to_go(type)
                )
            )
        return "\n\t".join(temp_list)

    def _generate_InitLedger(self, bindings: dict[str, dict] = {}):
        choreography = self._choreography
        temp_list = []
        start_event: StartEvent = choreography.query_element_with_type(
            NodeType.START_EVENT
        )[0]
        end_events: EndEvent = choreography.query_element_with_type(NodeType.END_EVENT)
        message_flows: List[MessageFlow] = choreography.query_element_with_type(
            EdgeType.MESSAGE_FLOW
        )
        gateways: List[Union[ExclusiveGateway, ParallelGateway, EventBasedGateway]] = (
                choreography.query_element_with_type(NodeType.EXCLUSIVE_GATEWAY)
                + choreography.query_element_with_type(NodeType.PARALLEL_GATEWAY)
                + choreography.query_element_with_type(NodeType.EVENT_BASED_GATEWAY)
        )

        participants_exist = set(
            [
                element.id
                for element in choreography.query_element_with_type(
                NodeType.PARTICIPANT
            )
            ]
        )
        participants_in_bindings = set(bindings.keys())
        participant_assigned_in_bindings = participants_exist & participants_in_bindings
        participant_left_for_init = participants_exist - participants_in_bindings
        participant_to_be_added = [
                                      {
                                          "id": participant,
                                          "msp": bindings[participant]["msp"],
                                          "attributes": bindings[participant]["attributes"],
                                      }
                                      for participant in participant_assigned_in_bindings
                                  ] + [
                                      {
                                          "id": participant,
                                          "msp": "",
                                          "attributes": {},
                                      }
                                      for participant in participant_left_for_init
                                  ]

        temp_list.append(
            snippet.InitLedger_code(
                start_event=start_event.id,
                end_events=[end_event.id for end_event in end_events],
                messages=[
                    {
                        "name": message_flow.message.id,
                        "sender": message_flow.source.id,
                        "receiver": message_flow.target.id,
                        "properties": message_flow.message.documentation,
                    }
                    for message_flow in message_flows
                ],
                gateways=[gateway.id for gateway in gateways],
                participants=participant_to_be_added,
            )
        )
        return temp_list

    def _generate_change_state_code(
            self, element: Element, state: str = "ENABLED"
    ) -> str:
        match element.type:
            case NodeType.CHOREOGRAPHY_TASK:
                return snippet.ChangeMsgState_code(
                    element.init_message_flow.message.id, state
                )
            case (
            NodeType.EXCLUSIVE_GATEWAY
            | NodeType.PARALLEL_GATEWAY
            | NodeType.EVENT_BASED_GATEWAY
            ):
                return snippet.ChangeGtwState_code(element.id, state)
            case NodeType.END_EVENT:
                return snippet.ChangeEventState_code(element.id, state)
            case NodeType.MESSAGE:
                return snippet.ChangeMsgState_code(element.id, state)

    def _generate_check_state_code(self, element: Element, state: str = "ENABLED"):
        match element.type:
            case NodeType.CHOREOGRAPHY_TASK:
                return snippet.CheckMsgState_code(
                    element.init_message_flow.message.id, state
                )
            case (
            NodeType.EXCLUSIVE_GATEWAY
            | NodeType.PARALLEL_GATEWAY
            | NodeType.EVENT_BASED_GATEWAY
            ):
                return snippet.CheckGatewayState_code(element.id, state)
            case NodeType.END_EVENT:
                return snippet.CheckEventState_code(element.id, state)

    def _get_message_params(self, message: Message):
        global_parameters = self._global_parameters
        params_to_add = []
        for parameter in global_parameters:
            if message.id in global_parameters[parameter]["definition"]["message_id"]:
                params_to_add.append(
                    (parameter, global_parameters[parameter]["definition"]["type"])
                )
        return params_to_add

    def _generate_message_record_parameters_code(self, message: Message):
        params_to_add = self._get_message_params(message)
        if params_to_add:
            more_params_code = ", " + ", ".join(
                [
                    public_the_name(param[0]) + " " + type_change_from_bpmn_to_go(param[1])
                    for param in params_to_add
                ]
            )
        else:
            more_params_code = ""

        put_more_params_code = "\n".join(
            [
                snippet.SetGlobalVariable_code(
                    name=public_the_name(param[0]), value=public_the_name(param[0])
                )
                for param in params_to_add
            ]
        )
        return more_params_code, put_more_params_code

    def _event_based_gateway_hook_code(
            self, event_based_gateway: EventBasedGateway, currentElement: Element
    ):
        other_elements = []
        for outgoing in event_based_gateway.outgoings:
            if outgoing.target != currentElement:
                other_elements.append(outgoing.target)
        temp_list = [
            self._generate_change_state_code(element, "DISABLED")
            for element in other_elements
        ]
        return "\n".join(temp_list)

    def _parallel_gateway_merge_hook_code(
            self, parallel_gateway: ParallelGateway, currentElement: Element
    ):
        other_elements = []
        for incoming in parallel_gateway.incomings:
            if incoming.source != currentElement:
                other_elements.append(incoming.source)
        conditions = [
            self._generate_check_state_code(element, "COMPLETED")
            for element in other_elements
        ]
        combined_condition = snippet.CombineConditions_Any_False_code(conditions)
        return snippet.ConditionToHalt_code(combined_condition)

    def _generate_chaincode_for_choreography_task(
            self,
            choreography_task: ChoreographyTask,
    ):
        temp_list = []
        next_element = choreography_task.outgoing.target
        init_message_flow = choreography_task.init_message_flow
        return_message_flow = choreography_task.return_message_flow

        pre_activate_next_hook = self._hook_codes[choreography_task.id][
            "pre_activate_next"
        ]
        when_triggered_code = self._hook_codes[choreography_task.id]["when_triggered"]

        if not init_message_flow:
            return temp_list

        if not return_message_flow:
            more_parameters, put_more_parameters = (
                self._generate_message_record_parameters_code(init_message_flow.message)
            )
            temp_list.append(
                snippet.MessageSend_code(
                    message=init_message_flow.message.id,
                    after_all_hook="\n".join(when_triggered_code),
                    more_parameters=more_parameters,
                    put_more_parameters=put_more_parameters,
                )
            )
            temp_list.append(
                snippet.MessageComplete_code(
                    message=init_message_flow.message.id,
                    change_next_state_code=self._generate_change_state_code(
                        return_message_flow.message
                        if return_message_flow
                        else next_element
                    ),
                    pre_activate_next_hook="\n\t".join(pre_activate_next_hook),
                )
            )
            return temp_list

        more_parameters, put_more_parameters = (
            self._generate_message_record_parameters_code(init_message_flow.message)
        )
        temp_list.append(
            snippet.MessageSend_code(
                message=init_message_flow.message.id,
                after_all_hook="\n\t".join(when_triggered_code),
                more_parameters=more_parameters,
                put_more_parameters=put_more_parameters,
            )
        )
        temp_list.append(
            snippet.MessageComplete_code(
                message=init_message_flow.message.id,
                change_next_state_code=self._generate_change_state_code(
                    return_message_flow.message if return_message_flow else next_element
                ),
            )
        )
        more_parameters, put_more_parameters = (
            self._generate_message_record_parameters_code(return_message_flow.message)
        )
        temp_list.append(
            snippet.MessageSend_code(
                message=return_message_flow.message.id,
                more_parameters=more_parameters,
                put_more_parameters=put_more_parameters,
            )
        )
        temp_list.append(
            snippet.MessageComplete_code(
                message=return_message_flow.message.id,
                change_next_state_code=self._generate_change_state_code(next_element),
                pre_activate_next_hook="\n\t".join(pre_activate_next_hook),
            )
        )
        return temp_list

    def _generate_fullfill_condition_code(self, sequence_flow: SequenceFlow):
        judge_parameters = self._judge_parameters
        if sequence_flow.id in judge_parameters:
            parameter = judge_parameters[sequence_flow.id]
            return (
                    public_the_name(parameter["name"])
                    + parameter["relation"]
                    + parameter["value"]
            )
        return "true"

    def _generate_chaincode_for_exclusive_gateway(
            self,
            exclusive_gateway: ExclusiveGateway,
    ):
        judge_parameters = self._judge_parameters
        temp_list = []
        pre_activate_next_hook = self._hook_codes[exclusive_gateway.id][
            "pre_activate_next"
        ]
        when_triggered_code = self._hook_codes[exclusive_gateway.id]["when_triggered"]

        if len(exclusive_gateway.incomings) == 1:
            code = snippet.ExclusiveGateway_split_code(
                gateway=exclusive_gateway.id,
                change_next_state_code="\n".join(
                    [snippet.ReadGlobalMemory_code()]
                    + list(
                        set(
                            [
                                snippet.ReadState_code(
                                    public_the_name(
                                        judge_parameters[outgoing.id]["name"]
                                    )
                                )
                                for outgoing in exclusive_gateway.outgoings
                                if outgoing.id in judge_parameters
                            ]
                        )
                    )
                    + [
                        snippet.ConditionToDo_code(
                            self._generate_fullfill_condition_code(outgoing),
                            self._generate_change_state_code(outgoing.target),
                        )
                        for outgoing in exclusive_gateway.outgoings
                    ]
                ),
                pre_activate_next_hook="\n\t".join(pre_activate_next_hook),
                after_all_hook="\n\t".join(when_triggered_code),
            )
            temp_list.append(code)
        else:
            code = snippet.ExclusiveGateway_merge_code(
                gateway=exclusive_gateway.id,
                change_next_state_code=self._generate_change_state_code(
                    exclusive_gateway.outgoings[0].target
                ),
                pre_activate_next_hook="\n\t".join(pre_activate_next_hook),
                after_all_hook="\n\t".join(when_triggered_code),
            )
            temp_list.append(code)
        return temp_list

    def _generate_chaincode_for_parallel_gateway(
            self, parallel_gateway: ParallelGateway
    ):
        temp_list = []
        pre_activate_next_hook = self._hook_codes[parallel_gateway.id][
            "pre_activate_next"
        ]
        when_triggered_code = self._hook_codes[parallel_gateway.id]["when_triggered"]
        if len(parallel_gateway.incomings) == 1:
            code = snippet.ParallelGateway_split_code(
                gateway=parallel_gateway.id,
                change_next_state_code="\n".join(
                    [
                        self._generate_change_state_code(outgoing.target)
                        for outgoing in parallel_gateway.outgoings
                    ]
                ),
                pre_activate_next_hook="\n\t".join(pre_activate_next_hook),
                after_all_hook="\n\t".join(when_triggered_code),
            )
            temp_list.append(code)
        else:
            code = snippet.ParallelGateway_merge_code(
                gateway=parallel_gateway.id,
                change_next_state_code="\n".join(
                    [
                        self._generate_change_state_code(
                            parallel_gateway.outgoings[0].target
                        )
                    ]
                ),
                pre_activate_next_hook="\n\t".join(pre_activate_next_hook),
                after_all_hook="\n\t".join(when_triggered_code),
            )
            temp_list.append(code)

        return temp_list

    def _generate_chaincode_for_event_based_gateway(
            self,
            event_based_gateway: EventBasedGateway,
    ):
        temp_list = []
        pre_activate_next_hook = self._hook_codes[event_based_gateway.id][
            "pre_activate_next"
        ]
        when_triggered_code = self._hook_codes[event_based_gateway.id]["when_triggered"]
        code = snippet.EventBasedGateway_code(
            gateway=event_based_gateway.id,
            change_next_state_code="\n".join(
                [
                    self._generate_change_state_code(outgoing.target)
                    for outgoing in event_based_gateway.outgoings
                ]
            ),
            pre_activate_next_hook="\n\t".join(pre_activate_next_hook),
            after_all_hook="\n\t".join(when_triggered_code),
        )
        temp_list.append(code)
        return temp_list

    def _generate_chaincode_for_start_event(self, start_event: StartEvent):
        temp_list = []
        temp_list.append(
            snippet.StartEvent_code(
                start_event.id,
                change_next_state_code=self._generate_change_state_code(
                    start_event.outgoing.target
                ),
            )
        )
        return temp_list

    def _generate_chaincode_for_end_event(self, end_event: EndEvent):
        temp_list = []
        when_triggered_code = self._hook_codes[end_event.id]["when_triggered"]
        temp_list.append(
            snippet.EndEvent_code(
                end_event.id,
                after_all_hook="\n\t".join(when_triggered_code),
            )
        )
        return temp_list

    def generate_chaincode(
            self,
            bindings: dict[str, dict],
            output_path: str = "contract/chaincode/chaincode.go",
    ):
        self._hook_codes = {
            key: {"pre_activate_next": [], "when_triggered": []}
            for key in [node.id for node in self._choreography.nodes]
        }

        chaincode_list = []
        chaincode_list.append(snippet.package_code())
        chaincode_list.append(snippet.import_code())
        chaincode_list.append(snippet.contract_definition_code())
        chaincode_list.append(
            snippet.StateMemoryDefinition_code(self._generate_parameters_code())
        )
        chaincode_list.append(snippet.fix_part_code())
        chaincode_list.append(snippet.globale_variable_read_and_set_code())
        chaincode_list.extend(self._generate_InitLedger(bindings=bindings))

        for event_based_gateway in self._choreography.query_element_with_type(
                NodeType.EVENT_BASED_GATEWAY
        ):
            if len(event_based_gateway.outgoings) > 1:
                for outgoing in event_based_gateway.outgoings:
                    self._hook_codes[outgoing.target.id].setdefault(
                        "when_triggered", []
                    ).append(
                        self._event_based_gateway_hook_code(
                            event_based_gateway, outgoing.target
                        )
                    )
        for parallel_gateway in self._choreography.query_element_with_type(
                NodeType.PARALLEL_GATEWAY
        ):
            if len(parallel_gateway.incomings) > 1:
                for incoming in parallel_gateway.incomings:
                    self._hook_codes[incoming.id].setdefault(
                        "pre_activate_next", []
                    ).append(
                        self._parallel_gateway_merge_hook_code(
                            parallel_gateway, incoming.source
                        )
                    )

        for element in self._choreography.nodes:
            if element.type == NodeType.CHOREOGRAPHY_TASK:
                chaincode_list.extend(
                    self._generate_chaincode_for_choreography_task(element)
                )
            if element.type == NodeType.EXCLUSIVE_GATEWAY:
                chaincode_list.extend(
                    self._generate_chaincode_for_exclusive_gateway(element)
                )
            if element.type == NodeType.PARALLEL_GATEWAY:
                chaincode_list.extend(
                    self._generate_chaincode_for_parallel_gateway(element)
                )
            if element.type == NodeType.EVENT_BASED_GATEWAY:
                chaincode_list.extend(
                    self._generate_chaincode_for_event_based_gateway(element)
                )
            if element.type == NodeType.START_EVENT:
                chaincode_list.extend(self._generate_chaincode_for_start_event(element))
            if element.type == NodeType.END_EVENT:
                chaincode_list.extend(self._generate_chaincode_for_end_event(element))

        with open(output_path, "w", encoding="utf-8") as f:
            f.write("\n\n".join(chaincode_list))
        return "\n\n".join(chaincode_list)

    def get_participants(self):
        return {
            participant.id: participant.name
            for participant in self._choreography.query_element_with_type(
                NodeType.PARTICIPANT
            )
        }


if __name__ == "__main__":
    go_chaincode_translator = GoChaincodeTranslator(
        None, bpmn_file="bpmn/test.bpmn"
    )
    bindings = {
        "Participant_0nrc0gf": {
            "msp": "Org1MSP",
            "attributes": {},
        },
        "Participant_0vpw1po": {
            "msp": "Org2MSP",
            "attributes": {},
        }
    }
    go_chaincode_translator.generate_chaincode(bindings=bindings)
