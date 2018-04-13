package net

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"../elev"
)

type HallOrderElement struct {
	Command   elev.ActionCommand
	Direction string
	Floor     int
}

var HallOrderTable []HallOrderElement

/*-----------------------------------------------------
Function:		ReceiveHandler
Arguments:	receiveChannel chan elev.Action
Affected:		HallOrderTable
ReceiveHandler is activated when the receiveChannel != nil.
The function checks the type of command sent to the server,
and uses this information to update the table of HallOrders.
-----------------------------------------------------*/
func ReceiveHandler(receiveChannel chan elev.Action) {
	for {
		select {
		case command := <-receiveChannel:
			switch command.Command {
			case elev.COMMAND_ORDER:
				addToTable(command)
			case elev.COMMAND_ORDER_DONE:
				removeOrder(command)
			}
		}
	}
}

/*-----------------------------------------------------
Function:		addToTable
Arguments:	elev.Action
Affected:		HallOrderTable
Adds the content of the command to the table of hall orders.
-----------------------------------------------------*/
func addToTable(command elev.Action) {
	tableElement := decodeCommand(command)
	if !isElementInHallTable(tableElement) {
		HallOrderTable = append(HallOrderTable, tableElement)
	}
}

/*-----------------------------------------------------
Function:		removeOrder
Arguments:	elev.Action
Affected:		HallOrderTable
Removes content equal, to the content of the argument,
from the hall order table.
-----------------------------------------------------*/
func removeOrder(command elev.Action) {
	tableElement := decodeCommand(command)
	for index, element := range HallOrderTable {
		if element == tableElement {
			HallOrderTable = append(HallOrderTable[:index], HallOrderTable[index+1:]...)
			break
		}
	}
}

/*-----------------------------------------------------
Function:		isElementInHallTable
Arguments:	HallOrderElement
Returns:		bool
Affected:		None
Checks if the element passed as an argument is in the table
of hall orders.
-----------------------------------------------------*/
func isElementInHallTable(tableElement HallOrderElement) bool {
	for _, element := range HallOrderTable {
		if element == tableElement {
			return true
		}
	}
	return false
}

/*-----------------------------------------------------
Function:		removeOrder
Arguments:	elev.Action
Affected:		HallOrderTable
Decodes the elev.Action argument passed,
into a usable element for the hall order table.
-----------------------------------------------------*/
func decodeCommand(command elev.Action) HallOrderElement {
	parameterList := strings.Split(command.Parameters, " ")
	direction := parameterList[0]
	floor, _ := strconv.Atoi(parameterList[1])

	return HallOrderElement{Command: command.Command, Direction: direction, Floor: floor}
}

// Only for testing purposes!
func PrintHallTable() {
	for {
		for _, tableElement := range HallOrderTable {
			fmt.Println("Hall-Table: " + string(tableElement.Command) + " " + tableElement.Direction + " " + strconv.Itoa(tableElement.Floor))
		}
		time.Sleep(1 * time.Second)
	}
}
