package utils

import (
	"context"
	"math"
	"math/rand/v2"
	"strconv"

	rtp_rpc_v1 "github.com/card-engine/game_common/api/rtp/v1"
	rtp_rpc_client "github.com/card-engine/game_common/api/rtp/v1/client"
	google_grpc "google.golang.org/grpc"
)

// 飞机类游戏的方法,计算出坠机的概率
func CalculateCrashX(
	rtpGrpcConn *google_grpc.ClientConn,
	appId string,
	gameBrand string,
	gameId string,
	roundId string,
	rtp string,
) (float64, error) {
	rtpNum, err := strconv.ParseFloat(rtp, 64)
	if err != nil {
		return 1, nil
	}
	resp, err := rtp_rpc_client.GetBaccaratRtp(context.Background(), rtpGrpcConn, &rtp_rpc_v1.GetBaccaratRtpRequest{
		AppId:     appId,
		GameBrand: gameBrand,
		GameId:    gameId,
		RoundId:   roundId,
		Rtp:       rtp,
	})
	if err != nil {
		return 1, err
	}

	// 计算坠机的比例
	if resp.Rate1 > 0 || resp.Rate2 > 0 {
		randomValue := rand.Float64()
		if randomValue < resp.Rate1 && resp.Rate1 > 0 {
			return 1, nil
		}
		if randomValue < resp.Rate2 && resp.Rate2 > 0 {
			return 2, nil
		}
	}

	var crash float64 = 0
	if int32(rtpNum) == 50 {
		crash = min(100, rtpNum*500/(50000-RandFloat(0, 49999)))
	} else if int32(rtpNum) == 65 {
		crash = min(200, rtpNum*500/(50000-RandFloat(0, 49999)))
	} else if int32(rtpNum) == 75 {
		crash = min(500, rtpNum*500/(50000-RandFloat(0, 49999)))
	} else if int32(rtpNum) == 85 {
		crash = min(750, rtpNum*500/(50000-RandFloat(0, 49999)))
	} else if int32(rtpNum) == 90 {
		crash = min(1000, rtpNum*500/(50000-RandFloat(0, 49999)))
	} else {
		//容错的旧算法
		crash = rtpNum * 500 / (50000 - RandFloat(0, 49999))
	}

	// 向下取整，保留两位小数
	crash = math.Floor(crash*100) / 100

	if crash < 1.0 {
		return 1.0, nil
	}

	return crash, nil
}
