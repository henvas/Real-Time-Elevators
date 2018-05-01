# Elevator project

In this submission peer-to-peer (P2P) networking is used, where all elevators (peers) have the same information. Furthermore, if an elevator with an order is disconnected from the network, its order is reassigned to another elevator. Hall orders are assigned to elevators through a cost function, which calculates cost based on the amount of stops and distance to desired location.

### Status
In this module remote orders are assigned, and the elevators states are updated. 
Also included is a ticker/heartbeat. Every 50 milliseconds a heartbeat is sent, and the elevator map updates. The map is defined in *com.go* as ElevatorMap, and includes the ID, state, floor, direction and requests of all the elevators.
When a hall button is pressed, this module receives the order message, and assigns it to the local elevator.
Under the event of an elevator disconnecting, its hall requests are assigned to any other live elevator.
Also, if an elevator is added, the elevator map is updated.

### ElevController
This module contains events and states, and it is here events are handled. First, the elevators initialize to first floor, then each elevator performs an action based on the event and its state. For cab buttons the orders are executed normally, while for hall buttons a cost function is calculated. The cost function CostFunction is defined in *reqassigner.go*, and is used in ClosestElevator. This function returns the closest elevator with the lowest cost. When the elevator is chosen, it is transmitted into the order channel, and the status file receives orders and checks if the order shall be executed.

 - RemoteElev - for hall orders. After choosing which elevator should execute an order, the order is copied into an Elevator.
 - Elevator - the local elevators which actually execute orders.
