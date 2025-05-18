package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/wcharczuk/go-chart"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/rtmididrv"
)

func main() {
	drv, err := rtmididrv.New()
	exitOnError(err)
	defer drv.Close()

	inPtr := flag.Uint("in", 0, "In Device ID")
	outPtr := flag.Uint("out", 0, "Out Device ID")
	graphTitlePtr := flag.String("graph-title", "", "Graph Title")
	graphFileNamePtr := flag.String("graph-filename", "results/output.png", "Graph Filename")
	listPtr := flag.Bool("list", false, "List Devices")

	flag.Parse()

	if *listPtr {
		listDevices(drv)
		return
	}

	times, durations := pingPong(drv, int(*inPtr), int(*outPtr))

	saveGraph(times, durations, *graphTitlePtr, *graphFileNamePtr)
}

func listDevices(drv *rtmididrv.Driver) {
	ins, err := drv.Ins()
	exitOnError(err)
	outs, err := drv.Outs()
	exitOnError(err)

	fmt.Println("Input Devices:")
	for i, in := range ins {
		fmt.Printf("ID: %d\tName: %s\n", i, in.String())
	}

	fmt.Println("\nOutput Devices:")
	for i, out := range outs {
		fmt.Printf("ID: %d\tName: %s\n", i, out.String())
	}
}

func pingPong(drv *rtmididrv.Driver, inID, outID int) (times []float64, durations []float64) {
	ins, err := drv.Ins()
	exitOnError(err)
	outs, err := drv.Outs()
	exitOnError(err)

	if inID >= len(ins) || outID >= len(outs) {
		panic("Invalid device ID")
	}

	in := ins[inID]
	out := outs[outID]

	fmt.Printf("In: %v\n", in.String())
	fmt.Printf("Out: %v\n", out.String())

	err = in.Open()
	exitOnError(err)
	defer in.Close()

	err = out.Open()
	exitOnError(err)
	defer out.Close()

	sysexChan := make(chan []byte)
	stopChan := make(chan struct{})

	pongSysEx := []byte{0x00, 0x22, 0x77, 0x02}

	r := reader.New(
		reader.NoLogger(),
		reader.SysEx(func(_ *reader.Position, b []byte) {
			msg := make([]byte, len(b))
			copy(msg, b)
			sysexChan <- msg
		}),
	)
	go func() {
		err := r.ListenTo(in)
		if err != nil {
			close(sysexChan)
		}
		close(stopChan)
	}()

	globalStartTime := time.Now()

	for time.Now().Sub(globalStartTime) < time.Second*30 {
		startTime := time.Now()

		pingSysEx := []byte{0xF0, 0x00, 0x22, 0x77, 0x01, 0xF7}
		if _, err := out.Write(pingSysEx); err != nil {
			exitOnError(err)
		}
		timestamp := time.Now().Sub(globalStartTime)
		fmt.Printf("%v: Ping? ", timestamp)

		var event []byte
		select {
		case event = <-sysexChan:
			// got a SysEx event
		case <-time.After(2 * time.Second):
			fmt.Println("Timeout waiting for pong SysEx")
			continue
		}

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
	close(stopChan)
	return times, durations
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

func exitOnError(err error) {
	if err != nil {
		panic(err.Error())
	}
}
