package mailserver

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	v1 "github.com/heyinLab/common/api/gen/go/mail/v1"
	middleware "github.com/heyinLab/common/pkg/middleware/grpc"
	"google.golang.org/grpc"
)

type Client struct {
	config *Config
	conn   *grpc.ClientConn
	logger *log.Helper

	// 子服务客户端
	mailClient *MailClient
}

func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	logger := log.NewHelper(log.With(
		log.GetLogger(),
		"module", "mail-client",
	))

	conn, err := middleware.CreateGRPCConn(config, nil, logger)
	if err != nil {
		return nil, fmt.Errorf("创建 gRPC 连接失败: %w", err)
	}

	return &Client{
		config:     config,
		conn:       conn,
		logger:     logger,
		mailClient: newMailClient(conn, logger),
	}, nil
}

func NewClientWithDiscovery(config *Config, discovery registry.Discovery) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if discovery == nil {
		return nil, fmt.Errorf("服务发现实例不能为空")
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	logger := log.NewHelper(log.With(
		log.GetLogger(),
		"module", "mail-client",
	))

	conn, err := middleware.CreateGRPCConn(config, discovery, logger)
	if err != nil {
		return nil, fmt.Errorf("创建 gRPC 连接失败: %w", err)
	}

	logger.Infof("平台服务客户端连接成功 (服务发现): endpoint=%s, timeout=%v", config.Endpoint, config.Timeout)

	return &Client{
		config:     config,
		conn:       conn,
		logger:     logger,
		mailClient: newMailClient(conn, logger),
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) MailClient() *MailClient {
	return c.mailClient
}

type MailClient struct {
	client v1.MailInternalServiceClient
	logger *log.Helper
}

func newMailClient(conn *grpc.ClientConn, logger *log.Helper) *MailClient {
	return &MailClient{
		client: v1.NewMailInternalServiceClient(conn),
		logger: logger,
	}
}

func (c *MailClient) SendTemplateMail(ctx context.Context, req *v1.InternalSendEmailRequest) (*v1.InternalSendEmailResponse, error) {
	resp, err := c.client.InternalSendEmail(ctx, req)

	if err != nil {
		c.logger.WithContext(ctx).Errorf("发送邮件失败, opt=%v, err=%v", req, err)
		return nil, err
	}

	return resp, nil
}
