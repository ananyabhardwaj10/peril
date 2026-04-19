package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connection_string := "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connection_string)
	if err != nil {
		log.Fatalf("Unable to create a connection", err)
	}
	defer connection.Close()
	fmt.Println("Successful Connection")

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Unable to get the username", err)
	}

	queue_name := routing.PauseKey + "." + userName

	_, queue, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, queue_name, routing.PauseKey, pubsub.SimpleQueueTransient)
	if err != nil {
		log.Fatalf("Error creating and binding the queue", err)
	}

	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	game_state := gamelogic.NewGameState(userName)

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		command := input[0]

		switch command {
		case "spawn":
			err = game_state.CommandSpawn(input)
			if err != nil {
				fmt.Printf("error spawning the location: %v", err)
			}
		
		case "move":
			_, err := game_state.CommandMove(input)
			if err != nil {
				fmt.Printf("error moving the army: %v", err)
			} 
		
		case "status":
			game_state.CommandStatus()

		case  "help":
			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet!")

		case "quit":
			gamelogic.PrintQuit()
			return 

		default:
			fmt.Println("Please enter a valid command")
			continue
		}

	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Program shutting down. Closing the connection")
}
