package pubsubops

import (
	"context"
	"encoding/json"
	"fmt"

	initpack "managedata/init_pack"
	"managedata/util"

	"cloud.google.com/go/pubsub"
)

func PublishData(fetchedData util.Employee) (string, error) {
	ctx := context.Background()

	jsonData, marshalErr := json.Marshal(&fetchedData)
	if marshalErr != nil {
		fmt.Println("Error in marshal: ", marshalErr)
		return "", marshalErr
	}

	res := initpack.Topic.Publish(ctx, &pubsub.Message{
		Data: jsonData,
	})

	id, err := res.Get(ctx)
	if err != nil {
		fmt.Println("Id generation error from pubsub..")
		return "", err
	} else {
		fmt.Println("Published message with msg ID:", id)
		return id, nil
	}

}
