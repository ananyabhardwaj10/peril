package pubsub 
import(
	"encoding/json"
	"log"
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
    SimpleQueueDurable   SimpleQueueType = iota
    SimpleQueueTransient
)

type AckType int 

const (
	ACK AckType = iota
	NACKREQUEUE 
	NACKDISCARD 
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing {
		ContentType: "application/json", 
		Body: data,
	})
	if err != nil {
		return err
	}

	return nil 
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	var dur, autodel, exclusive bool

	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err 
	}

	if queueType == SimpleQueueDurable {
		dur = true
	} 

	if queueType == SimpleQueueTransient {
		autodel = true
		exclusive = true
	}

	queue, err := ch.QueueDeclare(queueName, dur, autodel, exclusive, false, amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}) 
	if err != nil {
		return nil, amqp.Queue{}, err 
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil) 
	if err != nil {
		return nil, amqp.Queue{}, err  
	}

	return ch, queue, nil 
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) AckType) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err 
	}

	delivery_chan, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err 
	}

	go func() {
		for message := range delivery_chan {
			var data T
			err = json.Unmarshal(message.Body, &data)
			if err != nil {
				log.Printf("error: %v", err)
				continue
			}
		
			ack_type := handler(data)
			if ack_type == ACK {
				message.Ack(false) 
				fmt.Println("ack occurred")
			} else if ack_type == NACKREQUEUE {
				message.Nack(false, true)
				fmt.Println("nack and requeued")
			} else {
				message.Nack(false, false)
				fmt.Println("nack and discarded")
			}
		}
	}()	

	return nil 
}