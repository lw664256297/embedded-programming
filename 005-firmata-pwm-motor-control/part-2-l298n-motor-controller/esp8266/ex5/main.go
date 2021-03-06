package main

// Circuit: esp8266-and-l298n-motor-controller
// Objective: dual speed and direction control using MotorDriver
//
// | Enable | Dir 1 | Dir 2 | Motor         |
// +--------+-------+-------+---------------+
// | 0      | X     | X     | Off           |
// | 1      | 0     | 0     | 0ff           |
// | 1      | 0     | 1     | On (forward)  |
// | 1      | 1     | 0     | On (backward) |
// | 1      | 1     | 1     | Off           |

import (
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/firmata"
)

const (
	defaultPort = "10.10.100.175:3030"
)

/*

URL: https://tronixlabs.com.au/news/tutorial-l298n-dual-motor-controller-module-2a-and-arduino/

Motor Shield  | NodeMCU        | GPIO  | Purpose
--------------+----------------+-------+----------
A-Enable      | PWMA (Motor A) | 14	   | Speed
A-Dir1        | DIR1 (Motor A) | 5	   | Direction
A-Dir2        | DIR2 (Motor A) | 4	   | Direction
B-Enable      | PWMA (Motor B) | 12	   | Speed
B-Dir1        | DIR1 (Motor B) | 0	   | Direction
B-Dir2        | DIR2 (Motor B) | 2	   | Direction

*/
const (
	maPWMPin  = "14"
	maDir1Pin = "5"
	maDir2Pin = "4"
	mbPWMPin  = "12"
	mbDir1Pin = "0"
	mbDir2Pin = "2"
)

const (
	maIndex = iota
	mbIndex
)

var (
	motorSpeed [2]byte
	motorInc   = [2]int{1, 1}
	counter    = [2]int{}
	motors     [2]*gpio.MotorDriver
)

func main() {
	flag.Parse()

	port := defaultPort

	if len(flag.Args()) == 1 {
		port = flag.Args()[0]
	}

	log.Infof("Using port %v\n", port)

	board1 := firmata.NewTCPAdaptor(port)
	motorA := gpio.NewMotorDriver(board1, maPWMPin)
	motorA.ForwardPin = maDir1Pin
	motorA.BackwardPin = maDir2Pin
	motorA.SetName("Motor-A")
	motorB := gpio.NewMotorDriver(board1, mbPWMPin)
	motorB.ForwardPin = mbDir1Pin
	motorB.BackwardPin = mbDir2Pin
	motorB.SetName("Motor-B")

	motors[maIndex] = motorA
	motors[mbIndex] = motorB

	work := func() {
		motorA.Direction("forward")
		motorB.Direction("backward")
		
		gobot.Every(40*time.Millisecond, func() {
			motorControl(maIndex)
		})

		gobot.Every(20*time.Millisecond, func() {
			motorControl(mbIndex)
		})
	}

	robot := gobot.NewRobot("my-robot",
		[]gobot.Connection{board1},
		[]gobot.Device{motorA, motorB},
		work,
	)

	robot.Start()
}

func motorControl(idx int) {
	m := motors[idx]

	motorSpeed[idx] = byte(int(motorSpeed[idx]) + motorInc[idx])
	// log.Infof("Setting %v speed to %v\n", m.Name(), motorSpeed[idx])
	m.Speed(motorSpeed[idx])

	counter[idx]++
	if counter[idx]%256 == 255 {
		if motorInc[idx] == 1 {
			motorInc[idx] = 0
		} else if motorInc[idx] == 0 {
			motorInc[idx] = -1
		} else {
			motorInc[idx] = 1
		}
	}

	if counter[idx]%766 == 765 {
		if m.CurrentDirection == "forward" {
			m.Direction("backward")
		} else {
			m.Direction("forward")
		}
	}
}
