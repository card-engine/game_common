package utils

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/qd2ss/sfs"
)

func Unpack(buff []byte) (int16, uint8, sfs.SFSObject, error) {
	unpacker := sfs.NewUnpacker(buff)

	result, err := unpacker.Unpack()
	if err != nil {
		return 0, 0, nil, err
	}
	var action int16
	var controller uint8
	var payload sfs.SFSObject

	if m, ok := result.(sfs.SFSObject); ok {
		if a, ok := m["a"].(int16); ok {
			action = a
		}
		if c, ok := m["c"].(uint8); ok {
			controller = c
		}
		if p, ok := m["p"].(sfs.SFSObject); ok {
			payload = p
		}
	} else {
		return 0, 0, nil, fmt.Errorf("unpack error, data type is not map[string]interface{}")
	}
	return action, controller, payload, nil
}

func Pack(controller int16, action uint8, data sfs.SFSObject) ([]byte, error) {
	sendData := sfs.SFSObject{
		"a": action,
		"c": controller,
		"p": data,
	}

	packer := sfs.NewPacker()
	buff, err := packer.Pack(sendData)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func PackCustomData(cmd string, data sfs.SFSObject) ([]byte, error) {
	rsp := sfs.SFSObject{
		"c": cmd,
		"p": data,
	}

	buff, err := Pack(1, 13, rsp)
	if err != nil {
		return nil, err
	}

	return buff, err
}

// 格式化一个文档手册
// 一行测试。标注:base64的二进制
func GenSfsDoc(txtFile string, docFile string) error {
	// 打开文件
	file, err := os.Open(txtFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建输出Markdown文件
	outputFile, err := os.Create(docFile)
	if err != nil {
		log.Fatal("创建输出文件失败:", err)
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	processedCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		if line == "" {
			continue // 跳过空行
		}

		// 处理包含冒号分隔的行（标注:base64文本）
		if strings.Contains(line, ":") {
			if err := processLine(line, outputFile, processedCount); err != nil {
				log.Printf("第%d行处理失败: %v", lineCount, err)
			} else {
				processedCount++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("读取文件失败:", err)
	}

	fmt.Printf("处理完成! 共处理了%d行数据，成功处理%d条记录\n", lineCount, processedCount)
	return nil
}

func processLine(line string, outputFile *os.File, count int) error {
	// 分割标注和base64文本
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return fmt.Errorf("格式错误，缺少冒号分隔")
	}

	label := strings.TrimSpace(parts[0])
	base64Str := strings.TrimSpace(parts[1])

	// 解码base64
	decoded, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return fmt.Errorf("base64解码失败: %v", err)
	}
	unpacker := sfs.NewUnpacker(decoded)
	v, err := unpacker.Unpack()
	if err != nil {
		return err
	}

	formattedJSON, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}

	// 写入Markdown格式
	writeMarkdownEntry(outputFile, label, base64Str, string(formattedJSON), count)

	return nil
}

func writeMarkdownEntry(file *os.File, label, base64Str, formattedJSON string, count int) {
	// 如果不是第一条记录，添加分隔符
	if count > 0 {
		file.WriteString("\n\n---\n\n")
	}

	// 写入Markdown内容
	file.WriteString(fmt.Sprintf("# %s\n\n", label))

	file.WriteString("## Base64原始文本\n```\n")
	file.WriteString(base64Str)
	file.WriteString("\n```\n\n")

	file.WriteString("## 格式化后的JSON\n```json\n")
	file.WriteString(formattedJSON)
	file.WriteString("\n```")
}
