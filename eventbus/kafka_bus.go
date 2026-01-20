package eventbus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/segmentio/kafka-go"
)

type KafkaBus struct {
	brokers        []string
	writers        map[string]*kafka.Writer
	readers        map[string]*kafka.Reader
	kafkaWriteLock sync.Mutex
	kafkaReadLock  sync.Mutex
}

// 创建一个 Kafka 事件总线
func NewKafkaBus(brokers []string) *KafkaBus {
	return &KafkaBus{
		brokers: brokers,
		writers: make(map[string]*kafka.Writer),
		readers: make(map[string]*kafka.Reader),
	}
}

// 发布事件
func (b *KafkaBus) Publish(ctx context.Context, topic string, evt *Event) error {
	if evt == nil {
		return fmt.Errorf("topic %s publish event is nil", topic)
	}
	writer, ok := b.writers[topic]
	if !ok {
		b.kafkaWriteLock.Lock()
		writer, ok = b.writers[topic]
		if !ok {
			writer = &kafka.Writer{
				Addr:         kafka.TCP(b.brokers...),
				Topic:        topic,
				Balancer:     &kafka.LeastBytes{},
				Async:        true,
				RequiredAcks: kafka.RequireAll,
				Completion: func(messages []kafka.Message, err error) {
					if err != nil {
						log.Errorf("Kafka write failed for topic %s: %v", topic, err)
					}
				},
			}
			b.writers[topic] = writer
			log.Infof("Created new Kafka %s writer for topic %s", b.brokers, topic)
		}
		b.kafkaWriteLock.Unlock()
	}

	data, err := evt.Encode()
	if err != nil {
		return err
	}

	return writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(evt.EventID),
		Value: data,
	})
}

// 订阅事件
func (b *KafkaBus) Subscribe(ctx context.Context, topic, group string, handler EventHandler) error {
	// 使用 topic+group 作为唯一标识避免冲突
	key := fmt.Sprintf("%s-%s", topic, group)
	b.kafkaReadLock.Lock()
	// 检查是否已存在相同订阅
	if _, exists := b.readers[key]; exists {
		b.kafkaReadLock.Unlock()
		return fmt.Errorf("already subscribed to topic %s with group %s", topic, group)
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  b.brokers,
		Topic:    topic,
		GroupID:  group,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	b.readers[key] = reader
	log.Infof("Created new Kafka %s reader for topic %s, group %s", b.brokers, topic, group)
	b.kafkaReadLock.Unlock()
	// 启动消费者 goroutine
	go func() {
		defer func() {
			// 清理资源
			b.kafkaReadLock.Lock()
			delete(b.readers, key)
			b.kafkaReadLock.Unlock()
		}()

		for {
			select {
			case <-ctx.Done():
				log.Infof("Stopping consumer for topic %s, group %s", topic, group)
				return
			default:
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return // Context cancelled
					}
					log.Infof("Kafka read error (topic: %s, group: %s): %v", topic, group, err)
					time.Sleep(time.Second)
					continue
				}

				evt, err := DecodeEvent(msg.Value)
				if err != nil {
					log.Infof("Decode event error (topic: %s, group: %s): %v", topic, group, err)
					continue
				}

				if err := handler(ctx, evt); err != nil {
					log.Infof("Event handler error: %v (event: %s, topic: %s, group: %s)",
						err, evt.Type, topic, group)
				}
			}
		}
	}()
	return nil
}

func (b *KafkaBus) Close() error {
	for _, w := range b.writers {
		_ = w.Close()
	}
	for _, r := range b.readers {
		_ = r.Close()
	}
	return nil
}
