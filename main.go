package main

import (
	"os"

	"github.com/CodFrm/learnMicroService/auth"
	"github.com/CodFrm/learnMicroService/post"
)

func main() {
	if len(os.Args) < 2 {
		println("auth post")
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
}
