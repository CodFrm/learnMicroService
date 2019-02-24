package commands

import (
	"github.com/CodFrm/learnMicroService/ddd"
)

func CommandBus(command ddd.CommandMessage) error {
	return command.ResolveHandler()
}
