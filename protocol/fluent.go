package protocol

import "github.com/tinylib/msgp/msgp"

//go:generate msgp -tests=false

//msgp:tuple FluentMsg
type FluentMsg struct {
	Tag    msgp.Raw
	Time   int64
	Record FluentRecord
	Option map[string]interface{}
}

//msgp:tuple FluentRawMsg
type FluentRawMsg struct {
	Tag    msgp.Raw
	Time   msgp.Raw
	Record msgp.Raw
	Option msgp.Raw
}

// - Unmapped fields are ignored
// - Using 'raw' to avoid allocations
type FluentRecord struct {
	//ContainerId   string `msg:"container_id"`
	ContainerName msgp.Raw `msg:"container_name"`
	Log           msgp.Raw `msg:"log"`
	//Source        string `msg:"source"`
}