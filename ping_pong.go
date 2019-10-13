package main

import (
	"fmt"
	"os"
	"strconv"

	"gitlab.com/gomidi/midi/mid"
	driver "gitlab.com/gomidi/rtmididrv"
)

func main() {
	drv, err := driver.New()
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

	if len(os.Args) == 3 {
		portInNum, err := strconv.Atoi(os.Args[1])
		exitOnError(err)
		portOutNum, err := strconv.Atoi(os.Args[2])
		exitOnError(err)
		pingPong(ins[portInNum], outs[portOutNum])
		return
	}
}

func pingPong(in mid.In, out mid.Out) {
	fmt.Printf("In: %s\n", in.String())
	fmt.Printf("Out: %s\n", out.String())

	pingSysEx := []byte{0x00, 0x22, 0x77, 0x01}

	err := out.Open()
	exitOnError(err)
	defer out.Close()
	err = in.Open()
	exitOnError(err)
	defer in.Close()

	wr := mid.ConnectOut(out)
	err = wr.SysEx(pingSysEx)
	exitOnError(err)
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
