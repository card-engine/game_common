package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

type ApiClient struct {
	httpClient *http.Client
}

func NewApiClient(maxIdleConns, maxIdleConnsPerHost int, timeout time.Duration) *ApiClient {
	client := &ApiClient{}
	client.httpClient = &http.Client{
		Transport: &http.Transport{
			// 保留默认的 Dialer 配置
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          maxIdleConns,
			MaxIdleConnsPerHost:   maxIdleConnsPerHost,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: timeout,
	}
	return client
}

func (c *ApiClient) Do(accessKeyID, accessKeySecret, apiURL string, data []byte) ([]byte, error) {
	// 生成随机字符串和时间戳
	nonce := generateNonce(128)
	timestamp := time.Now().Unix()

	// 计算签名
	signature := generateSignature(accessKeySecret, nonce, timestamp)

	// 设置请求头
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Api Callback Error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AccessKeyId", accessKeyID)
	req.Header.Set("Nonce", nonce)
	req.Header.Set("Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("Sign", signature)

	// 发送 POST 请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Api Callback Error sending request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Api Callback Error reading response body: %w", err)
	}

	return body, nil
}
