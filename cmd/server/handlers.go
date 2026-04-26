package main
import(
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
)

func handlerLogs(gl routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(gl) 
		if err != nil {
			fmt.Printf("error writing the game logs: %v", err)
			return pubsub.NACKDISCARD
		}

		return pubsub.ACK 
}