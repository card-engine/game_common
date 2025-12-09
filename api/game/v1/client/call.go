package utils

import (
	"context"

	v1 "github.com/card-engine/game_common/api/game/v1"
	"github.com/go-kratos/kratos/v2/metadata"
	google_grpc "google.golang.org/grpc"
)

// 获取余额
func Balance(ctx context.Context, grpcClient *google_grpc.ClientConn, appid string, req *v1.BalanceRequest) (*v1.BalanceReply, error) {
	client := v1.NewGameApiClient(grpcClient)
	ctx = metadata.AppendToClientContext(ctx, "x-md-global-appid", appid)
	return client.Balance(ctx, req)
}

// 投注
func Bet(ctx context.Context, grpcClient *google_grpc.ClientConn, appid string, req *v1.BetRequest) (*v1.BetReply, error) {
	client := v1.NewGameApiClient(grpcClient)
	ctx = metadata.AppendToClientContext(ctx, "x-md-global-appid", appid)
	return client.Bet(ctx, req)
}

// 派奖
func Win(ctx context.Context, grpcClient *google_grpc.ClientConn, appid string, req *v1.WinRequest) (*v1.WinReply, error) {
	client := v1.NewGameApiClient(grpcClient)
	ctx = metadata.AppendToClientContext(ctx, "x-md-global-appid", appid)
	return client.Win(ctx, req)
}

// 撤销退款
func Refund(ctx context.Context, grpcClient *google_grpc.ClientConn, appid string, req *v1.RefundRequest) (*v1.RefundReply, error) {
	client := v1.NewGameApiClient(grpcClient)
	ctx = metadata.AppendToClientContext(ctx, "x-md-global-appid", appid)
	return client.Refund(ctx, req)
}

// 游戏历史记录
func GameHistory(ctx context.Context, grpcClient *google_grpc.ClientConn, appid string, req *v1.GameHistoryRequest) (*v1.GameHistoryReply, error) {
	client := v1.NewGameApiClient(grpcClient)
	ctx = metadata.AppendToClientContext(ctx, "x-md-global-appid", appid)
	return client.GameHistoryList(ctx, req)
}

// Jili游戏历史记录
func JiliGameHistory(ctx context.Context, grpcClient *google_grpc.ClientConn, appid string, req *v1.JiliGameHistoryRequest) (*v1.JiliGameHistoryReply, error) {
	client := v1.NewGameApiClient(grpcClient)
	ctx = metadata.AppendToClientContext(ctx, "x-md-global-appid", appid)
	return client.JiliGameHistory(ctx, req)
}

// 商户游戏列表
func AppGameList(ctx context.Context, grpcClient *google_grpc.ClientConn, appid string, req *v1.AppGameListRequest) (*v1.AppGameListReply, error) {
	client := v1.NewGameApiClient(grpcClient)
	ctx = metadata.AppendToClientContext(ctx, "x-md-global-appid", appid)
	return client.AppGameList(ctx, req)
}
