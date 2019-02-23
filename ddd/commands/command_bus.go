package commands

import (
	"github.com/CodFrm/learnMicroService/ddd"
)

func CommandBus(command ddd.CommandMessage) {
	command.ResolveHandler()
}
