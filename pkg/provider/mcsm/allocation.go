package mcsm

import (
	"fmt"

	"github.com/kasefuchs/lazygate/pkg/provider"
	"github.com/kasefuchs/lazygate/pkg/utils"
)

var _ provider.Allocation = (*Allocation)(nil)

// Allocation represents Docker provider item.
type Allocation struct {
	item           *item
	client         *Client
	configFilePath string
}

func NewAllocation(client *Client, item *item) *Allocation {
	return &Allocation{
		item:   item,
		client: client,
	}
}

func (a *Allocation) Stop() error {
	return a.client.ServerStop(a.item.instance.Id, a.client.daemon_id)
}

func (a *Allocation) Start() error {
	return a.client.ServerStart(a.item.instance.Id, a.client.daemon_id)
}

func (a *Allocation) State() provider.AllocationState {
	inst, err := a.client.GetInstance(a.item.instance.Id, a.client.daemon_id)
	if err != nil {
		fmt.Errorf("failed to get isntance for state %v", err)
	}
	switch inst.Data.Status {
	case 0:
		return provider.AllocationStateStopped
	case 3:
		return provider.AllocationStateStarted
	case -1:
	case 1:
	case 2:
		return provider.AllocationStateUnknown
	}
	return provider.AllocationStateUnknown
}

func (a *Allocation) ParseConfig(cfg interface{}, rootLabel string) (interface{}, error) {
	inst, err := a.client.GetInstance(a.item.instance.Id, a.client.daemon_id)
	if err != nil {
		return nil, err
	}
	if rootLabel == utils.ChildLabel("allocation") {
		cfg.(*provider.AllocationConfig).Server = inst.Data.Config.Nickname
	}
	cfg, err = utils.ParseTags(inst.Data.Config.Tag, cfg, rootLabel)
	if err != nil {
		return nil, fmt.Errorf("Error parsing tags for mcsm: %v", err)
	}

	return cfg, nil
}
