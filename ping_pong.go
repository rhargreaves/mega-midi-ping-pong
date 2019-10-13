package main

import (
	"fmt"
	"os"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/portmididrv"
)

func main() {
	drv, err := portmididrv.New()
	exitOnError(err)
	defer drv.Close()

	ins, err := drv.Ins()
	exitOnError(err)

	outs, err := drv.Outs()
	exitOnError(err)

	if len(os.Args) == 2 && os.Args[1] == "list" {
		printInPorts(ins)
		printOutPorts(outs)
		return
	}
}

func exitOnError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func printPort(port mid.Port) {
	fmt.Printf("[%v] %s\n", port.Number(), port.String())
}

func printInPorts(ports []mid.In) {
	fmt.Printf("MIDI IN Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	fmt.Printf("\n\n")
}

func printOutPorts(ports []mid.Out) {
	fmt.Printf("MIDI OUT Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	fmt.Printf("\n\n")
}
