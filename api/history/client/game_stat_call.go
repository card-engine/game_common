package client

import (
	"context"

	v1 "github.com/card-engine/game_common/api/history/v1"

	google_grpc "google.golang.org/grpc"
)

// 用户统计
func GetUserStat(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.GetUserStatRequest) (*v1.GetUserStatReply, error) {
	client := v1.NewGameStatApiClient(grpcClient)
	return client.GetUserStat(ctx, req)
}
