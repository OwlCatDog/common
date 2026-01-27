package mailserver

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	consulapi "github.com/hashicorp/consul/api"
	v1 "github.com/heyinLab/common/api/gen/go/mail/v1"
)

func TestSendMain(t *testing.T) {
	//t.Skip()
	config := consulapi.DefaultConfig()
	config.Address = "192.168.3.6:8500"
	config.Token = ""
	config.Datacenter = "dc1"
	config.Scheme = "http"
	// 创建 Consul 客户端
	consulClient, err := consulapi.NewClient(config)
	if err != nil {
		t.Skipf("无法连接到 Consul: %v", err)
		return
	}

	// 创建 Consul 服务发现
	discovery := consul.New(consulClient)

	// 创建平台服务客户端
	client, err := NewClientWithDiscovery(DefaultConfig(), discovery)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	_, err = client.MailClient().SendTemplateMail(ctx, &v1.InternalSendEmailRequest{
		TriggerEvent: v1.TriggerEvent_InvitationEmail,
		DestAddr:     "a@b.c",
		CountryCode:  "CN",
		Language:     "zh-CN",
		Variables: map[string]string{
			"test": "123",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
