package eventbus

import (
	"cn.qingdou.server/game_common/eventbus/types"
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"testing"
	"time"
)

const brokers = "sss.cpigame.com:19092"

func TestPublishWinEvent(t *testing.T) {
	bus := NewKafkaBus([]string{brokers})
	for i := 0; i < 100; i++ {
		err := bus.Publish(context.Background(), types.GAME_EVENT_TOPIC, NewEvent("win", "test", &types.WinEvent{
			AppId:            "1001",
			Bet:              1,
			BetTransactionId: fmt.Sprintf("BetTransactionId%d", i),
			Currency:         "test",
			GameBrand:        "jili",
			GameId:           "223",
			GameType:         "slot",
			PlayerId:         fmt.Sprintf("Player%d", i%2),
			RoundId:          fmt.Sprintf("RoundId%d", i),
			TransactionId:    "test",
			Win:              1,
		}))
		if err != nil {
			t.Errorf("Failed to publish event: %v", err)
		}
	}
	bus.Close()
}

func TestPublish(t *testing.T) {
	bus := NewKafkaBus([]string{brokers})
	for i := 0; i < 100; i++ {
		err := bus.Publish(context.Background(), "test", NewEvent("test", "test", fmt.Sprintf("event test %d", i)))
		if err != nil {
			t.Errorf("Failed to publish event: %v", err)
		}
	}
	bus.Close()
}

func TestSubscribe(t *testing.T) {
	bus := NewKafkaBus([]string{brokers})
	err := bus.Subscribe(context.Background(), "test", "test", func(ctx context.Context, evt *Event) error {
		log.Infof("Received event: %#v", evt)
		log.Infof("Event payload: %s", string(evt.Payload))
		return nil
	})
	if err != nil {
		t.Errorf("Failed to subscribe to topic: %v", err)
	}
	time.Sleep(time.Hour)
	bus.Close()
}
