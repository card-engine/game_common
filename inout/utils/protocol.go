package utils

import (
	"fmt"
	"regexp"
)

func ParseCustomMessage(message string) (string, string, error) {
	// 使用正则表达式匹配 "数字[消息体]" 格式
	re := regexp.MustCompile(`^(\d+)(\[.*\])?$`)
	matches := re.FindStringSubmatch(message)

	if len(matches) < 2 {
		return "", "", fmt.Errorf("无效的消息格式: %s", message)
	}

	msgType := matches[1]
	payload := ""
	if len(matches) > 2 && matches[2] != "" {
		payload = matches[2]
	}

	return msgType, payload, nil
}
