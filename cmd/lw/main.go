package main

import "os"

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		switch args[0] {
		case "review":
			cmdReview(args[1:])
			return
		case "done":
			cmdDone(args[1:])
			return
		}
	}

	cmdCreate(args)
}
