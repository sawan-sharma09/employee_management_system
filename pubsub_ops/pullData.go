package pubsubops

import (
	"context"
	"fmt"
	"managedata/app_errors"
	initpack "managedata/init_pack"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber"
)

func Pull_PubsubData(c *gin.Context) {
	subID := os.Getenv("SUBSCRIPTION_ID")
	ctx := context.Background()

	cctxx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := initpack.Client.Subscription(subID).Receive(cctxx, func(ctx context.Context, msg *pubsub.Message) {
		fmt.Println("Pulled message: ", string(msg.Data))

		response := fmt.Sprintf("\nReceived message:%s", string(msg.Data))
		c.String(http.StatusOK, response)

		msg.Ack()
	})
	if err != nil {
		fmt.Println("Error in pulling message from pubsub topic: ", err)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "INFO", Message: app_errors.ErrPubsubMessage, Endpoint: c.Request.URL.Path, Status_code: http.StatusInternalServerError}
		c.JSON(http.StatusUnauthorized, logDetails)
		return
	}
	c.Status(fiber.StatusOK)
}
