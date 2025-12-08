package client

import (
	"context"

	v1 "github.com/zuodazuoqianggame/game_common/api/history/v1"

	google_grpc "google.golang.org/grpc"
)

// 获取某一时间段的下注的输赢
func GetTotalBetAndWin(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.GetTotalBetAndWinRequest) (*v1.GetTotalBetAndWinReply, error) {
	client := v1.NewHistoryApiClient(grpcClient)
	return client.GetTotalBetAndWin(ctx, req)
}

func GameHistoryByTime(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.GameHistoryByTimeRequest) (*v1.GameHistoryByTimeReply, error) {
	client := v1.NewHistoryApiClient(grpcClient)
	return client.GameHistoryByTime(ctx, req)
}

func GameHistoryList(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.GameHistoryListRequest) (*v1.GameHistoryListReply, error) {
	client := v1.NewHistoryApiClient(grpcClient)
	return client.GameHistoryList(ctx, req)
}

func GameHistoryDetail(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.GameHistoryDetailRequest) (*v1.GameHistoryDetailReply, error) {
	client := v1.NewHistoryApiClient(grpcClient)
	return client.GameHistoryDetail(ctx, req)
}
