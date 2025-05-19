import mido
import threading
import logging


class VirtMdmi:
    def __init__(self, md_in, md_out):
        self.md_in = md_in
        self.md_out = md_out

    def stop(self):
        if self.stop_event:
            self.stop_event.set()
        if self.responder_thread:
            self.responder_thread.join()

    def start(self):
        self.received_count = 0
        self.stop_event = threading.Event()
        self.responder_thread = threading.Thread(
            target=self._responder, args=(self.stop_event,)
        )
        self.responder_thread.start()

    def _responder(self, stop_event):
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
