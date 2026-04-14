package pubsub 
import(
	"encoding/json"
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

	err = ch.ExchangeDeclare(
    exchange, // name
    "direct", // kind
    true,     // durable
    false,    // auto-delete
    false,    // internal
    false,    // no-wait
    nil,      // args
	)
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