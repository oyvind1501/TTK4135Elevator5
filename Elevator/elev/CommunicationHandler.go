package elev

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func NewOrderEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
	if nodeId == masterId {
		tableElement := createTableElement(message)
		if !isElementInHallTable(tableElement) {
			HallOrderTable = append(HallOrderTable, tableElement)
		}
		sendChannel <- ElevatorOrderMessage{
			Event:     EVENT_ACK_NEW_ORDER,
			Direction: message.Direction,
			Floor:     message.Floor,
			Origin:    message.Origin,
			Sender:    masterId,
		}
	}
	if nodeId == backupId {
		tableElement := createTableElement(message)
		if !isElementInHallTable(tableElement) {
			HallOrderTable = append(HallOrderTable, tableElement)
		}
	}
}

func AckNewOrderEvent(message ElevatorOrderMessage, lightChannel chan Light) {
	var lightButton int
	switch message.Direction {
	case DIR_UP:
		lightButton = BUTTON_HALL_UP
	case DIR_DOWN:
		lightButton = BUTTON_HALL_DOWN
	}
	lightChannel <- Light{
		LightType:   lightButton,
		LightOn:     true,
		FloorNumber: message.Floor,
	}
}

func OrderReserveEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
	if nodeId == masterId || nodeId == backupId {
		nextFloor := UNDEFINED

		var bestParticipantId string
		var participantId string

		participantFloor := -1
		shortestDistance := -1

		time.Sleep(100 * time.Millisecond)

		for index, participant := range ClientTable {
			participantId = participant.ClientInfo.Id
			participantFloor = participant.ClientInfo.Info.Floor
			closestOrder := ClosestFloor(participantFloor)

			if closestOrder == UNDEFINED {
				continue
			}
			if index == 0 {
				bestParticipantId = participantId
				shortestDistance = int(math.Abs(float64(participantFloor - closestOrder)))
				nextFloor = closestOrder
			} else {
				distance := int(math.Abs(float64(participantFloor - closestOrder)))

				if distance < shortestDistance {
					shortestDistance = distance
					bestParticipantId = participantId
					nextFloor = closestOrder
				}
			}

		}
		if bestParticipantId != message.Origin {
			return
		}
		isReserved := IsHallOrderReserved(nextFloor)
		if isReserved {
			return
		}

		SetOrderStatus(STATUS_OCCUPIED, message.Origin, nextFloor)
		if nodeId == masterId {
			sendChannel <- ElevatorOrderMessage{
				Event:      EVENT_ACK_ORDER_RESERVE,
				Floor:      nextFloor,
				AssignedTo: message.Origin,
				Origin:     message.Origin,
				Sender:     masterId,
			}
		}
	}
}

func AckOrderReserveEvent(message ElevatorOrderMessage) {
	if message.Origin == nodeId {
		if message.Floor != UNDEFINED {
			hallTarget = message.Floor
			doorOpened = false
			TargetFloor = message.Floor
		}
	}
}

func OrderReserveSpecificEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
	if nodeId == masterId {
		if IsOrderAt(message.Floor, message.Direction) {
			sendChannel <- ElevatorOrderMessage{
				Event:      EVENT_ACK_ORDER_RESERVE_SPECIFIC,
				Floor:      message.Floor,
				AssignedTo: message.Origin,
				Origin:     message.Origin,
				Sender:     masterId,
			}
		} else {
			sendChannel <- ElevatorOrderMessage{
				Event:      EVENT_ACK_ORDER_RESERVE_SPECIFIC,
				Floor:      UNDEFINED,
				AssignedTo: message.Origin,
				Origin:     message.Origin,
				Sender:     masterId,
			}
		}
	}
}

func AckOrderReserveSpecificEvent(message ElevatorOrderMessage) {
	if message.Origin == nodeId {
		if message.Floor != UNDEFINED {
			IsIntermediateStop = true
		} else {
			IsIntermediateStop = false
		}
	}
}

func OrderDoneEvent(message ElevatorOrderMessage, sendChannel chan ElevatorOrderMessage) {
	if nodeId == masterId {

		RemoveHallOrder(message.Floor)
		sendChannel <- ElevatorOrderMessage{
			Event:      EVENT_ACK_ORDER_DONE,
			Floor:      message.Floor,
			AssignedTo: message.Origin,
			Origin:     message.Origin,
			Sender:     masterId,
		}
	}
	if nodeId == backupId && message.Floor != UNDEFINED {
		RemoveHallOrder(message.Floor)
	}
}

func AckOrderDoneEvent(message ElevatorOrderMessage, lightChannel chan Light) {
	lightChannel <- Light{
		LightType:   BUTTON_HALL_UP,
		LightOn:     false,
		FloorNumber: message.Floor,
	}
	lightChannel <- Light{
		LightType:   BUTTON_HALL_DOWN,
		LightOn:     false,
		FloorNumber: message.Floor,
	}
	if message.Origin == nodeId && doorOpened == false {
		open = true
	}
}

func createTableElement(message ElevatorOrderMessage) HallOrderElement {
	tableElement := HallOrderElement{
		Command:   message.Event,
		Direction: message.Direction,
		Floor:     message.Floor,
		ReserveID: "RESERVER_UNDEFINED",
		Status:    STATUS_AVAILABLE,
	}
	return tableElement
}

func isElementInHallTable(element HallOrderElement) bool {
	for _, tableElement := range HallOrderTable {
		if isTableElementEqual(element, tableElement) {
			return true
		}
	}
	return false
}

func isTableElementEqual(element HallOrderElement, tableElement HallOrderElement) bool {
	if element.Command == tableElement.Command && element.Direction == tableElement.Direction && element.Floor == tableElement.Floor {
		return true
	}
	return false
}

