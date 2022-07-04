package craft

import (
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"health-check/library/utils"
	"health-check/log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func NewRaft(raftAddr, raftId, raftDir string) (*raft.Raft, *Fsm, error) {

	if !utils.PathExists(raftDir) {
		if err := os.MkdirAll(raftDir, os.ModePerm); err != nil {
			return nil, nil, log.Warn(err.Error())
		}
	}

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(raftId)

	logStore, _ := raftboltdb.NewBoltStore(filepath.Join(raftDir,
		"raft-log.bolt")) //用来存储raft的日志

	stableStore, _ := raftboltdb.NewBoltStore(filepath.Join(raftDir,
		"raft-stable.bolt")) //稳定存储，用来存储raft集群的节点信息等

	snapshotStore, _ := raft.NewFileSnapshotStore(raftDir, 1, os.Stderr)

	addr, err := net.ResolveTCPAddr("tcp", raftAddr)

	if err != nil {
		return nil, nil, err
	}
	transport, err := raft.NewTCPTransport(raftAddr, addr, 2, 5*time.Second, os.Stderr)

	if err != nil {
		return nil, nil, err
	}

	fsm := NewFsm() //有限状态机
	rf, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, nil, err
	}

	return rf, fsm, err
}

func Bootstrap(rf *raft.Raft, raftId, raftAddr, raftCluster string) {
	servers := rf.GetConfiguration().Configuration().Servers
	if len(servers) > 0 {
		return
	}
	peerArray := strings.Split(raftCluster, ",")
	if len(peerArray) == 0 {
		return
	}

	var configuration raft.Configuration
	for _, peerInfo := range peerArray {
		peer := strings.Split(peerInfo, "/")
		id := peer[0]
		addr := peer[1]
		server := raft.Server{
			ID:      raft.ServerID(id),
			Address: raft.ServerAddress(addr),
		}
		configuration.Servers = append(configuration.Servers, server)
	}
	rf.BootstrapCluster(configuration)
	return
}
