package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/bradhe/stopwatch"
	"github.com/rakyll/portmidi"
	"github.com/wcharczuk/go-chart"
)

func main() {
	portmidi.Initialize()

	if len(os.Args) == 2 && os.Args[1] == "list" {
		listDevices()
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

func listDevices() {
	for id := 0; id < portmidi.CountDevices(); id++ {
		inf := portmidi.Info(portmidi.DeviceID(id))
		fmt.Printf("ID: %d\tName: %s\tInput: %t\tOutput: %t\n",
			id,
			inf.Name,
			inf.IsInputAvailable,
			inf.IsOutputAvailable)
	}
}

func pingPong(inID portmidi.DeviceID, outID portmidi.DeviceID) {
	fmt.Printf("In: %v\n", portmidi.Info(inID).Name)
	fmt.Printf("Out: %v\n", portmidi.Info(outID).Name)

	in, err := portmidi.NewInputStream(inID, 1024)
	exitOnError(err)
	defer in.Close()

	out, err := portmidi.NewOutputStream(outID, 1024, 0)
	exitOnError(err)
	defer out.Close()

	pingSysEx := []byte{0xF0, 0x00, 0x22, 0x77, 0x01, 0xF7}

	var times []float64
	var durations []float64

	startTime := time.Now()

	for time.Now().Sub(startTime) < time.Second*30 {

		watch := stopwatch.Start()
		err = out.WriteSysExBytes(portmidi.Time(), pingSysEx)
		exitOnError(err)
		duration := time.Now().Sub(startTime)
		fmt.Printf("%v: Ping? ", duration)
		waitForEvent(in)
		event, err := in.ReadSysExBytes(6)
		exitOnError(err)

		pongSysEx := []byte{0xF0, 0x00, 0x22, 0x77, 0x02, 0xF7, 0x00, 0x00}
		res := bytes.Compare(event, pongSysEx)
		if res == 0 {
			watch.Stop()
			fmt.Printf("Pong! (%v)\n", watch.String())

			times = append(times, float64(duration.Seconds()))
			durations = append(durations, float64(watch.Milliseconds()))

		} else {
			fmt.Printf("Mismatch! %02x\n", event)
		}

		time.Sleep(time.Millisecond * 20)
	}

	graph := chart.Chart{
		XAxis: chart.XAxis{
			ValueFormatter: func(v interface{}) string {
				if vf, isFloat := v.(float64); isFloat {
					return fmt.Sprintf("%0.2f", vf)
				}
				return ""
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: times,
				YValues: durations,
			},
		},
	}

	f, _ := os.Create("output.png")
	defer f.Close()
	err = graph.Render(chart.PNG, f)
	exitOnError(err)
}

func waitForEvent(stream *portmidi.Stream) {
	for {
		ready, err := stream.Poll()
		exitOnError(err)
		if ready {
			break
		}
		time.Sleep(time.Microsecond)
	}
}

func exitOnError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
