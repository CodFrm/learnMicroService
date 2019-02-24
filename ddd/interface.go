package ddd

type CommandMessage interface {
	ResolveHandler() error
}

type EventMessage interface {
	GetGroupId() string
	GetEventNames() []string
	Handler(name string, value []byte) error
}
