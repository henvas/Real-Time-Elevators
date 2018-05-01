package ElevController

import (
	"../driver-go-master/elevio"
	"fmt"
	"../Queue"
	"../Com"
	"../ReqAssigner"
	"time"
)

// The local Elevator
var Elevator Com.Elev
// Holds remote hall orders
var RemoteElev Com.Elev
var LivingPeers []string

func ElevController(id string,
					FloorCh chan int,
					ButtonCh chan elevio.ButtonEvent,
					drv_stopCh <-chan bool,
					TxOrderCh chan Com.Message,
					doorTimerCh <-chan time.Time,
					peerTxEnable chan bool,
					) {

	Elevator.Floor = elevio.GetFloor()
	ElevInit(&Elevator, id)
	for{

		// Update the direction of the elevator
		if Elevator.State != Com.INITIALIZE {
			if Elevator.State == Com.DOORSOPEN {
				Elevator.Direction = elevio.MD_Stop
			}else {
				Elevator.Direction = Queue.SetDirection(Elevator)
			}
			elevio.SetMotorDirection(Elevator.Direction)
			if Elevator.Direction != elevio.MD_Stop {
				Elevator.State = Com.MOVING
			}
		}

		select{
		case button := <- ButtonCh:
			switch Elevator.State {
			case Com.INITIALIZE:
				break

			case Com.DOORSOPEN:
				if Elevator.Floor == button.Floor {
					elevio.SetButtonLamp(button.Button,button.Floor,false)
					elevio.SetDoorOpenLamp(true)
					doorTimerCh = time.After(Com.DoorTimer)
				} else if button.Button == elevio.BT_Cab {
					Queue.SetQueueReq(&Elevator,button)
					elevio.SetButtonLamp(button.Button,button.Floor,true)
				} else {
					RemoteElev = ReqAssigner.ClosestElevator(button)
					Queue.SetQueueReq(&RemoteElev,button)
					OrderMessage := Com.Message{Elevator:RemoteElev, ID:RemoteElev.ID}
					TxOrderCh <- OrderMessage
				}
				break

			case Com.IDLE:
				if Elevator.Floor == button.Floor {
					elevio.SetButtonLamp(button.Button,button.Floor,false)
					elevio.SetDoorOpenLamp(true)
					Elevator.State = Com.DOORSOPEN
					doorTimerCh = time.After(Com.DoorTimer)
				} else if button.Button == elevio.BT_Cab {
					Queue.SetQueueReq(&Elevator,button)
					elevio.SetButtonLamp(button.Button,button.Floor,true)
					Elevator.Direction = Queue.SetDirection(Elevator)
					elevio.SetMotorDirection(Elevator.Direction)
					if Elevator.Direction != elevio.MD_Stop {
						Elevator.State = Com.MOVING
					}

				} else {
					RemoteElev = ReqAssigner.ClosestElevator(button)
					Queue.SetQueueReq(&RemoteElev,button)
					OrderMessage := Com.Message{Elevator:RemoteElev, ID:RemoteElev.ID}
					TxOrderCh <- OrderMessage
				}
				break

			case Com.MOVING:
				if button.Button == elevio.BT_Cab {
					Queue.SetQueueReq(&Elevator,button)
					elevio.SetButtonLamp(button.Button,button.Floor,true)
				} else {
					RemoteElev = ReqAssigner.ClosestElevator(button)
					Queue.SetQueueReq(&RemoteElev,button)
					OrderMessage := Com.Message{Elevator:RemoteElev, ID:RemoteElev.ID}
					TxOrderCh <- OrderMessage
				}
				break
			}

		case arriveFloor := <- FloorCh:
			fmt.Printf("%+v\n", arriveFloor)
			Elevator.Floor = arriveFloor
			elevio.SetFloorIndicator(arriveFloor)
			switch Elevator.State {
			case Com.INITIALIZE:
				if arriveFloor == 0 {
					fmt.Println("Elevator Initialized at floor 0")
					elevio.SetMotorDirection(elevio.MD_Stop)
					Elevator.State = Com.IDLE
					Elevator.Direction = elevio.MD_Stop
					break
				}

			case Com.IDLE:
				fmt.Println("Arrive at floor while idle(error)")
				break

			case Com.MOVING:
				fmt.Println("Arrive at floor while moving")
				if Queue.ShouldStop(Elevator) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					Queue.ClearReqsAtFloor(&Elevator, arriveFloor)
					elevio.SetDoorOpenLamp(true)
					Elevator.State = Com.DOORSOPEN
					doorTimerCh = time.After(Com.DoorTimer)
					break
				}

			case Com.DOORSOPEN:
				fmt.Println("Arrive at floor while dooropen(error)")
				break
			}

		case <- doorTimerCh:
			fmt.Println("Door timer activated with State ", Elevator.State)
			switch Elevator.State {
			case Com.INITIALIZE:
				break

			case Com.IDLE:
				break

			case Com.MOVING:
				break

			case Com.DOORSOPEN:
				elevio.SetDoorOpenLamp(false)
				Elevator.Direction = Queue.SetDirection(Elevator)
				if Elevator.Direction != elevio.MD_Stop {
					Elevator.State = Com.MOVING
				}else {
					Elevator.State = Com.IDLE
				}
				elevio.SetMotorDirection(Elevator.Direction)
				break
			}

		case stopButton := <- drv_stopCh:
			fmt.Printf("%+v\n", stopButton)
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetStopLamp(stopButton)

		case <-time.After(15*time.Second):
			if Elevator.State != Com.IDLE && Elevator.State != Com.DOORSOPEN {
				peerTxEnable <- false
				elevio.SetMotorDirection(elevio.MD_Stop)
				time.Sleep(20*time.Second)
				peerTxEnable <- true
				elevio.SetMotorDirection(Elevator.Direction)
			}
		}
	}

}

func ElevInit(elev *Com.Elev, id string) {
	for i := 0; i < Com.NumFloors; i++ {
		Queue.ClearReqsAtFloor(elev,i)
	}
	fmt.Println("Initialize elevator")
	if elev.Floor != 0 {
		elevio.SetMotorDirection(elevio.MD_Down)
	}
	elev.State = Com.INITIALIZE
	elev.ID = id
}
