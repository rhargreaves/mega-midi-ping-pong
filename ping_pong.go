package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rakyll/portmidi"
)

func main() {
	portmidi.Initialize()

	if len(os.Args) == 2 && os.Args[1] == "list" {
		for id := 0; id < portmidi.CountDevices(); id++ {
			deviceID := portmidi.DeviceID(id)
			inf := portmidi.Info(deviceID)
			fmt.Printf("ID: %d\nName: %s\nInput: %t\nOutput: %t\n",
				id,
				inf.Name, inf.IsInputAvailable, inf.IsOutputAvailable)
		}
		return
	}

	if len(os.Args) == 3 {
		portInNum, err := strconv.Atoi(os.Args[1])
		exitOnError(err)
		portOutNum, err := strconv.Atoi(os.Args[2])
		exitOnError(err)
		pingPong(portmidi.DeviceID(portInNum),
			portmidi.DeviceID(portOutNum))
		return
	}
}

func pingPong(inID portmidi.DeviceID, outID portmidi.DeviceID) {
	fmt.Printf("In: %v\n", inID)
	fmt.Printf("Out: %v\n", outID)

	in, err := portmidi.NewInputStream(inID, 1024)
	exitOnError(err)

	out, err := portmidi.NewOutputStream(outID, 1024, 0)
	exitOnError(err)
	defer out.Close()

	pingSysEx := []byte{0xF0, 0x00, 0x22, 0x77, 0x01, 0xF7}

	err = out.WriteSysExBytes(portmidi.Time(), pingSysEx)
	exitOnError(err)

	msg, err := in.Read(1024)
	exitOnError(err)

	for i, b := range msg {
		fmt.Printf("SysEx message byte %d = %02x\n", i, b)
	}
}

func exitOnError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
