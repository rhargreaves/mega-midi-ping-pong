import logging
import argparse
import time
from typing import List, Tuple
import mido
import matplotlib.pyplot as plt
import signal


logging.basicConfig(level=logging.ERROR)
logger = logging.getLogger(__name__)
user_interrupted = False


def signal_handler(sig, frame):
    print("User interrupted test")
    global user_interrupted
    user_interrupted = True


def list_devices():
    """List all available MIDI devices."""
    print("Input Devices:")
    for name in mido.get_input_names():
        print(name)

    print("\nOutput Devices:")
    for name in mido.get_output_names():
        print(name)


def ping_pong(
    in_port_name: str, out_port_name: str, max_count: int
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
    global_start_time = time.perf_counter()
    max_time = 30

    # SysEx messages for ping and pong (all bytes must be 0-127)
    ping_sysex = (0x00, 0x22, 0x77, 0x01)
    pong_sysex = (0x00, 0x22, 0x77, 0x02)

    logger.info(f"Starting test with max_count: {max_count}")
    logger.info(f"Input port: {in_port_name}")
    logger.info(f"Output port: {out_port_name}")

    with mido.open_input(in_port_name) as inport, mido.open_output(
        out_port_name
    ) as outport:
        while (
            (max_count == 0 or ping_count < max_count)
            and time.perf_counter() - global_start_time < max_time
            and not user_interrupted
        ):
            logger.info(f"sending ping {ping_count + 1}/{max_count}")
            ping_msg = mido.Message("sysex", data=ping_sysex)
            start_time = time.perf_counter()
            outport.send(ping_msg)
            timestamp = start_time - global_start_time
            print(f"{timestamp:.6f}: Ping? ", end="", flush=True)

            timeout = 2.0
            while time.perf_counter() - start_time < timeout:
                try:
                    msg = inport.poll()
                    if msg:
                        logger.debug(f"received: {msg.hex()}")
                        if msg.type == "sysex":
                            if msg.data == pong_sysex:
                                rtt = time.perf_counter() - start_time
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

            if time.perf_counter() - start_time >= timeout:
                logger.error("timeout waiting for pong")

    logger.info("test completed.")
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
    signal.signal(signal.SIGINT, signal_handler)

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
    args = parser.parse_args()

    if args.list:
        list_devices()
        return

    if not args.in_port or not args.out_port:
        parser.error("--in and --out are required")

    times, durations = ping_pong(args.in_port, args.out_port, args.count)
    save_graph(times, durations, args.graph_title, args.graph_filename)
