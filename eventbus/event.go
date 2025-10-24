package eventbus

import (
	"encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"time"
)

// Event 是通用事件结构
type Event struct {
	EventID   string            `json:"id"`        // 唯一ID（可用UUID）
	Type      string            `json:"type"`      // 事件类型（例如 "bet"）
	Timestamp int64             `json:"timestamp"` // 事件时间戳
	Source    string            `json:"source"`    // 来源服务名（如 "api_server"）
	Metadata  map[string]string `json:"metadata"`  // 附加信息（trace_id 等）
	Payload   []byte            `json:"payload"`   // 实际事件内容
}

// 序列化
func (e *Event) Encode() ([]byte, error) {
	return json.Marshal(e)
}

// 反序列化
func DecodeEvent(data []byte) (*Event, error) {
	var evt Event
	err := json.Unmarshal(data, &evt)
	return &evt, err
}

// 反序列化负载内容
func (e *Event) DecodePayload(out interface{}) error {
	return json.Unmarshal(e.Payload, out)
}

// 工具函数：创建新事件
func NewEvent(eventType, source string, payload interface{}) *Event {
	payloadData, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Failed to marshal payload: %v", err)
		return nil
	}
	return &Event{
		EventID:   generateUUID(),
		Type:      eventType,
		Source:    source,
		Payload:   payloadData,
		Timestamp: time.Now().Unix(),
	}
}

// 简单 UUID 生成
func generateUUID() string {
	return time.Now().Format("20060102150405.000000000")
}
