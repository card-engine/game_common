package eventbus

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/segmentio/kafka-go"
	"testing"
)

// readByReader 通过Reader接收消息
func TestReader(t *testing.T) {
	// 创建Reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		Topic:     "test",
		Partition: 0,
		MaxBytes:  10e6, // 10MB
	})
	//r.SetOffset(42) // 设置Offset

	// 接收消息
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	// 程序退出前关闭Reader
	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}

func TestWrite(t *testing.T) {
	// 创建一个writer 向topic-A发送消息
	w := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:9092", "localhost:9093", "localhost:9094"),
		Topic:                  "test",
		Balancer:               &kafka.LeastBytes{}, // 指定分区的balancer模式为最小字节分布
		RequiredAcks:           kafka.RequireAll,    // ack模式
		Async:                  true,                // 异步
		AllowAutoTopicCreation: true,
	}

	err := w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte("Key-A"),
			Value: []byte("Hello World!"),
		},
		kafka.Message{
			Key:   []byte("Key-B"),
			Value: []byte("One!"),
		},
		kafka.Message{
			Key:   []byte("Key-C"),
			Value: []byte("Two!"),
		},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := w.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}

}
