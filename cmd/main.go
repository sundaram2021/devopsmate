package main

import (
	"devopsmate/pkg"
)

const (
	apiKey     = "1hfjEZFogUnUcRcGljhPqL5nEhsm7GWBbYYHB3sel7j1YkPl5S"
	regionCode = "NYC1"
	sshKeyPath = "civo"
)

func main() {
	pkg.CreateComputeInstance(apiKey, regionCode, sshKeyPath)
}
