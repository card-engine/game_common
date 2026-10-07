package client

import (
	"context"

	v1 "github.com/card-engine/game_common/api/history/v1"

	google_grpc "google.golang.org/grpc"
)

// Decide 开奖前询问运营活动。未命中或调用失败时，调用方继续基础 RTP。
func Decide(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.DecideRequest) (*v1.DecideReply, error) {
	return v1.NewControlApiClient(grpcClient).Decide(ctx, req)
}

// Settle 整局结算后入账。saved=false 表示没写入，不代表钱包失败。
func Settle(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.SettleRequest) (*v1.SettleReply, error) {
	return v1.NewControlApiClient(grpcClient).Settle(ctx, req)
}

// Release 退款或超时，放开未结算占用。
func Release(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.ReleaseRequest) (*v1.ReleaseReply, error) {
	return v1.NewControlApiClient(grpcClient).Release(ctx, req)
}
