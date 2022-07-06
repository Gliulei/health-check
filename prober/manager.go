package prober

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"health-check/library/errors"
	"health-check/log"
	"io"
)

var Manager = manager{
	groups: make(map[string]*Group, 0),
}

type manager struct {
	groups map[string]*Group
}

func (m *manager) Persist(sink raft.SnapshotSink) error {
	data, err := json.Marshal(m.groups)
	if err != nil {
		return err
	}
	sink.Write(data)
	sink.Close()
	return nil
}

func (m *manager) Release() {}

func (m *manager) Restore(reader io.ReadCloser) error {
	if err := json.NewDecoder(reader).Decode(&m.groups); err != nil {
		return err
	}

	for _, g:= range m.groups {
		if g.running {
			log.Warn("group %s is running ...", g.Name)
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		g.InitGroup(cancel)
		go g.Run(ctx)
		log.Notice("group %s restore running success", g.Name)
	}

	return nil
}

func (m *manager) AddInstance(groupName string, ip string, port int, probeType string, url string) error {
	addr := fmt.Sprintf("%s:%d", ip, port)
	g, ok := m.groups[groupName]

	var isNeedCreate bool
	if ok {
		_, ok := g.Instances[addr]
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
		return g.Instances, nil
	}

	return nil, errors.ErrInstanceHasExists
}

func (m *manager) GetInstanceOne(groupName string) ([]map[string]interface{}, error) {
	g, ok := m.groups[groupName]
	if ok {
		var data []map[string]interface{}
		for _, val := range g.Instances {
			tmp := map[string]interface{}{
				"ip":   val.Ip,
				"port": val.Port,
			}
			data = append(data, tmp)
		}
		return data, nil
	}

	return nil, errors.ErrInstanceHasExists
}
