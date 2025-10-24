package slot

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	redisClient "github.com/redis/go-redis/v9"
)

// rtp计算工具，以及对应的样本数据缓存(仅缓存数据库id)
type Rtp struct {
	brand       string
	redisClient *redisClient.Client
	rtpConfig   *sync.Map

	cacheByLocal  bool
	localCacheMap sync.Map //本地存储的容器
}

type SpinData struct {
	ID       uint    `gorm:"primaryKey;column:id"` // 主键ID
	Rate     float64 `gorm:"column:rate"`          // 倍率字段
	GameType int     `gorm:"column:gameType"`      // 游戏类型字段
}

func NewRtp(brand string, redisClient *redisClient.Client, consulAdd string, consulToken string) *Rtp {
	rtp := &Rtp{brand: brand, redisClient: redisClient, rtpConfig: new(sync.Map), cacheByLocal: true}
	rtp.loadRtpConfig(consulAdd, consulToken)
	return rtp
}

func (r *Rtp) loadRtpConfig(consulAdd string, consulToken string) {
	consulClient, err := api.NewClient(&api.Config{
		Address: consulAdd,
		Token:   consulToken,
	})
	if err != nil {
		panic(err)
	}

	// 使用结构体封装共享状态
	type keyState struct {
		sync.RWMutex
		modifyIndex map[string]uint64
	}

	state := &keyState{
		modifyIndex: make(map[string]uint64),
	}

	// 统一处理KV对的函数
	processKV := func(kvPair *api.KVPair) {
		fmt.Printf("Key: %s, Value: %s\n", kvPair.Key, string(kvPair.Value))

		var config RtpConfig
		if err := json.Unmarshal(kvPair.Value, &config); err != nil {
			fmt.Printf("Error parsing JSON for key %s: %v\n", kvPair.Key, err)
			return
		}

		// 提取 key 中的数字部分
		parts := strings.Split(kvPair.Key, "/")
		if len(parts) == 0 {
			fmt.Printf("Invalid key format: %s\n", kvPair.Key)
			return
		}
		extractedKey := parts[len(parts)-1]
		// extractedKey, err := strconv.Atoi(extractedKeyStr)
		// if err != nil {
		// 	fmt.Printf("Failed to convert %s to integer: %v\n", extractedKeyStr, err)
		// 	return
		// }

		tableName := r.brand + "_spin_" + extractedKey

		r.rtpConfig.Store(tableName, &config)

		// 清除缓存
		r.ClearCacheByRate(tableName)
	}

	// 初始化加载和监听变化的通用函数
	watchKeys := func() (uint64, error) {
		kvPairs, meta, err := consulClient.KV().List("aigc/"+r.brand+"/", &api.QueryOptions{
			WaitIndex: 0,
			WaitTime:  5 * time.Minute,
		})
		if err != nil {
			return 0, err
		}

		if kvPairs == nil {
			fmt.Println("No keys found under aigc/" + r.brand + "/")
			return meta.LastIndex, nil
		}

		newIndexes := make(map[string]uint64)
		for _, kvPair := range kvPairs {
			newIndexes[kvPair.Key] = kvPair.ModifyIndex
			processKV(kvPair)
		}

		state.Lock()
		defer state.Unlock()
		state.modifyIndex = newIndexes

		return meta.LastIndex, nil
	}

	// 初始加载
	lastIndex, err := watchKeys()
	if err != nil {
		panic(err)
	}

	// 启动监听goroutine
	go func() {
		for {
			kvPairs, meta, err := consulClient.KV().List("aigc/"+r.brand+"/", &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  5 * time.Minute,
			})
			if err != nil {
				fmt.Printf("Error listening for KV changes: %v. Retrying...\n", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if meta.LastIndex <= lastIndex {
				continue
			}

			lastIndex = meta.LastIndex
			newIndexes := make(map[string]uint64)

			if kvPairs != nil {
				for _, kvPair := range kvPairs {
					newIndexes[kvPair.Key] = kvPair.ModifyIndex

					state.RLock()
					oldIndex, exists := state.modifyIndex[kvPair.Key]
					state.RUnlock()

					if !exists || oldIndex != kvPair.ModifyIndex {
						processKV(kvPair)
					}
				}
			}

			// 检查被删除的key
			state.RLock()
			for key := range state.modifyIndex {
				if _, exists := newIndexes[key]; !exists {
					fmt.Printf("Key %s has been deleted.\n", key)
				}
			}
			state.RUnlock()

			// 更新索引
			state.Lock()
			state.modifyIndex = newIndexes
			state.Unlock()
		}
	}()
}

// 有没有缓存数据
func (r *Rtp) HasCacheSimpleRate(gi string) (bool, error) {
	tableName := r.brand + "_spin_" + gi
	cacheKey := fmt.Sprintf("%s_normal:all_rates", tableName)

	_, ok := r.localCacheMap.Load(cacheKey)
	return ok, nil
}

func (r *Rtp) HasCacheSimpleRate2(tableName string) (bool, error) {

	cacheKey := fmt.Sprintf("%s_normal:all_rates", tableName)

	_, ok := r.localCacheMap.Load(cacheKey)
	return ok, nil
}

// 以rate为key，缓存样本的id
func (r *Rtp) CacheSimpleByRate(gi string, records []SpinData) error {
	tableName := r.brand + "_spin_" + gi
	return r.CacheSimpleByRate2(tableName, records)
}

func (r *Rtp) CacheSimpleByRate2(tableName string, records []SpinData) error {
	// 使用更高效的数据结构来存储
	rateSetMap := make(map[string][]interface{})
	allRateSetMap := make(map[string]map[string]struct{}) // 使用map实现去重

	for _, rec := range records {
		rateStr := fmt.Sprintf("%.6f", rec.Rate)

		// 根据gameType确定key的前缀
		prefix := "normal"
		if rec.GameType == 1 {
			prefix = "special"
		}

		// 构建keys
		rateKey := fmt.Sprintf("%s_%s:rate:%s", tableName, prefix, rateStr)
		allRateKey := fmt.Sprintf("%s_%s:all_rates", tableName, prefix)

		// 聚合ID到对应的rateKey
		rateSetMap[rateKey] = append(rateSetMap[rateKey], rec.ID)

		// 初始化allRateSetMap的子map（如果不存在）
		if _, exists := allRateSetMap[allRateKey]; !exists {
			allRateSetMap[allRateKey] = make(map[string]struct{})
		}

		// 使用map特性自动去重
		allRateSetMap[allRateKey][rateStr] = struct{}{}
	}

	for key, members := range rateSetMap {
		log.Print("key", key)
		r.localCacheMap.Store(key, members)
	}

	for key, rateSet := range allRateSetMap {
		var rateSlice []interface{}
		for rate := range rateSet {
			rateSlice = append(rateSlice, rate)
		}
		r.localCacheMap.Store(key, rateSlice)
	}

	return nil
}

// 获取样本所有的rate
func (r *Rtp) GetRoundRate(gi string, isSpecial bool) float64 {
	tableName := r.brand + "_spin_" + gi
	return r.GetRoundRate2(tableName, isSpecial)
}

func (r *Rtp) GetRoundRate2(tableName string, isSpecial bool) float64 {
	allRateKey := ""
	if isSpecial {
		allRateKey = fmt.Sprintf("%s_special:all_rates", tableName)
	} else {
		allRateKey = fmt.Sprintf("%s_normal:all_rates", tableName)
	}
	if list, ok := r.localCacheMap.Load(allRateKey); ok {
		if rates, ok := list.([]interface{}); ok && len(rates) > 0 {
			// 随机选择一个索引
			randomIndex := rand.IntN(len(rates))
			if rateStr, ok := rates[randomIndex].(string); ok {
				if rate, err := strconv.ParseFloat(rateStr, 64); err == nil {
					return rate
				}
			}
		}
	}

	return 0
}

// 通过rate随机获取一条样本数据
func (r *Rtp) GetOneSimpleByRate(gi string, isSpecial bool, rate float64) (interface{}, error) {
	tableName := r.brand + "_spin_" + gi
	return r.GetOneSimpleByRate2(tableName, isSpecial, rate)
}

func (r *Rtp) GetOneSimpleByRate2(tableName string, isSpecial bool, rate float64) (interface{}, error) {
	rateKey := ""
	// 如果是特殊模式的，则保存到special:rate:xxx的集合中
	if isSpecial {
		rateKey = fmt.Sprintf("%s_special:rate:%.6f", tableName, rate)
	} else {
		rateKey = fmt.Sprintf("%s_normal:rate:%.6f", tableName, rate)
	}

	var id interface{}
	if list, ok := r.localCacheMap.Load(rateKey); ok {
		if ids, ok := list.([]interface{}); ok {
			if len(ids) > 0 {
				// 从本地缓存中随机选择一个 ID
				randomIndex := rand.IntN(len(ids))
				id = ids[randomIndex]
			}
		} else {
			return nil, errors.New("没有发现配置信息")
		}
	} else {
		return nil, errors.New("没有发现配置信息")
	}

	return id, nil
}

func (r *Rtp) ClearCacheByRate(gi string) error {
	// tableName := "pg_spin_" + strconv.Itoa(gi)
	// ctx := context.Background()

	// // 构建可能的缓存键模式
	// specialRatePattern := fmt.Sprintf("%s_special:rate:*", tableName)
	// normalRatePattern := fmt.Sprintf("%s_normal:rate:*", tableName)
	// specialAllRatesKey := fmt.Sprintf("%s_special:all_rates", tableName)
	// normalAllRatesKey := fmt.Sprintf("%s_normal:all_rates", tableName)

	// // 查找并删除特殊模式的 rate 缓存键
	// specialRateKeys, err := r.redisClient.Keys(ctx, specialRatePattern).Result()
	// if err != nil {
	// 	log.Printf("Failed to get special rate keys: %v", err)
	// 	return err
	// }
	// if len(specialRateKeys) > 0 {
	// 	if err := r.redisClient.Del(ctx, specialRateKeys...).Err(); err != nil {
	// 		log.Printf("Failed to delete special rate keys: %v", err)
	// 		return err
	// 	}
	// }

	// // 查找并删除普通模式的 rate 缓存键
	// normalRateKeys, err := r.redisClient.Keys(ctx, normalRatePattern).Result()
	// if err != nil {
	// 	log.Printf("Failed to get normal rate keys: %v", err)
	// 	return err
	// }
	// if len(normalRateKeys) > 0 {
	// 	if err := r.redisClient.Del(ctx, normalRateKeys...).Err(); err != nil {
	// 		log.Printf("Failed to delete normal rate keys: %v", err)
	// 		return err
	// 	}
	// }

	// // 删除特殊模式和普通模式的 all_rates 键
	// if err := r.redisClient.Del(ctx, specialAllRatesKey, normalAllRatesKey).Err(); err != nil {
	// 	log.Printf("Failed to delete all_rates keys: %v", err)
	// 	return err
	// }

	return nil
}

func (r *Rtp) GetRateByWeight(ratesWithWeights []RateWeight) float64 {
	if len(ratesWithWeights) == 0 {
		return 0
	}
	if len(ratesWithWeights) == 0 {
		return 0
	}

	// 计算总权重
	totalWeight := 0
	for _, rw := range ratesWithWeights {
		totalWeight += rw.Weighting
	}

	// 生成一个 0 到总权重之间的随机数
	randomNum := rand.IntN(totalWeight)

	currentWeight := 0
	for _, rw := range ratesWithWeights {
		currentWeight += rw.Weighting
		if randomNum < currentWeight {
			return rw.Rate
		}
	}

	// 理论上不会执行到这里，但为了保证函数完整性返回最后一个 rate
	return 0
}

// 验证所有的rate都有对应的权重
func (r *Rtp) VerifyAllRatesHaveWeights(allRates []float64, ratesWithWeights map[float64]int) (bool, error) {
	for _, rate := range allRates {
		if _, exists := ratesWithWeights[rate]; !exists {
			return false, fmt.Errorf("rate %.2f 未配置权重", rate)
		}
	}
	return true, nil
}

func (r *Rtp) GetRtpConfig(gi string, rtp string) (*RateConfig, bool) {
	tableName := r.brand + "_spin_" + gi
	return r.GetRtpConfig2(tableName, rtp)
}

func (r *Rtp) GetRtpConfig2(tableName string, rtp string) (*RateConfig, bool) {

	// // 1. 检查rtp参数是否为空
	// if rtp == "" {
	// 	return nil, false
	// }

	// 2. 从sync.Map中加载配置
	ret, ok := r.rtpConfig.Load(tableName)
	if !ok {
		return nil, false
	}

	// 3. 类型断言检查
	rtpConfig, ok := ret.(*RtpConfig)
	if !ok {
		return nil, false
	}

	// 4. 检查Data map是否初始化
	if rtpConfig.Data == nil {
		return nil, false
	}

	if rtp == "" {
		// for test
		rtp = rtpConfig.Use
	}

	// 5. 检查请求的rtp配置是否存在
	rateConfig, exists := rtpConfig.Data[rtp]
	if !exists {
		return nil, false
	}

	return &rateConfig, true
}

func (r *Rtp) IsSpecialModeTriggered(rateConfig *RateConfig) bool {

	// 5. 校验Rate值合法性 (0 <= Rate <= 1)
	if rateConfig.Rate < 0 || rateConfig.Rate > 1 {
		// r.logger.Warnf("invalid rate value %.2f, should be between 0 and 1", rateConfig.Rate)
		return false
	}

	// 6. 边界情况处理
	switch {
	case rateConfig.Rate == 0:
		return false
	case rateConfig.Rate == 1:
		return true
	}

	return rand.Float64() < rateConfig.Rate
}

func (r *Rtp) LoadWeightsFromJSON(rateConfig *RateConfig, isSpecial bool) ([]RateWeight, error) {
	var weights []RateWeight
	if isSpecial {
		weights = rateConfig.Special
	} else {
		weights = rateConfig.Normal
	}

	getModeName := func(isSpecial bool) string {
		if isSpecial {
			return "special"
		}
		return "normal"
	}
	// 6. 检查权重配置是否有效
	if weights == nil {
		return nil, fmt.Errorf("%s weights are not initialized (nil)", getModeName(isSpecial))
	}

	// 7. 检查空数组情况
	if len(weights) == 0 {
		return nil, fmt.Errorf("%s weights array is empty", getModeName(isSpecial))
	}

	// 8. 验证权重值的合法性
	// if err := validateWeights(weights); err != nil {
	// 	return nil, fmt.Errorf("invalid %s weights: %v", getModeName(isSpecial), err)
	// }

	return weights, nil
}
