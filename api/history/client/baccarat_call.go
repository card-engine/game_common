package client

import (
	v1 "cn.qingdou.server/game_common/api/history/v1"
	"context"

	google_grpc "google.golang.org/grpc"
)

// 创建百人场游戏记录
func CreateBaccarat(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.CreateBaccaratRequest) (*v1.CreateBaccaratReply, error) {
	client := v1.NewBaccaratApiClient(grpcClient)
	return client.CreateBaccarat(ctx, req)
}

// 获取百人场游戏记录
func GetBaccarat(ctx context.Context, grpcClient *google_grpc.ClientConn, req *v1.GetBaccaratRequest) (*v1.GetBaccaratReply, error) {
	client := v1.NewBaccaratApiClient(grpcClient)
	return client.GetBaccarat(ctx, req)
}
