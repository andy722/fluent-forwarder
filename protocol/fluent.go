package protocol

import "github.com/tinylib/msgp/msgp"

//go:generate msgp -tests=false

// FluentMsg is a single Message Event of Forward protocol
//msgp:tuple FluentMsg
//msgp:decode ignore FluentMsg
//msgp:encode ignore FluentMsg
//msgp:marshal ignore FluentMsg
type FluentMsg struct {
	Tag    msgp.Raw
	Time   int64
	Record FluentRecord
	Option map[string]interface{}
}

// FluentRawMsg uses raw byte buffers for efficient caching
// and zero allocations in unmarshalling
//msgp:tuple FluentRawMsg
//msgp:encode ignore FluentRawMsg
//msgp:unmarshal ignore FluentRawMsg
type FluentRawMsg struct {
	Tag    msgp.Raw
	Time   msgp.Raw
	Record msgp.Raw
	Option msgp.Raw
}

// FluentRecord is a subset of record hash such as:
// - Unmapped fields are ignored
// - Using 'raw' to avoid allocations
type FluentRecord struct {
	//ContainerId   string `msg:"container_id"`
	ContainerName msgp.Raw `msg:"container_name"`
	Log           msgp.Raw `msg:"log"`
	//Source        string `msg:"source"`
}