func UpdateReservationTable(sendChannel chan ElevatorOrderMessage) {
	if len(CabOrderTable) != 0 {
		for _, cabOrder := range CabOrderTable {
			ReserveTable = append(ReserveTable, ReserveElement{Floor: cabOrder.Floor})
			removeCabOrder(cabOrder)
		}
	}
	sendChannel <- ElevatorOrderMessage{
		Event:     EVENT_ORDER_RESERVE_SPECIFIC,
		Direction: ElevatorDirection,
		Floor:     LastFloor,
		Origin:    nodeId,
		Sender:    nodeId,
	}
}

func CheckForOrders(sendChannel chan ElevatorOrderMessage) {
	if len(CabOrderTable) != 0 {
		for _, cabOrder := range CabOrderTable {
			TargetFloor = GetCabOrder(LastFloor, ElevatorDirection)
			removeCabOrder(cabOrder)
			break
		}
	} else {
		sendChannel <- ElevatorOrderMessage{
			Event:     EVENT_ORDER_RESERVE,
			Direction: ElevatorDirection,
			Floor:     LastFloor,
			Origin:    nodeId,
			Sender:    nodeId,
		}
	}
}

func removeCabOrder(cabOrder CabOrderElement) {
	for index, element := range CabOrderTable {
		if element == cabOrder {
			CabOrderTable = append(CabOrderTable[:index], CabOrderTable[index+1:]...)
		}
	}
}

func PrintHallTable() {
	fmt.Println("-----------------Hall Order Table:----------------------")
	if len(HallOrderTable) == 0 {
		fmt.Println("	No hall")
	} else {
		for _, tableElement := range HallOrderTable {
			switch tableElement.Direction {
			case DIR_UP:
				if tableElement.Status == STATUS_AVAILABLE {
					fmt.Println(string(tableElement.Command) + " " + "UP" + " " + strconv.Itoa(tableElement.Floor) + " " + strconv.Itoa(tableElement.TimeReserved.Minute()) + ":" + strconv.Itoa(tableElement.TimeReserved.Second()) + " " + tableElement.ReserveID + " " + "AVAILABLE")
				}
				if tableElement.Status == STATUS_OCCUPIED {
					fmt.Println(string(tableElement.Command) + " " + "UP" + " " + strconv.Itoa(tableElement.Floor) + " " + strconv.Itoa(tableElement.TimeReserved.Minute()) + ":" + strconv.Itoa(tableElement.TimeReserved.Second()) + " " + tableElement.ReserveID + " " + "OCCUPIED")
				}

			case DIR_DOWN:
				if tableElement.Status == STATUS_AVAILABLE {
					fmt.Println(string(tableElement.Command) + " " + "DOWN" + " " + strconv.Itoa(tableElement.Floor) + " " + strconv.Itoa(tableElement.TimeReserved.Minute()) + ":" + strconv.Itoa(tableElement.TimeReserved.Second()) + " " + tableElement.ReserveID + " " + "AVAILABLE")
				}
				if tableElement.Status == STATUS_OCCUPIED {
					fmt.Println(string(tableElement.Command) + " " + "DOWN" + " " + strconv.Itoa(tableElement.Floor) + " " + strconv.Itoa(tableElement.TimeReserved.Minute()) + ":" + strconv.Itoa(tableElement.TimeReserved.Second()) + " " + tableElement.ReserveID + " " + "OCCUPIED")
				}
			}
		}
	}
	fmt.Println("--------------------------------------------------------")
}

func PrintCabTable() {
	fmt.Println("-----------------Cab Order Table------------------------")
	if len(CabOrderTable) == 0 {
		fmt.Println("	No cab")
	} else {
		for _, tableElement := range CabOrderTable {
			fmt.Println("	Order at Floor:\t\t" + strconv.Itoa(tableElement.Floor))
		}
	}
	fmt.Println("--------------------------------------------------------")
}

func PrintCommunicationTable() {
	var nodeType string
	var nodeTypeId string

	for _, clients := range ClientTable {
		if clients.ClientInfo.Id == masterId {
			nodeType = "Master"
			nodeTypeId = clients.ClientInfo.Id
		}
		if clients.ClientInfo.Id == backupId {
			nodeType = "Backup"
			nodeTypeId = clients.ClientInfo.Id
		}
		if clients.ClientInfo.Id != masterId && clients.ClientInfo.Id != backupId {
			nodeType = "Node"
			nodeTypeId = clients.ClientInfo.Id
		}
	}
	fmt.Println("----------------Node Network Information----------------")
	fmt.Println("	Type:		", nodeType)
	fmt.Println("	Id:		", nodeTypeId)
	fmt.Println("--------------------------------------------------------")
}

func PrintStateInfo() {
	fmt.Println("-----------------Elevator Info--------------------------")
	fmt.Println("	Target floor:\t\t" + strconv.Itoa(TargetFloor))
	fmt.Println("	Last floor: \t\t" + strconv.Itoa(LastFloor))
	switch ElevatorDirection {
	case DIR_STOP:
		fmt.Println("	Direction:\t\tDIR_STOP")
	case DIR_UP:
		fmt.Println("	Direction:\t\tDIR_UP")
	case DIR_DOWN:
		fmt.Println("	Direction:\t\tDIR_DOWN")
	}
	fmt.Println("--------------------------------------------------------")
}

func PrintElevatorInfo() {
	for {
		PrintCommunicationTable()
		fmt.Println()
		PrintStateInfo()
		fmt.Println()
		PrintCabTable()
		fmt.Println()
		PrintHallTable()
		fmt.Println()

		time.Sleep(500 * time.Millisecond)
	}
}
