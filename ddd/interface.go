package ddd

type CommandMessage interface {
	ResolveHandler() error
}
