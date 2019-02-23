package main

import (
	"github.com/CodFrm/learnMicroService/auth"
	"github.com/CodFrm/learnMicroService/post"
	"os"
)

func main() {
	if len(os.Args) < 1 {
		println("auth post test")
		return
	}
	switch os.Args[1] {
	case "post":
		{
			post.Start()
			break
		}
	case "auth":
		{
			auth.Start()
			break
		}
	default:
		{

		}
	}
	// post := &commands.PostCommand{Uid: 1, Title: "哈哈"}

	// commands.CommandBus(post)
}
