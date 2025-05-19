#!/usr/bin/env python3

import logging
import argparse
import time
from typing import List, Tuple
import mido
import matplotlib.pyplot as plt

logging.basicConfig(level=logging.ERROR)
logger = logging.getLogger(__name__)


def list_devices():
    """List all available MIDI devices."""
    print("Input Devices:")
    for i, name in enumerate(mido.get_input_names()):
        print(f"ID: {i}\tName: {name}")

    print("\nOutput Devices:")
    for i, name in enumerate(mido.get_output_names()):
        print(f"ID: {i}\tName: {name}")


def ping_pong(
    in_port_name: str, out_port_name: str, count: int
) -> Tuple[List[float], List[float]]:
    """
    Perform MIDI ping-pong test.

    Args:
        in_port_name: Name of input MIDI port
        out_port_name: Name of output MIDI port
        count: Number of ping-pongs to perform (0 for unlimited)
    Returns:
        Tuple of (timestamps, durations) in seconds
    """
    times = []
    durations = []
    ping_count = 0
    global_start_time = time.time()

    # SysEx messages for ping and pong (all bytes must be 0-127)
    ping_sysex = (0x00, 0x22, 0x77, 0x01)
    pong_sysex = (0x00, 0x22, 0x77, 0x02)

    logger.info(f"Starting ping-pong test with count={count}")
    logger.info(f"Input port: {in_port_name}")
    logger.info(f"Output port: {out_port_name}")

    with mido.open_input(in_port_name) as inport, mido.open_output(
        out_port_name
    ) as outport:
        while count == 0 or ping_count < count:
            start_time = time.time()

            logger.info(f"Sending ping {ping_count + 1}/{count if count > 0 else '∞'}")
            ping_msg = mido.Message("sysex", data=ping_sysex)
            outport.send(ping_msg)
            timestamp = time.time() - global_start_time
            print(f"{timestamp:.6f}: Ping? ", end="", flush=True)

            timeout = 2.0
            while time.time() - start_time < timeout:
                try:
                    msg = inport.poll()
                    if msg:
                        logger.debug(f"received: {msg.hex()}")
                        if msg.type == "sysex":
                            if msg.data == pong_sysex:
                                rtt = time.time() - start_time
                                print(f"Pong! ({rtt * 1000:.3f}ms)")

                                times.append(timestamp)
                                durations.append(rtt * 1000)
                                ping_count += 1
                                break
                            else:
                                logger.error(
                                    f"mismatch: expected {pong_sysex}, got {msg.data}"
                                )
                except Exception as e:
                    logger.error(f"error: {e}")
                    break

                time.sleep(0.01)

            if time.time() - start_time >= timeout:
                logger.error("timeout waiting for pong")

            time.sleep(0.02)

    logger.info("Test completed.")
    return times, durations


def save_graph(times: List[float], durations: List[float], title: str, filename: str):
    """Save a graph of ping-pong times."""
    plt.figure(figsize=(10, 6))
    plt.plot(times, durations, "b.-")
    plt.title(title or "MIDI Ping-Pong Latency")
    plt.xlabel("Time (seconds)")
    plt.ylabel("Round-trip time (ms)")
    plt.grid(True)
    plt.savefig(filename)
    plt.close()


def main():
    parser = argparse.ArgumentParser(description="MIDI Ping-Pong Test")
    parser.add_argument("--in", dest="in_port", type=str, help="Input MIDI port name")
    parser.add_argument(
        "--out", dest="out_port", type=str, help="Output MIDI port name"
    )
    parser.add_argument("--graph-title", type=str, help="Graph title")
    parser.add_argument(
        "--graph-filename", default="results/output.png", help="Graph filename"
    )
    parser.add_argument("--list", action="store_true", help="List MIDI devices")
    parser.add_argument(
        "--count",
        type=int,
        default=0,
        help="Number of ping-pongs to perform (0 for unlimited)",
    )
    parser.add_argument("--network", action="store_true", help="Use network MIDI")
    parser.add_argument("--port", type=int, default=1292, help="Port for network MIDI")

    args = parser.parse_args()

    if args.list:
        list_devices()
        return

    if args.network:
        # Use network MIDI
        in_port = f"tcp://:{args.port}"
        out_port = f"tcp://localhost:{args.port}"
    else:
        # Use physical MIDI devices
        if not args.in_port or not args.out_port:
            parser.error("--in and --out are required when not using network MIDI")
        in_port = args.in_port
        out_port = args.out_port

    times, durations = ping_pong(in_port, out_port, args.count)
    save_graph(times, durations, args.graph_title, args.graph_filename)


if __name__ == "__main__":
    main()
