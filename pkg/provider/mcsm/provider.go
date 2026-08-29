package mcsm

import (
	"context"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/kasefuchs/lazygate/pkg/provider"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/netutil"
)

var _ provider.Provider = (*Provider)(nil)

const name = "mcsm"

// Provider based on Docker API.
type Provider struct {
	log    logr.Logger     // Provider logger.
	ctx    context.Context // Provider context.
	client *Client         // PufferPanel API client.
	config *Config         // Provider config.
}

func (p *Provider) Name() string {
	return name
}

func (p *Provider) DefaultConfig() interface{} {
	return &Config{}
}

func (p *Provider) initClient() {
	client, err := NewClient(p.ctx, p.config.BaseUrl, p.config.ApiKey, p.config.DaemonId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initalizing client: %v", err)
		os.Exit(1)
	}
	p.client = client
}

func (p *Provider) Init(opt *provider.InitOptions) error {
	p.log = logr.FromContextOrDiscard(opt.Ctx).WithName(provider.LogName)
	p.ctx = opt.Ctx
	p.config = opt.Config.(*Config)

	p.initClient()

	p.log.Info("initialized")
	return nil
}

func (p *Provider) AllocationGet(srv proxy.RegisteredServer) (provider.Allocation, error) {
	allocs, err := p.AllocationList()
	if err != nil {
		return nil, err
	}

	for _, alloc := range allocs {
		cfg, err := provider.ParseAllocationConfig(alloc)
		if err != nil {
			continue
		}

		if cfg.Server == srv.ServerInfo().Name() {
			return alloc, nil
		}
	}

	return nil, fmt.Errorf("no allocation found")
}

// ServerList returns all registered servers from MCSM with their configuration.
func (p *Provider) ServerList() ([]proxy.ServerInfo, error) {
	search, err := p.client.ServerSearch()
	if err != nil {
		return nil, err
	}

	var servers []proxy.ServerInfo
	for _, inst := range search.Data.Data {
		if !slices.Contains(inst.Config.Tag, "lazy") {
			continue
		}
		if inst.Config.Nickname == "" || inst.Config.PingConfig.IP == "" || inst.Config.PingConfig.Port == 0 {
			continue
		}

		addr := netutil.NewAddr(
			net.JoinHostPort(inst.Config.PingConfig.IP, strconv.Itoa(inst.Config.PingConfig.Port)),
			"tcp",
		)
		servers = append(servers, proxy.NewServerInfo(inst.Config.Nickname, addr))
	}

	return servers, nil
}

func (p *Provider) AllocationList() ([]provider.Allocation, error) {
	items, err := p.itemList()
	if err != nil {
		return nil, err
	}

	allocs := make([]provider.Allocation, len(items))
	for i, it := range items {
		allocs[i] = NewAllocation(p.client, it)
	}

	return allocs, nil
}
