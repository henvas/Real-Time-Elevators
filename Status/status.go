package Status

import (
	"../Com"
	"../Queue"
	"../driver-go-master/elevio"
	fsm "../ElevController"
	"../Network-go-master/network/peers"
	"time"
	"fmt"
)

//var ElevMap Com.ElevatorMap
type HeartBeat struct {
	Ticker *time.Ticker
}

func (tick *HeartBeat)Status(id string,
			RxOrderCh chan Com.Message,
			peerUpdateCh chan peers.PeerUpdate,
			doorTimerCh <-chan time.Time,
			updateTxElevatorCh chan<- Com.Message,
			updateRxElevatorCh <-chan Com.Message,
			) {

	Com.ElevMap = make(Com.ElevatorMap)

	if tick.Ticker == nil {
		tick.Stop()
	}
	for {
		select {
		// Received an order (closest elevator) and check if we should take it
		case RemoteOrder := <- RxOrderCh:
			if RemoteOrder.Elevator.ID == id {
				fmt.Println("This is the closest elevator")
				Queue.SetRemoteQueueReq(&fsm.Elevator,RemoteOrder.Elevator)
				HallLampsThisElev(fsm.Elevator)
				// If elevator already at floor
				if Queue.ShouldStop(fsm.Elevator) && fsm.Elevator.Direction == elevio.MD_Stop {
					elevio.SetMotorDirection(elevio.MD_Stop)
					Queue.ClearReqsAtFloor(&fsm.Elevator, fsm.Elevator.Floor)
					elevio.SetDoorOpenLamp(true)
					fsm.Elevator.State = Com.DOORSOPEN
					doorTimerCh = time.After(Com.DoorTimer)
				} else {
					if fsm.Elevator.State != Com.INITIALIZE {
						if fsm.Elevator.State == Com.DOORSOPEN {
							fsm.Elevator.Direction = elevio.MD_Stop
						} else {
							fsm.Elevator.Direction = Queue.SetDirection(fsm.Elevator)
						}
						elevio.SetMotorDirection(fsm.Elevator.Direction)
						if fsm.Elevator.Direction != elevio.MD_Stop {
							fsm.Elevator.State = Com.MOVING
						}
					}
				}
			}

		case p := <-peerUpdateCh:
			if len(Com.ElevMap) == 0 {
				InitElevMap(p)
			}

			fsm.LivingPeers = p.Peers

			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			// Check if elevator is lost
			if len(p.Lost) > 0 {
				for lostPeer := 0; lostPeer < len(p.Lost);lostPeer++ {
					for id, elev := range Com.ElevMap {
						// Found the elevator in the map
						if id == p.Lost[lostPeer] {
							// Choose the first elevator in LivingPeers
							if fsm.Elevator.ID == fsm.LivingPeers[0] {
								Queue.SetRemoteQueueReq(&fsm.Elevator,elev)
								HallLampsThisElev(fsm.Elevator)
								// If elevator already at floor
								if Queue.ShouldStop(fsm.Elevator) && fsm.Elevator.Direction == elevio.MD_Stop {
									elevio.SetMotorDirection(elevio.MD_Stop)
									Queue.ClearReqsAtFloor(&fsm.Elevator, fsm.Elevator.Floor)
									elevio.SetDoorOpenLamp(true)
									fsm.Elevator.State = Com.DOORSOPEN
									doorTimerCh = time.After(Com.DoorTimer)
								} else {
									if fsm.Elevator.State != Com.INITIALIZE {
										if fsm.Elevator.State == Com.DOORSOPEN {
											fsm.Elevator.Direction = elevio.MD_Stop
										} else {
											fsm.Elevator.Direction = Queue.SetDirection(fsm.Elevator)
										}
										elevio.SetMotorDirection(fsm.Elevator.Direction)
										if fsm.Elevator.Direction != elevio.MD_Stop {
											fsm.Elevator.State = Com.MOVING
										}
									}
								}
							}
							
						}
					}
				}
				// Remove lost elevator/peer from map
				for _,lost := range p.Lost {
					delete(Com.ElevMap,lost)
				}
			}

		// Sending elevators message to all other elevators when heartbeat ticks
		case <- tick.Ticker.C:
			msgTx := Com.Message{ID:fsm.Elevator.ID,Elevator:fsm.Elevator}
			updateTxElevatorCh <- msgTx

		// Updating elevators when receiving heartbeat tick
		case msgRx := <-updateRxElevatorCh:
			Com.ElevMap[msgRx.Elevator.ID] = msgRx.Elevator

		case <- doorTimerCh:
			fmt.Println("Door timer activated with State ", fsm.Elevator.State)
			switch fsm.Elevator.State {
			case Com.INITIALIZE:
				break

			case Com.IDLE:
				break

			case Com.MOVING:
				break

			case Com.DOORSOPEN:
				elevio.SetDoorOpenLamp(false)
				fsm.Elevator.Direction = Queue.SetDirection(fsm.Elevator)
				if fsm.Elevator.Direction != elevio.MD_Stop {
					fsm.Elevator.State = Com.MOVING
				} else {
					fsm.Elevator.State = Com.IDLE
				}
				elevio.SetMotorDirection(fsm.Elevator.Direction)
				break
			}
		}
	}
}

func InitElevMap(p peers.PeerUpdate) {
	fmt.Println("Init elevator map")
	requests := [Com.NumFloors][Com.NumButtons]bool{}
	Com.ElevMap[p.New] = Com.Elev {
		ID:        p.New,
		State:     Com.INITIALIZE,
		Floor:     0,
		Direction: elevio.MD_Stop,
		Requests:  requests,
	}
}

func (tick *HeartBeat)Stop() {
	if tick.Ticker != nil {
		tick.Ticker.Stop()
		tick.Ticker = nil
	}
}

// Update hall lamps for the elevator with the request
func HallLampsThisElev(elev Com.Elev) {
	for floor := 0;floor < Com.NumFloors; floor++ {
		if elev.Requests[floor][elevio.BT_HallUp] {
				elevio.SetButtonLamp(elevio.BT_HallUp,floor,true)
			}
			if elev.Requests[floor][elevio.BT_HallDown] {
				elevio.SetButtonLamp(elevio.BT_HallDown,floor,true)
			}
	}
}


