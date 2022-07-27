package cluster

import (
	"fmt"

	"github.com/hashicorp/memberlist"
	"github.com/rs/zerolog/log"
)

type NodeState int

const (
	NodeStateUnknown NodeState = iota
	NodeStateAlive
	NodeStateSuspect
	NodeStateDead
	NodeStateLeft
)

// Enum value maps for NodeState.
var NodeState_name = map[NodeState]string{
	NodeStateUnknown: "unknown",
	NodeStateAlive:   "alive",
	NodeStateSuspect: "suspect",
	NodeStateDead:    "dead",
	NodeStateLeft:    "left",
}

type NodeEventDelegate struct{}

func NewNodeEventDelegate() *NodeEventDelegate {
	return &NodeEventDelegate{}
}

func (d *NodeEventDelegate) NotifyJoin(node *memberlist.Node) {
	log.Debug().Str("event", "join").Str("node", fmt.Sprintf("%s (%s)", node.FullAddress().Name, node.FullAddress().Addr)).Msg("")
}
func (d *NodeEventDelegate) NotifyLeave(node *memberlist.Node) {
	log.Debug().Str("event", "leave").Str("node", fmt.Sprintf("%s (%s)", node.FullAddress().Name, node.FullAddress().Addr)).Msg("")
}
func (d *NodeEventDelegate) NotifyUpdate(node *memberlist.Node) {
	log.Debug().Str("event", "update").Str("node", fmt.Sprintf("%s (%s)", node.FullAddress().Name, node.FullAddress().Addr)).Msg("")
}

func makeNodeState(state memberlist.NodeStateType) NodeState {
	switch state {
	case memberlist.StateAlive:
		return NodeStateAlive
	case memberlist.StateSuspect:
		return NodeStateSuspect
	case memberlist.StateDead:
		return NodeStateDead
	case memberlist.StateLeft:
		return NodeStateLeft
	default:
		return NodeStateUnknown
	}
}

type NodeMetadataDelegate struct {
	meta     *meta
	nodeName string
}

func NewNodeMetadataDelegate(meta *meta, nodeName string) *NodeMetadataDelegate {
	return &NodeMetadataDelegate{meta, nodeName}
}

func (d *NodeMetadataDelegate) NodeMeta(limit int) []byte {
	log.Debug().Str("delegate", "node_meta").Int("limit", limit).Msg("")
	return d.meta.GetNodeMeta(d.nodeName)
}

func (d *NodeMetadataDelegate) LocalState(join bool) []byte {
	log.Debug().Str("delegate", "local_state").Bool("join", join).Msg("")
	return []byte{}
}
func (d *NodeMetadataDelegate) NotifyMsg(msg []byte) {
	log.Debug().Str("delegate", "notify_msg").Str("data", string(msg)).Msg("")
}

func (d *NodeMetadataDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	// log.Debug().Str("delegate", "get_broadcasts").Msg("")
	return [][]byte{}
}

func (d *NodeMetadataDelegate) MergeRemoteState(buf []byte, join bool) {
	log.Debug().Str("delegate", "merge_remote_state").Str("buf", string(buf)).Bool("join", join).Msg("")
}
