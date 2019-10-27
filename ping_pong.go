package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rakyll/portmidi"
	"github.com/wcharczuk/go-chart"
)

func main() {
	portmidi.Initialize()

	inPtr := flag.Uint("in", 0, "In Device ID")
	outPtr := flag.Uint("out", 0, "Out Device ID")
	graphTitlePtr := flag.String("graph-title", "", "Graph Title")
	graphFileNamePtr := flag.String("graph-filename", "results/output.png", "Graph Filename")
	listPtr := flag.Bool("list", false, "List Devices")

	flag.Parse()

	if *listPtr {
		listDevices()
		return
	}

	pingPong(portmidi.DeviceID(*inPtr),
		portmidi.DeviceID(*outPtr),
		*graphTitlePtr,
		*graphFileNamePtr)
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

func pingPong(inID portmidi.DeviceID,
	outID portmidi.DeviceID,
	graphTitle string,
	graphFileName string) {

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

	globalStartTime := time.Now()

	for time.Now().Sub(globalStartTime) < time.Second*30 {

		startTime := time.Now()
		err = out.WriteSysExBytes(portmidi.Time(), pingSysEx)
		exitOnError(err)
		timestamp := time.Now().Sub(globalStartTime)
		fmt.Printf("%v: Ping? ", timestamp)
		waitForEvent(in)
		event, err := in.ReadSysExBytes(6)
		exitOnError(err)

		pongSysEx := []byte{0xF0, 0x00, 0x22, 0x77, 0x02, 0xF7, 0x00, 0x00}
		res := bytes.Compare(event, pongSysEx)
		if res == 0 {
			rtt := time.Now().Sub(startTime)
			fmt.Printf("Pong! (%v)\n", rtt)

			times = append(times, float64(timestamp.Seconds()))
			durations = append(durations, float64(rtt.Nanoseconds())/1000000.0)

		} else {
			fmt.Printf("Mismatch! %02x\n", event)
		}

		time.Sleep(time.Millisecond * 20)
	}

	saveGraph(times, durations, graphTitle, graphFileName)
}

func saveGraph(
	times []float64,
	durations []float64,
	graphTitle string,
	graphFileName string) {

	graph := chart.Chart{
		Title: graphTitle,
		XAxis: chart.XAxis{
			ValueFormatter: func(v interface{}) string {
				if vf, isFloat := v.(float64); isFloat {
					return fmt.Sprintf("%0.2f", vf)
				}
				return ""
			},
			Name: "Time (seconds)",
		},
		YAxis: chart.YAxis{
			Name: "Round-trip time (ms)",
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: times,
				YValues: durations,
			},
		},
	}

	f, _ := os.Create(graphFileName)
	defer f.Close()
	err := graph.Render(chart.PNG, f)
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
