package Queue

import (
	"../driver-go-master/elevio"
	"../Com"
	"fmt"
)

func SetQueueReq(elev *Com.Elev, button elevio.ButtonEvent) {
	elev.Requests[button.Floor][button.Button] = true
}

// Copy hall requests from remote elevator to local elevator
func SetRemoteQueueReq(LocalElev *Com.Elev, RemoteElev Com.Elev) {
	for floor := 0; floor < Com.NumFloors; floor++ {
		if RemoteElev.Requests[floor][elevio.BT_HallUp] {
			LocalElev.Requests[floor][elevio.BT_HallUp] = true
		}
		if RemoteElev.Requests[floor][elevio.BT_HallDown] {
			LocalElev.Requests[floor][elevio.BT_HallDown] = true
		}
	}
}

func CheckAbove(elev Com.Elev) bool {
	if elev.Floor == 3 {
		return false
	}
	for i := elev.Floor + 1; i < Com.NumFloors; i++ {
		for j := 0; j < Com.NumButtons; j++ {
			if elev.Requests[i][j] == true {
				return true
			}
		}
	}
	return false
}

func CheckBelow(elev Com.Elev) bool {
	if elev.Floor == 0 {
		return false
	}
	for i := 0; i < elev.Floor; i++ {
		for j := 0; j < Com.NumButtons; j++ {
			if elev.Requests[i][j] == true {
				return true
			}
		}
	}
	return false
}

func SetDirection(elev Com.Elev) elevio.MotorDirection {
	getUp := CheckAbove(elev)
	getDown := CheckBelow(elev)

	if getUp == true && getDown == false {
		return elevio.MD_Up
	} else if getUp == false && getDown == true {
		return elevio.MD_Down
	}

	if getDown == true && elev.Direction == elevio.MD_Down {
		return elevio.MD_Down
	}
	if getUp == true && (elev.Direction == elevio.MD_Up || getDown == true) {
		return elevio.MD_Up
	}
	return elevio.MD_Stop
}

func ShouldStop(elev Com.Elev) bool {
	if elev.Direction == elevio.MD_Up {
		return 	elev.Requests[elev.Floor][elevio.BT_HallUp] 	||
				elev.Requests[elev.Floor][elevio.BT_Cab] 		||
				!CheckAbove(elev)
	}
	if elev.Direction == elevio.MD_Down {
		return 	elev.Requests[elev.Floor][elevio.BT_HallDown] 	||
				elev.Requests[elev.Floor][elevio.BT_Cab] 		||
				!CheckBelow(elev)
	}
	if elev.Direction == elevio.MD_Stop {
		for button := 0; button < Com.NumButtons;button++ {
			if elev.Requests[elev.Floor][button] {
				return true
			}
		}
	}
	return false
}

func ClearReqsAtFloor(elev *Com.Elev,floor int) {
	elev.Requests[floor][elevio.BT_HallDown] = false
	elev.Requests[floor][elevio.BT_HallUp] = false
	elev.Requests[floor][elevio.BT_Cab] = false
	elevio.SetButtonLamp(elevio.BT_Cab,floor,false)
	if(floor != 3) {
		elevio.SetButtonLamp(elevio.BT_HallUp,floor,false)
	}
	if(floor != 0) {
		elevio.SetButtonLamp(elevio.BT_HallDown,floor,false)
	}
}

func PrintQueue(elev Com.Elev) {
	fmt.Println("Queue matrix: ")
	for f := 0; f < Com.NumFloors; f++ {
		for b := 0; b < Com.NumButtons; b++ {
			if elev.Requests[f][b] {
				fmt.Printf("1 ")
			}
			if !elev.Requests[f][b] {
				fmt.Printf("0 ")
			}
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}