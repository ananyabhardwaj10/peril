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
		log.Fatal("Unable to create a connection")
	}
	defer connection.Close()
	fmt.Println("Successful Connection")

	ch, err := connection.Channel()
	if err != nil {
		log.Fatal("Unable to create a new channel using the connection")
	}

	_, queue, err := pubsub.DeclareAndBind(connection, routing.ExchangePerilDirect, "game_logs", "game_logs.*", pubsub.SimpleQueueDurable)
	if err != nil {
		log.Fatalf("Error creating and binding the queue", err)
	}

	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	gamelogic.PrintServerHelp()

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		first_command := input[0]

		if first_command == "pause" {
			log.Println("Publishing paused game state")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
			if err != nil {
				log.Printf("Could not publish time: %v", err)
			}
		} else if first_command == "resume" {
			log.Println("Publishing resumes game state")
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
			if err != nil {
				log.Printf("Could not publish time: %v", err)
			}
		} else if first_command == "quit" {
			log.Println("Quitting the game")
			break
		} else {
			log.Println("Unknown command")
		}
	}


	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Program shutting down. Closing the connection")
}
