#!/usr/bin/env python3

import unittest
import threading
import mido
from ping_pong import ping_pong
import logging

logging.basicConfig(level=logging.ERROR)


class TestPingPong(unittest.TestCase):
    def setUp(self):
        self.md_in = mido.open_input("md_in", virtual=True)
        self.md_out = mido.open_output("md_out", virtual=True)
        self.received_count = 0

    def tearDown(self):
        self.md_in.close()
        self.md_out.close()

    def test_ping_pong(self):
        # Start a responder thread
        stop_event = threading.Event()
        responder_thread = threading.Thread(target=self._responder, args=(stop_event,))
        responder_thread.start()

        try:
            # Run 5 ping-pongs
            max_count = 5
            times, durations = ping_pong("md_out", "md_in", count=max_count)

            # Check results
            self.assertEqual(len(times), max_count, f"Expected {max_count} ping-pongs")
            self.assertEqual(
                len(durations), max_count, f"Expected {max_count} durations"
            )
            self.assertEqual(
                self.received_count,
                max_count,
                f"Expected {max_count} received messages",
            )

            for duration in durations:
                self.assertGreater(duration, 0, "Duration should be positive")
                self.assertLess(duration, 1000, "Duration should be less than 1 second")

        finally:
            stop_event.set()
            responder_thread.join()

    def _responder(self, stop_event):
        """Respond to ping messages with pong."""

        ping_sysex = (0x00, 0x22, 0x77, 0x01)
        pong_sysex = (0x00, 0x22, 0x77, 0x02)

        logger = logging.getLogger("virt-mdmi")
        logger.info("starting")
        while not stop_event.is_set():
            try:
                msg = self.md_in.poll()
                if msg:
                    logger.info(f"received: {msg.hex()}")
                    if msg.type == "sysex" and msg.data == ping_sysex:
                        pong_msg = mido.Message("sysex", data=pong_sysex)
                        logger.info(f"sending: {pong_msg.hex()}")
                        self.md_out.send(pong_msg)
                        self.received_count += 1
            except Exception as e:
                logger.error(f"error: {e}")
                continue


if __name__ == "__main__":
    unittest.main()
