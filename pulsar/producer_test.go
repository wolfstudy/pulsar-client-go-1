// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package pulsar

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/util"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

func TestInvalidURL(t *testing.T) {
	client, err := NewClient(ClientOptions{})

	if client != nil || err == nil {
		t.Fatal("Should have failed to create client")
	}
}

func TestProducerConnectError(t *testing.T) {
	client, err := NewClient(ClientOptions{
		URL: "pulsar://invalid-hostname:6650",
	})

	assert.Nil(t, err)

	defer client.Close()

	producer, err := client.CreateProducer(ProducerOptions{
		Topic: newTopicName(),
	})

	// Expect error in creating producer
	assert.Nil(t, producer)
	assert.NotNil(t, err)

	assert.Equal(t, err.Error(), "connection error")
}

func TestProducerNoTopic(t *testing.T) {
	client, err := NewClient(ClientOptions{
		URL: "pulsar://localhost:6650",
	})

	if err != nil {
		t.Fatal(err)
		return
	}

	defer client.Close()

	producer, err := client.CreateProducer(ProducerOptions{})

	// Expect error in creating producer
	assert.Nil(t, producer)
	assert.NotNil(t, err)

	assert.Equal(t, err.(*Error).Result(), ResultInvalidTopicName)
}

func TestSimpleProducer(t *testing.T) {
	client, err := NewClient(ClientOptions{
		URL: serviceURL,
	})
	assert.NoError(t, err)

	producer, err := client.CreateProducer(ProducerOptions{
		Topic: newTopicName(),
	})

	assert.NoError(t, err)
	assert.NotNil(t, producer)

	for i := 0; i < 10; i++ {
		err = producer.Send(context.Background(), &ProducerMessage{
			Payload: []byte("hello"),
		})

		assert.NoError(t, err)
	}

	err = producer.Close()
	assert.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}

func TestProducerAsyncSend(t *testing.T) {
	client, err := NewClient(ClientOptions{
		URL: serviceURL,
	})
	assert.NoError(t, err)

	producer, err := client.CreateProducer(ProducerOptions{
		Topic:                   newTopicName(),
		BatchingMaxPublishDelay: 1 * time.Second,
	})

	assert.NoError(t, err)
	assert.NotNil(t, producer)

	wg := sync.WaitGroup{}
	wg.Add(10)
	errors := util.NewBlockingQueue(10)

	for i := 0; i < 10; i++ {
		producer.SendAsync(context.Background(), &ProducerMessage{
			Payload: []byte("hello"),
		}, func(id MessageID, message *ProducerMessage, e error) {
			if e != nil {
				log.WithError(e).Error("Failed to publish")
				errors.Put(e)
			} else {
				log.Info("Published message ", id)
			}
			wg.Done()
		})

		assert.NoError(t, err)
	}

	err = producer.Flush()
	assert.Nil(t, err)

	wg.Wait()

	assert.Equal(t, 0, errors.Size())

	err = producer.Close()
	assert.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}

func TestProducerCompression(t *testing.T) {

	type testProvider struct {
		name            string
		compressionType CompressionType
	}

	var providers = []testProvider{
		{"zlib", ZLib},
		{"lz4", LZ4},
		{"zstd", ZSTD},
	}

	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			client, err := NewClient(ClientOptions{
				URL: serviceURL,
			})
			assert.NoError(t, err)

			producer, err := client.CreateProducer(ProducerOptions{
				Topic:           newTopicName(),
				CompressionType: p.compressionType,
			})

			assert.NoError(t, err)
			assert.NotNil(t, producer)

			for i := 0; i < 10; i++ {
				err = producer.Send(context.Background(), &ProducerMessage{
					Payload: []byte("hello"),
				})

				assert.NoError(t, err)
			}

			err = producer.Close()
			assert.NoError(t, err)

			err = client.Close()
			assert.NoError(t, err)
		})
	}
}

func TestProducerLastSequenceID(t *testing.T) {
	client, err := NewClient(ClientOptions{
		URL: serviceURL,
	})
	assert.NoError(t, err)

	producer, err := client.CreateProducer(ProducerOptions{
		Topic: newTopicName(),
	})

	assert.NoError(t, err)
	assert.NotNil(t, producer)

	assert.Equal(t, int64(-1), producer.LastSequenceID())

	for i := 0; i < 10; i++ {
		err = producer.Send(context.Background(), &ProducerMessage{
			Payload: []byte("hello"),
		})

		assert.NoError(t, err)
		assert.Equal(t, int64(i), producer.LastSequenceID())
	}

	err = producer.Close()
	assert.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}

func TestEventTime(t *testing.T) {
	client, err := NewClient(ClientOptions{
		URL: serviceURL,
	})
	assert.NoError(t, err)
	defer client.Close()

	topicName := "test-event-time"
	producer, err := client.CreateProducer(ProducerOptions{
		Topic: topicName,
	})
	assert.Nil(t, err)
	defer producer.Close()

	consumer, err := client.Subscribe(ConsumerOptions{
		Topic:            topicName,
		SubscriptionName: "subName",
	})
	assert.Nil(t, err)
	defer consumer.Close()

	eventTime := timeFromUnixTimestampMillis(uint64(1565161612))
	err = producer.Send(context.Background(), &ProducerMessage{
		Payload:   []byte(fmt.Sprintf("test-event-time")),
		EventTime: &eventTime,
	})
	assert.Nil(t, err)

	msg, err := consumer.Receive(context.Background())
	assert.Nil(t, err)
	actualEventTime := msg.EventTime()
	assert.Equal(t, eventTime.Unix(), actualEventTime.Unix())
}

