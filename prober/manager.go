package prober

import (
	"context"
	"fmt"
	"health-check/library/errors"
)

var Manager = manager{
	groups: make(map[string]Group, 0),
}

type manager struct {
	groups map[string]Group
}

func (m *manager) AddInstance(groupName string, ip string, port int, probeType string, url string) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	g, ok := m.groups[groupName]

	var isNeedCreate bool
	if ok {
		_, ok := g.instances[addr]
		if ok {
			return errors.ErrInstanceHasExists
		}
		g.cancel()
		<-g.done
	} else {
		isNeedCreate = true
	}

	ctx, cancel := context.WithCancel(context.Background())
	if isNeedCreate {
		g = NewGroup(groupName, probeType, url, cancel)
		m.groups[groupName] = g
	}

	//again install cancel
	g.cancel = cancel

	//add instance
	g.AddInstance(ip, port)

	//start
	go g.Run(ctx)

	return nil
}

func (m *manager) RemoveInstance(groupName string, ip string, port int) error {
	g, ok := m.groups[groupName]
	if ok {
		return g.RemoveInstance(ip, port)
	}

	return errors.ErrInstanceHasExists
}

func (m *manager) GetInstance(groupName string) (map[string]instance, error) {
	g, ok := m.groups[groupName]
	if ok {
		return g.instances, nil
	}

	return nil, errors.ErrInstanceHasExists
}

func (m *manager) GetInstanceOne(groupName string) ([]map[string]interface{}, error) {
	g, ok := m.groups[groupName]
	if ok {
		var data []map[string]interface{}
		for _, val := range g.instances {
			tmp := map[string]interface{}{
				"ip":   val.ip,
				"port": val.port,
			}
			data = append(data, tmp)
		}
		return data, nil
	}

	return nil, errors.ErrInstanceHasExists
}
