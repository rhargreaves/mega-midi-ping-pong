import unittest
import mido
from ping_pong.ping_pong import ping_pong
import logging
from virt_mdmi import VirtMdmi

logging.basicConfig(level=logging.ERROR)

MD_IN_PORT = "md_in"
MD_OUT_PORT = "md_out"


class TestPingPong(unittest.TestCase):
    def setUp(self):
        self.md_in = mido.open_input(MD_IN_PORT, virtual=True)
        self.md_out = mido.open_output(MD_OUT_PORT, virtual=True)

    def tearDown(self):
        self.md_in.close()
        self.md_out.close()

    def test_ping_pong(self):
        virt_mdmi = VirtMdmi(self.md_in, self.md_out)
        virt_mdmi.start()

        try:
            max_count = 3
            times, durations = ping_pong(MD_OUT_PORT, MD_IN_PORT, max_count=max_count)

            self.assertEqual(len(times), max_count, f"Expected {max_count} ping-pongs")
            self.assertEqual(
                len(durations), max_count, f"Expected {max_count} durations"
            )
            self.assertEqual(
                virt_mdmi.received_count,
                max_count,
                f"Expected {max_count} received messages",
            )

            for duration in durations:
                self.assertGreater(duration, 0, "Duration should be positive")
                self.assertLess(duration, 1000, "Duration should be less than 1 second")

        finally:
            virt_mdmi.stop()
