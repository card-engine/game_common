package utils

import (
	v1 "cn.qingdou.server/game_common/api/rtp/v1"
	"context"

	google_grpc "google.golang.org/grpc"
)

// 选取spin结果
func SelectSpin(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.SelectSpinRequest) (*v1.SelectSpinReply, error) {
	client := v1.NewRtpApiClient(grpcClient)
	return client.SelectSpin(ctx, req)
}

// 获取rtp
func GetPlayerRtp(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.GetPlayerRtpRequest) (*v1.GetPlayerRtpReply, error) {
	client := v1.NewRtpApiClient(grpcClient)
	return client.GetPlayerRtp(ctx, req)
}

// 获取baccarat rtp
func GetBaccaratRtp(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.GetBaccaratRtpRequest) (*v1.GetBaccaratRtpReply, error) {
	client := v1.NewRtpApiClient(grpcClient)
	return client.GetBaccaratRtp(ctx, req)
}

// baccarat结算
func SettleBaccarat(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.SettleBaccaratRequest) (*v1.SettleBaccaratReply, error) {
	client := v1.NewRtpApiClient(grpcClient)
	return client.SettleBaccarat(ctx, req)
}
