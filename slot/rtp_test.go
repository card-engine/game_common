package slot

// "pg_api_server/utils"

// func TestCacheSimpleByRate(t *testing.T) {
// 	config := utils.LoadConfig("../config/server.yaml")

// 	redisClient := utils.NewRedisClient(config.ApiServer.Memcached)
// 	db, err := utils.GetGormDb(config.ApiServer.DB.Aigc.Type, config.ApiServer.DB.Aigc.DSN)
// 	if err != nil {
// 		t.Fatalf("failed to connect database: %v", err)
// 	}

// 	rtp := NewRtp(db, redisClient)
// 	rtp.ClearCacheByRate(98)
// 	rtp.CacheSimpleByRate(98)

// 	specialWeights, err := rtp.LoadWeightsFromJSON(98, true)
// 	if err != nil {
// 		t.Fatalf("Failed to load special weights: %v", err)
// 	}
// 	normalWeights, err := rtp.LoadWeightsFromJSON(98, false)
// 	if err != nil {
// 		t.Fatalf("Failed to load normal weights: %v", err)
// 	}

// 	allSpecialRates, err := rtp.GetSimpleAllRate(98, true)
// 	if err != nil {
// 		t.Fatalf("Failed to get all special rates: %v", err)
// 	}
// 	_, err = rtp.VerifyAllRatesHaveWeights(allSpecialRates, specialWeights)
// 	if err != nil {
// 		t.Fatalf("Special rates missing weights: %v", err)
// 	}

// 	allNormalRates, err := rtp.GetSimpleAllRate(98, false)
// 	if err != nil {
// 		t.Fatalf("Failed to get all normal rates: %v", err)
// 	}
// 	_, err = rtp.VerifyAllRatesHaveWeights(allNormalRates, normalWeights)
// 	if err != nil {
// 		t.Fatalf("Normal rates missing weights: %v", err)
// 	}

// 	var (
// 		specialNum, normalNum  int
// 		specialBet, specialWin float64
// 		normalBet, normalWin   float64
// 		totalBet, totalWin     float64
// 		normalWinCount         int
// 		totalSpins             = 1000000
// 		mu                     sync.Mutex
// 		wg                     sync.WaitGroup
// 	)

// 	// 控制最大并发为 30
// 	concurrencyLimit := 30
// 	semaphore := make(chan struct{}, concurrencyLimit)

// 	// var records []models.PgSpinData

// 	// tableName := "pg_spin_" + strconv.Itoa(98)
// 	// err = db.Table(tableName).Select("id,totalWin,bet,rate,GameType").Find(&records).Error
// 	// if err != nil {
// 	// 	t.Fatalf("Special rates missing weights: %v", err)
// 	// }

// 	// // 以id为key索引
// 	// idToRecord := make(map[uint64]models.PgSpinData)
// 	// for _, rec := range records {
// 	// 	idToRecord[rec.ID] = rec
// 	// }

// 	// // 使用 map 聚合 redis key -> set of values
// 	// rateSetMap := make(map[string][]uint64)

// 	// for _, rec := range records {
// 	// 	var rateKey string
// 	// 	rateStr := fmt.Sprintf("%.2f", rec.Rate)

// 	// 	if rec.GameType == 1 {
// 	// 		rateKey = fmt.Sprintf("%s_special:rate:%s", tableName, rateStr)
// 	// 	} else {
// 	// 		rateKey = fmt.Sprintf("%s_normal:rate:%s", tableName, rateStr)
// 	// 	}

// 	// 	// 聚合 ID 到对应的 rateKey
// 	// 	rateSetMap[rateKey] = append(rateSetMap[rateKey], rec.ID)
// 	// }

// 	// //节省内存，释放records
// 	// records = nil

// 	startTime := time.Now()

// 	for i := 0; i < totalSpins; i++ {
// 		semaphore <- struct{}{} // 占用一个槽位
// 		wg.Add(1)

// 		go func() {
// 			defer func() {
// 				<-semaphore // 释放槽位
// 				wg.Done()
// 			}()

// 			if rtp.IsSpecialModeTriggered() {
// 				rate := rtp.GetRateByWeight(specialWeights)
// 				// rateSetMapKey := fmt.Sprintf("%s_special:rate:%.2f", tableName, rate)
// 				// rateSet := rateSetMap[rateSetMapKey]
// 				// id := rateSet[rand.IntN(len(rateSet))]
// 				// ret := idToRecord[id]

// 				ret, err := rtp.GetOneSimpleByRate(98, true, rate)
// 				if err != nil {
// 					t.Errorf("GetOneSimpleByRate (special) error: %v", err)
// 					return
// 				}
// 				mu.Lock()
// 				specialNum++
// 				specialBet += float64(ret.Bet)
// 				specialWin += ret.TotalWin
// 				totalBet += float64(ret.Bet)
// 				totalWin += ret.TotalWin
// 				mu.Unlock()
// 			} else {
// 				rate := rtp.GetRateByWeight(normalWeights)
// 				// rateSetMapKey := fmt.Sprintf("%s_normal:rate:%.2f", tableName, rate)
// 				// rateSet := rateSetMap[rateSetMapKey]
// 				// id := rateSet[rand.IntN(len(rateSet))]
// 				// ret := idToRecord[id]
// 				ret, err := rtp.GetOneSimpleByRate(98, false, rate)
// 				if err != nil {
// 					t.Errorf("GetOneSimpleByRate (normal) error: %v", err)
// 					return
// 				}
// 				mu.Lock()
// 				if ret.TotalWin > 0 {
// 					normalWinCount++
// 				}
// 				normalNum++
// 				normalBet += float64(ret.Bet)
// 				normalWin += ret.TotalWin
// 				totalBet += float64(ret.Bet)
// 				totalWin += ret.TotalWin
// 				mu.Unlock()
// 			}
// 		}()
// 	}
// 	wg.Wait()

// 	// 输出统计信息
// 	elapsed := time.Since(startTime)
// 	t.Logf("耗时: %s", elapsed)
// 	t.Logf("总 SPIN 次数: %d", totalSpins)
// 	t.Logf("特殊玩法触发次数: %d", specialNum)
// 	t.Logf("普通玩法次数: %d", normalNum)

// 	t.Logf("特殊玩法总押注: %.2f", specialBet)
// 	t.Logf("特殊玩法总赢钱: %.2f", specialWin)
// 	t.Logf("特殊玩法 RTP: %.4f", specialWin/specialBet)
// 	t.Logf("特殊玩法触发率: %.4f", float64(specialNum)/float64(totalSpins))

// 	t.Logf("普通玩法总押注: %.2f", normalBet)
// 	t.Logf("普通玩法总赢钱: %.2f", normalWin)
// 	t.Logf("普通玩法中线次数（赢钱次数）: %d", normalWinCount)
// 	t.Logf("普通玩法 RTP: %.4f", normalWin/normalBet)
// 	t.Logf("普通玩法中线率: %.4f", float64(normalWinCount)/float64(totalSpins))

// 	t.Logf("总押注: %.2f", totalBet)
// 	t.Logf("总赢钱: %.2f", totalWin)
// 	t.Logf("总 RTP: %.4f", totalWin/totalBet)
// }
