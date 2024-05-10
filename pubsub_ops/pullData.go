package pubsubops

import (
	"context"
	"fmt"
	initpack "managedata/init_pack"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/gofiber/fiber/v2"
)

func Pull_PubsubData(c *fiber.Ctx) error {
	subID := os.Getenv("SUBSCRIPTION_ID")
	ctx := context.Background()

	cctxx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := initpack.Client.Subscription(subID).Receive(cctxx, func(ctx context.Context, msg *pubsub.Message) {
		fmt.Println("Pulled message: ", string(msg.Data))
		msg.Ack()
	})
	if err != nil {
		fmt.Println("Error in pulling message from pubsub topic: ", err)
		return err
	}
	return c.SendStatus(fiber.StatusOK)
}
