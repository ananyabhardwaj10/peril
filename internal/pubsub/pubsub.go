package pubsub 
import(
	"encoding/json"
	"log"
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
    SimpleQueueDurable   SimpleQueueType = iota
    SimpleQueueTransient
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

	queue, err := ch.QueueDeclare(queueName, dur, autodel, exclusive, false, nil) 
	if err != nil {
		return nil, amqp.Queue{}, err 
	}

	err = ch.QueueBind(queueName, key, exchange, false, nil) 
	if err != nil {
		return nil, amqp.Queue{}, err  
	}

	return ch, queue, nil 
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T)) error {
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
			}
		
			handler(data)
			message.Ack(false) 
		}
	}()	

	return nil 
}