package main

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

var topicPartitionCount sync.Map

type metaDataFetcher interface {
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
}

func syncTopicMetadata(ctx context.Context, producer metaDataFetcher) error {

	if err := processMetadata(producer); err != nil {
		return err
	}
	go func() {
		select {
		case <-ctx.Done():
			return

		case <-time.After(kafkaMetadataInterval):
			if err := processMetadata(producer); err != nil {
				logrus.WithError(err).Error("could not fetch topic metadata")
			}
		}
	}()
	return nil
}

func processMetadata(producer metaDataFetcher) error {
	metadata, err := producer.GetMetadata(nil, true, int(math.Ceil(kafkaMetadataTimeout.Seconds())))
	if err != nil {
		return err
	}
	for name, topic := range metadata.Topics {
		topicPartitionCount.Store(name, len(topic.Partitions))
	}
	return nil
}
