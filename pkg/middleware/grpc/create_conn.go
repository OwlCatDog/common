package middleware

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/registry"
	kratosGrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/heyinLab/common/pkg/common"
	"google.golang.org/grpc"
)

// createGRPCConn 创建 gRPC 连接
func CreateGRPCConn(config *common.ServiceConfig, discovery registry.Discovery, logger *log.Helper) (*grpc.ClientConn, error) {
	opts := []kratosGrpc.ClientOption{
		kratosGrpc.WithEndpoint(config.Endpoint),
		kratosGrpc.WithTimeout(config.Timeout),
		kratosGrpc.WithMiddleware(
			recovery.Recovery(),
			ForwardClaims(),
		),
	}

	// 如果有服务发现，添加服务发现选项
	if discovery != nil {
		opts = append(opts, kratosGrpc.WithDiscovery(discovery))
	}

	conn, err := kratosGrpc.DialInsecure(
		context.Background(),
		opts...,
	)
	if err != nil {
		return nil, err
	}

	logger.Infof("平台服务客户端连接成功: endpoint=%s, timeout=%v", config.Endpoint, config.Timeout)

	return conn, nil
}
