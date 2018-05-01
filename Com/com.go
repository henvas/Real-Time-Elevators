package Com

import (
	"../driver-go-master/elevio"
	"time"
)

const NumFloors = 4
const NumButtons = 3
var DoorTimer = time.Second * 3
var HeartBeatTimer = time.Millisecond * 50

type States int

const (
	INITIALIZE States = iota
	IDLE
	MOVING
	DOORSOPEN
)

type ElevatorMap map[string]Elev

type Elev struct {
	ID 				string
	State 			States
	Floor 			int
	Direction 		elevio.MotorDirection
	Requests 		[NumFloors][NumButtons]bool
}

type Message struct {
	Elevator 		Elev
	ID 				string
}

var ElevMap ElevatorMap