func TestFlushInProducer(t *testing.T) {
	client, err := NewClient(ClientOptions{
		URL: serviceURL,
	})
	assert.NoError(t, err)
	defer client.Close()

	topicName := "test-flush-in-producer"
	subName := "subscription-name"
	numOfMessages := 10
	ctx := context.Background()

	// set batch message number numOfMessages, and max delay 10s
	producer, err := client.CreateProducer(ProducerOptions{
		Topic:                   topicName,
		DisableBatching:         false,
		BatchingMaxMessages:     uint(numOfMessages),
		BatchingMaxPublishDelay: time.Second * 10,
		BlockIfQueueFull:        true,
		Properties: map[string]string{
			"producer-name": "test-producer-name",
			"producer-id":   "test-producer-id",
		},
	})
	defer producer.Close()

	consumer, err := client.Subscribe(ConsumerOptions{
		Topic:            topicName,
		SubscriptionName: subName,
	})
	assert.Nil(t, err)
	defer consumer.Close()

	prefix := "msg-batch-async"
	msgCount := 0

	wg := sync.WaitGroup{}
	wg.Add(5)
	errors := util.NewBlockingQueue(10)
	for i := 0; i < numOfMessages/2; i++ {
		messageContent := prefix + fmt.Sprintf("%d", i)
		producer.SendAsync(ctx, &ProducerMessage{
			Payload: []byte(messageContent),
		}, func(id MessageID, producerMessage *ProducerMessage, e error) {
			if e != nil {
				log.WithError(e).Error("Failed to publish")
				errors.Put(e)
			} else {
				log.Info("Published message ", id)
			}
			wg.Done()
		})
		assert.Nil(t, err)
	}
	err = producer.Flush()
	assert.Nil(t, err)
	wg.Wait()

	for i := 0; i < numOfMessages/2; i++ {
		_, err := consumer.Receive(ctx)
		assert.Nil(t, err)
		msgCount++
	}

	assert.Equal(t, msgCount, numOfMessages/2)

	wg.Add(5)
	for i := numOfMessages / 2; i < numOfMessages; i++ {
		messageContent := prefix + fmt.Sprintf("%d", i)
		producer.SendAsync(ctx, &ProducerMessage{
			Payload: []byte(messageContent),
		}, func(id MessageID, producerMessage *ProducerMessage, e error) {
			if e != nil {
				log.WithError(e).Error("Failed to publish")
				errors.Put(e)
			} else {
				log.Info("Published message ", id)
			}
			wg.Done()
		})
		assert.Nil(t, err)
	}

	err = producer.Flush()
	assert.Nil(t, err)
	wg.Wait()

	for i := numOfMessages / 2; i < numOfMessages; i++ {
		_, err := consumer.Receive(ctx)
		assert.Nil(t, err)
		msgCount++
	}
	assert.Equal(t, msgCount, numOfMessages)
}

func TestFlushInPartitionedProducer(t *testing.T) {
	topicName := "persistent://public/default/partition-testFlushInPartitionedProducer12"

	// call admin api to make it partitioned
	url := adminURL + "/" + "admin/v2/" + topicName + "/partitions"
	makeHTTPCall(t, http.MethodPut, url, "5")

	numberOfPartitions := 5
	numOfMessages := 10
	ctx := context.Background()

	client, err := NewClient(ClientOptions{
		URL: serviceURL,
	})
	assert.NoError(t, err)
	defer client.Close()

	// set batch message number numOfMessages, and max delay 10s
	producer, err := client.CreateProducer(ProducerOptions{
		Topic:                   topicName,
		DisableBatching:         false,
		BatchingMaxMessages:     uint(numOfMessages / numberOfPartitions),
		BatchingMaxPublishDelay: time.Second * 10,
		BlockIfQueueFull:        true,
	})
	defer producer.Close()

	consumer, err := client.Subscribe(ConsumerOptions{
		Topic:            topicName,
		SubscriptionName: "my-sub",
		Type:             Exclusive,
	})
	assert.Nil(t, err)

	prefix := "msg-batch-async-"
	wg := sync.WaitGroup{}
	wg.Add(5)
	errors := util.NewBlockingQueue(10)
	for i := 0; i < numOfMessages/2; i++ {
		messageContent := prefix + fmt.Sprintf("%d", i)
		producer.SendAsync(ctx, &ProducerMessage{
			Payload: []byte(messageContent),
		}, func(id MessageID, producerMessage *ProducerMessage, e error) {
			if e != nil {
				log.WithError(e).Error("Failed to publish")
				errors.Put(e)
			} else {
				log.Info("Published message: ", id)
			}
			wg.Done()
		})
		assert.Nil(t, err)
	}

	// After flush, should be able to consume.
	err = producer.Flush()
	assert.Nil(t, err)

	wg.Wait()

	// Receive all messages
	msgCount := 0
	for i := 0; i < numOfMessages/2; i++ {
		messageContent := prefix + fmt.Sprintf("%d", i)
		msg, err := consumer.Receive(ctx)
		fmt.Printf("Received message msgId: %#v -- content: '%s'\n",
			msg.ID(), string(msg.Payload()))
		assert.Nil(t, err)
		assert.Equal(t, messageContent, string(msg.Payload()))
		msgCount++
	}
	assert.Equal(t, msgCount, numOfMessages/2)
}
