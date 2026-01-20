package client

import (
	"context"

	v1 "github.com/card-engine/game_common/api/history/v1"

	google_grpc "google.golang.org/grpc"
)

// 获取游戏下注配置详情
func GetGameBetConfig(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.GetGameBetConfigRequest) (*v1.GetGameBetConfigReply, error) {
	client := v1.NewGameBetConfigApiClient(grpcClient)
	return client.GetGameBetConfig(ctx, req)
}
