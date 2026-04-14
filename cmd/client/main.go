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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Program shutting down. Closing the connection")
}
