package main_driver

import "./elevio"
import "fmt"
import c "../ElevController"
import q "../Queue"


func main(){

    numFloors := 4

    elevio.Init("localhost:15657", numFloors)

    var floor1 bool = false
    
    var d elevio.MotorDirection = elevio.MD_Up
    //elevio.SetMotorDirection(d)
    
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)    
    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)
    
    for {
        select {
        case a := <- drv_buttons:
            fmt.Printf("%+v\n", a)
            elevio.SetButtonLamp(a.Button, a.Floor, true)
            q.SetQueueReq(&c.RemoteElev,a)
            q.SetRemoteQueueReq(&c.Elevator,c.RemoteElev)
            fmt.Println("Remote elevator queue:")
            q.PrintQueue(c.RemoteElev)
            q.PrintQueue(c.Elevator)
            if a.Floor == 0 {
                floor1 = true
            }
            
        case a := <- drv_floors:
            fmt.Printf("%+v\n", a)
            if a == numFloors-1 {
                d = elevio.MD_Down
            } else if a == 0 {
                d = elevio.MD_Up
            }
            if a == 0 && floor1 {
                floor1 = false
                elevio.SetButtonLamp(0, 0, false)
            }
            elevio.SetMotorDirection(d)
            
            
        case a := <- drv_obstr:
            fmt.Printf("%+v\n", a)
            if a {
                elevio.SetMotorDirection(elevio.MD_Stop)
            } else {
                elevio.SetMotorDirection(d)
            }
            
        case a := <- drv_stop:
            fmt.Printf("%+v\n", a)
            for f := 0; f < numFloors; f++ {
                for b := elevio.ButtonType(0); b < 3; b++ {
                    elevio.SetButtonLamp(b, f, false)
                    elevio.SetMotorDirection(elevio.MD_Stop)
                }
            }
        }
    }    
}
