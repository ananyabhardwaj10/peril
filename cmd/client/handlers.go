package main

import(
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
)

func handlerPause(gs *gamelogic.GameState) (func(routing.PlayingState) pubsub.AckType) {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps) 
		return pubsub.ACK 
	} 
}

func handlerMove(gs *gamelogic.GameState) (func(gamelogic.ArmyMove) pubsub.AckType) {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		mvOutcome := gs.HandleMove(move)
		if mvOutcome == gamelogic.MoveOutComeSafe || mvOutcome == gamelogic.MoveOutcomeMakeWar {
			return pubsub.ACK
		} else {
			return pubsub.NACKDISCARD
		}
	}
}