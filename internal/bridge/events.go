package bridge

import "github.com/zishang520/socket.io/v3/pkg/types"

type EventData []byte

type Event types.EventName

const (
	VectorStatus   Event = "vector:ststus"   // Send
	AgentHeartBeat Event = "agent:heartbeat" // Send
	AgentStatus    Event = "agent:status"    // Send
	AgentData      Event = "agent:data"      // Send
	VectorStart    Event = "vector:start"    // Get -> Send
	VectorStop     Event = "vector:stop"     // Get -> Send
	VectorRestart  Event = "vector:restart"  // Get -> Send
	AgentStop      Event = "agent:stop"      // Get -> Send
)
