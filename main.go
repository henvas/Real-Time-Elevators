package main

import (
	"flag"
	"fmt"
	"time"
	"../project-askar-test/driver-go-master/elevio"
	"../project-askar-test/ElevController"
	"../project-askar-test/Network-go-master/network/bcast"
	"../project-askar-test/Network-go-master/network/peers"
	"../project-askar-test/Com"
	"../project-askar-test/Status"
	"os"
	"runtime"
	"os/signal"
)

func main() {
	numFloors := 4
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()
	port := os.Args[2]

	fmt.Println("Elevator ID: ", id)
	runtime.GOMAXPROCS(runtime.NumCPU())

	elevio.Init("localhost:"+port, numFloors)

	// Heartbeat ticker
	tick := Status.HeartBeat{Ticker:time.NewTicker(Com.HeartBeatTimer)}

	// Driver channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors  := make(chan int)
	drv_stop    := make(chan bool)
	// Timer channels
	doorTimerCh := make(<-chan time.Time)
	// Network channels
	TxOrderCh 	:= make(chan Com.Message)
	RxOrderCh 	:= make(chan Com.Message)
	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	updateTxElevatorCh := make(chan Com.Message)
	updateRxElevatorCh := make(chan Com.Message)

	// Driver functions
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollStopButton(drv_stop)
	// Network functions
	go bcast.Transmitter(33330, TxOrderCh)
	go bcast.Receiver(33330, RxOrderCh)
	go bcast.Transmitter(33331,updateTxElevatorCh)
	go bcast.Receiver(33331,updateRxElevatorCh)
	go peers.Transmitter(33332,id,peerTxEnable)
	go peers.Receiver(33332,peerUpdateCh)
	// Status function
	go tick.Status(id,RxOrderCh, peerUpdateCh,doorTimerCh,updateTxElevatorCh,updateRxElevatorCh)
	// Elevator state machine
	go ElevController.ElevController(id,drv_floors,drv_buttons,drv_stop,TxOrderCh, doorTimerCh,peerTxEnable)

	c := make(chan os.Signal)
	signal.Notify(c,os.Interrupt)

	fmt.Println("All goroutines active")
	select {
	case <-c:
		fmt.Println("Program terminated by interrupt")
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
}
