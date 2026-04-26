package main

import(
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) (func(routing.PlayingState) pubsub.AckType) {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps) 
		return pubsub.ACK 
	} 
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) (func(gamelogic.ArmyMove) pubsub.AckType) {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		userName := gs.GetUsername()
		mvOutcome := gs.HandleMove(move)
		if mvOutcome == gamelogic.MoveOutcomeMakeWar {
			key := routing.WarRecognitionsPrefix + "." + userName
			err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, key, gamelogic.RecognitionOfWar{
   				Attacker: move.Player,
   				Defender: gs.GetPlayerSnap(),
			})
			if err != nil {
				fmt.Printf("error: %v", err)
				return pubsub.NACKREQUEUE
			}

			return pubsub.ACK
		}

		if mvOutcome == gamelogic.MoveOutComeSafe {
			return pubsub.ACK
		} else {
			return pubsub.NACKDISCARD
		}
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) (func(gamelogic.RecognitionOfWar) pubsub.AckType) {
	userName := gs.GetUsername()
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outCome, winner, loser := gs.HandleWar(rw) 
		if outCome == gamelogic.WarOutcomeNotInvolved {
			return pubsub.NACKREQUEUE
		} else if outCome == gamelogic.WarOutcomeNoUnits {
			return pubsub.NACKDISCARD 
		} else if outCome == gamelogic.WarOutcomeOpponentWon || outCome == gamelogic.WarOutcomeYouWon {
			msg := winner + " won a war against " + loser
			err := pubsub.HelperPublish(ch, userName, msg)
			if err != nil {
				return pubsub.NACKREQUEUE
			}
			return pubsub.ACK 
		} else if outCome == gamelogic.WarOutcomeDraw {
			msg := "A war between " + winner + " and " + loser + " resulted in a draw"
			err := pubsub.HelperPublish(ch, userName, msg)
			if err != nil {
				return pubsub.NACKREQUEUE
			}
			return pubsub.ACK
		} else {
			fmt.Printf("Unknown outcome!")
			return pubsub.NACKDISCARD
		}
	}
}

