package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"
)

// DiceRoller rolls dice
type DiceRoller struct {
	rolling  bool
	stopChan chan struct{}
}

// Start begins rolling
func (dr *DiceRoller) Start() {
	if dr.rolling {
		fmt.Println("Already rolling!")
		return
	}
	dr.rolling = true
	dr.stopChan = make(chan struct{})

	go func() {
		count := 0
		d1, d2, total := 0, 0, 0
		for dr.rolling {
			select {
			case <-dr.stopChan:
				dr.rolling = false
				return
			default:
				d1, d2, total = roll()
				fmt.Printf("[roller] %d + %d = %d\n", d1, d2, total)

				if total == 4 {
					count++
				}

				if total == 7 {
					fmt.Printf("You rolled %d 4s!\n", count)
					count = 0
				}
				time.Sleep(2 * time.Second)
			}
		}
	}()
}

// Stop stops rolling
func (dr *DiceRoller) Stop() {
	if !dr.rolling {
		fmt.Println("Already stopped!")
		return
	}
	close(dr.stopChan)
	dr.rolling = false
	fmt.Println("Agent stopped.  Awaiting commands...")
}

func roll() (int, int, int) {
	die1 := rand.Intn(6) + 1
	die2 := rand.Intn(6) + 1
	return die1, die2, die1 + die2
}

func main() {
	// Connect to NATS server
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		fmt.Println("Failed to conect to NATS:", err)
		return
	}
	defer nc.Close()

	// Create a new DiceRoller
	dr := &DiceRoller{}

	// Subscribe to the control topic
	_, err = nc.Subscribe("control", func(msg *nats.Msg) {
		switch string(msg.Data) {
		case "start":
			fmt.Println("Rolling...")
			dr.Start()
		case "stop":
			fmt.Println("Stopping...")
			dr.Stop()
		default:
			fmt.Println("Unknown command:", string(msg.Data))
		}
	})
	if err != nil {
		fmt.Println("Failed to subscribe:", err)
		return
	}

	fmt.Println("Agent is operational.  Awaiting commands...")

	// Keep the service running
	select {}
}
