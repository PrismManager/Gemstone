#!/usr/bin/env python3
"""
Test application for Gemstone process manager.
This app demonstrates various features by running a simple web server,
logging messages, and consuming some CPU/memory.
"""

import time
import logging
import random
import sys
from http.server import HTTPServer, BaseHTTPRequestHandler
import threading
import signal
import os

# Setup logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class TestHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type', 'text/plain')
        self.end_headers()
        self.wfile.write(b'Hello from Gemstone test app!\n')
        self.wfile.write(f'PID: {os.getpid()}\n'.encode())
        self.wfile.write(f'Uptime: {time.time() - start_time:.1f}s\n'.encode())

    def log_message(self, format, *args):
        # Suppress default HTTP server logs
        pass

def run_server():
    """Run a simple HTTP server on port 8080"""
    try:
        server = HTTPServer(('localhost', 8080), TestHandler)
        logger.info("HTTP server started on http://localhost:8080")
        server.serve_forever()
    except Exception as e:
        logger.error(f"HTTP server error: {e}")

def simulate_work():
    """Simulate CPU and memory usage"""
    data = []
    while True:
        # Simulate CPU work
        for _ in range(10000):
            random.random() ** 2

        # Simulate memory usage (grow list occasionally)
        if random.random() < 0.1:
            data.append([random.random() for _ in range(1000)])
            if len(data) > 10:
                data.pop(0)

        time.sleep(1)

def main():
    logger.info("Gemstone test app started")
    logger.info(f"PID: {os.getpid()}")
    logger.info("Features being tested:")
    logger.info("- Process management (start/stop/restart)")
    logger.info("- Auto-restart capability")
    logger.info("- Logging (check gemstone logs)")
    logger.info("- Resource monitoring (CPU/memory)")
    logger.info("- HTTP server for external access")

    # Start HTTP server in background
    server_thread = threading.Thread(target=run_server, daemon=True)
    server_thread.start()

    # Start work simulation in background
    work_thread = threading.Thread(target=simulate_work, daemon=True)
    work_thread.start()

    # Main loop - log periodically
    counter = 0
    while True:
        counter += 1
        logger.info(f"Test app running - iteration {counter}")
        logger.info(f"Memory usage: {len(data) if 'data' in globals() else 0}KB simulated")

        # Simulate occasional errors (for testing auto-restart)
        if random.random() < 0.05:  # 5% chance
            logger.warning("Simulating an error (this should trigger auto-restart if enabled)")
            sys.exit(1)

        time.sleep(10)

if __name__ == '__main__':
    start_time = time.time()
    data = []  # For memory simulation

    def signal_handler(signum, frame):
        logger.info("Received signal, shutting down gracefully")
        sys.exit(0)

    signal.signal(signal.SIGTERM, signal_handler)
    signal.signal(signal.SIGINT, signal_handler)

    try:
        main()
    except KeyboardInterrupt:
        logger.info("Interrupted by user")
    except Exception as e:
        logger.error(f"Unexpected error: {e}")
        sys.exit(1)