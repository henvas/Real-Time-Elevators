package ReqAssigner

import (
	"../Com"
	"../driver-go-master/elevio"
	"../Queue"
	"fmt"
	"sort"
)

func CostFunction(elev Com.Elev, button elevio.ButtonEvent) int {
	cost := 0
	distance := distance(button.Floor,elev.Floor)
	dirToDest := directionToDestination(elev, distance)

	// Elevator moving in wrong direction
	if elev.State == Com.MOVING && dirToDest != elev.Direction {
		cost += 20
	}
	cost += abs(distance)
	// Adds 2 for each stop on the way
	cost += 2*checkAmountOfStops(elev, button, dirToDest)

	return cost
}

func ClosestElevator(button elevio.ButtonEvent) Com.Elev {
	CostMap := make(map[string]int)
	var costs []int
	var closestElev Com.Elev
	// Calculate cost for each elevator and add to cost map
	for _,elev := range Com.ElevMap {
		CostMap[elev.ID] = CostFunction(elev,button)
	}
	for _,value := range CostMap {
		costs = append(costs,value)
	}
	// Sort cost with increasing values
	sort.Ints(costs)
	for key,value := range CostMap {
		if value == costs[0] {
			for ip, elevs := range Com.ElevMap {
				if ip == key {
					closestElev = elevs
				}
			}
		}
	}
	fmt.Println("The closest elevator was chosen with ID: ", closestElev.ID)
	return closestElev
}

func distance(destination,currentFloor int) int {
	return destination - currentFloor
}

func directionToDestination(elev Com.Elev, distance int) elevio.MotorDirection {
	// If the elevator has just started moving from the desired floor
	if distance == 0 && elev.State == Com.MOVING {
		if elev.Direction == elevio.MD_Up {
			return elevio.MD_Down
		} else {
			return elevio.MD_Up
		}
	}
	if distance > 0 {
		return elevio.MD_Up
	} else {
		return elevio.MD_Down
	}
}

func checkAmountOfStops(elev Com.Elev, button elevio.ButtonEvent, directon elevio.MotorDirection) int {
	stops := 0
	if Queue.CheckBelow(elev) && Queue.CheckAbove(elev) && elev.Floor != button.Floor {
		return stops
	}
	if directon == elevio.MD_Up {
		for f := elev.Floor; f < button.Floor; f++ {
			if elev.Requests[f][button.Button] {
				stops++
			}
		}
	}
	if directon == elevio.MD_Down {
		for f := elev.Floor; f > button.Floor; f-- {
			if elev.Requests[f][button.Button] {
				stops++
			}
		}
	}
	return stops
}

func abs(distance int) int {
	if distance >= 0 {
		return distance
	} else {
		return -distance
	}
}