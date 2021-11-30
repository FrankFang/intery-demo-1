package main

import (
	"intery/cmd"
	"intery/initializers"
)

func main() {
	initializers.LoadEnv()
	cmd.Execute()
}
