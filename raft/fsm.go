package craft

import (
	"encoding/json"
	"github.com/hashicorp/raft"
	"health-check/library"
	"health-check/log"
	"health-check/prober"
	"io"
	"sync"
)

type Fsm struct {
	DataBase database
}

type StoreCommand struct {
	Op string `json:"op,omitempty"`
	//Value []byte `json:"value,omitempty"`
	Value library.GroupForm `json:"value,omitempty"`
}

func NewFsm() *Fsm {
	fsm := &Fsm{
		DataBase: NewDatabase(),
	}
	return fsm
}

func (f *Fsm) Apply(l *raft.Log) interface{} {

	var c StoreCommand
	err := json.Unmarshal(l.Data, &c)
	if err != nil {
		return err
	}

	g := c.Value

	log.Notice("apply data: %+v", g)

	switch c.Op {
	case "add":
		err := prober.Manager.AddInstance(g.GroupName, g.Ip, g.Port, g.ProbeType, g.Url)
		if err != nil {
			return err
		}
		break
	case "remove":
		err := prober.Manager.RemoveInstance(g.GroupName, g.Ip, g.Port)
		if err != nil {
			return err
		}
		break
	}

	return nil
}

func (f *Fsm) Snapshot() (raft.FSMSnapshot, error) {
	return &prober.Manager, nil
	//return &f.DataBase, nil
}

func (f *Fsm) Restore(reader io.ReadCloser) error {
	return prober.Manager.Restore(reader)
}

type database struct {
	Data map[string]string
	mu   sync.Mutex
}

func NewDatabase() database {
	return database{
		Data: make(map[string]string),
	}
}

func (d *database) Get(key string) string {
	d.mu.Lock()
	value := d.Data[key]
	d.mu.Unlock()
	return value
}

func (d *database) Set(key, value string) {
	d.mu.Lock()
	d.Data[key] = value
	d.mu.Unlock()
}

func (d *database) Persist(sink raft.SnapshotSink) error {
	d.mu.Lock()
	data, err := json.Marshal(d.Data)
	d.mu.Unlock()
	if err != nil {
		return err
	}
	sink.Write(data)
	sink.Close()
	return nil
}

func (d *database) Release() {}
