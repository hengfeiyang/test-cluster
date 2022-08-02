package cluster

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"strconv"
	"strings"

	"test-cluster/config"
	"test-cluster/etcd"

	"github.com/hashicorp/memberlist"
	"github.com/rs/zerolog/log"
)

type Cluster struct {
	meta          *meta
	Memberlist    *memberlist.Memberlist
	stopWatching  chan bool
	EtcdClient    *etcd.EtcdStorage
	LocalNodeName string
}

type Item struct {
	Ip     string `json:"ip"`
	Status string `json:"status"`
}

func NewCluster(meta *meta) *Cluster {
	return &Cluster{meta: meta}
}

func (c *Cluster) Join() {
	log.Printf("Inside Join Cluster")
	cfg := memberlist.DefaultLocalConfig()
	name, _ := os.Hostname()
	name += "-" + config.Global.Cluster.NodeID + "-" + config.Global.Cluster.NodePort
	cfg.Name = name
	cfg.BindAddr = "127.0.0.1"
	port, _ := strconv.Atoi(config.Global.Cluster.NodePort)
	cfg.BindPort = port

	h := md5.New()
	h.Write([]byte(config.Global.Cluster.Name))
	cfg.SecretKey = []byte(hex.EncodeToString(h.Sum(nil)))
	log.Printf("Secret key: %s, len: %d", cfg.SecretKey, len(cfg.SecretKey))

	cfg.Events = NewNodeEventDelegate()
	cfg.Delegate = NewNodeMetadataDelegate(c.meta, cfg.Name)

	ml, err := memberlist.Create(cfg)
	if err != nil {
		panic(err)
	}
	c.Memberlist = ml

	addrs := strings.Split(config.Global.Cluster.Hosts, ",")
	log.Printf("Join addrs", addrs)
	n, err := ml.Join(addrs)
	if err != nil {
		panic("Failed to join cluster: " + err.Error())
	}
	log.Printf("Joined the cluster: %d", n)
	log.Printf("cluster is up. URL: %s", c.Memberlist.LocalNode().Address())

	c.EtcdClient = etcd.New(config.Global.Etcd.Prefix)
	c.EtcdClient.Join(cfg.Name)
	c.LocalNodeName = cfg.Name
}

func (c *Cluster) Stop() error {
	c.stopWatching <- true

	return nil
}
