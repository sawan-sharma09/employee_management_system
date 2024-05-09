package pubsubops

import (
	"context"
	"fmt"
	initpack "managedata/init_pack"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
)

func Pull_PubsubData(w http.ResponseWriter, r *http.Request) {
	subID := os.Getenv("SUBSCRIPTION_ID")
	ctx := context.Background()

	cctxx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := initpack.Client.Subscription(subID).Receive(cctxx, func(ctx context.Context, msg *pubsub.Message) {
		fmt.Fprintln(w, "Got message: ", string(msg.Data))
		fmt.Println("Pulled message: ", string(msg.Data))
		msg.Ack()
	})
	if err != nil {
		fmt.Println("Error in pulling message from pubsub topic: ", err)
	}

}
