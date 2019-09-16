package main

import (
	"fmt"
	"os"

	"gitlab.com/gomidi/midi/mid"
	"gitlab.com/gomidi/portmididrv"
)

func main() {
	fmt.Println("Hello, World!")

	drv, err := portmididrv.New()
	if err != nil {
		panic(err.Error())
	}

	// make sure to close all open ports at the end
	defer drv.Close()

	ins, err := drv.Ins()
	if err != nil {
		panic(err.Error())
	}

	outs, err := drv.Outs()
	if err != nil {
		panic(err.Error())
	}

	if len(os.Args) == 2 && os.Args[1] == "list" {
		printInPorts(ins)
		printOutPorts(outs)
		return
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
