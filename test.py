#!/usr/bin/env python3

import time
import logging
import random
import sys
from http.server import HTTPServer, BaseHTTPRequestHandler
import threading
import signal
import os
import json
import urllib.request
import urllib.error

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

def fetch_api_data(endpoint):
    """Fetch data from the gemd API"""
    try:
        url = f"http://127.0.0.1:9876/api/v1/{endpoint}"
        req = urllib.request.Request(url)
        with urllib.request.urlopen(req, timeout=5) as response:
            data = json.loads(response.read().decode())
            return data
    except (urllib.error.URLError, json.JSONDecodeError, KeyError) as e:
        logger.warning(f"Failed to fetch {endpoint} from API: {e}")
        return None

class TestHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type', 'text/plain')
        self.end_headers()
        
        self.wfile.write(b'Hello from Gemstone test app!\n')
        self.wfile.write(f'PID: {os.getpid()}\n'.encode())
        self.wfile.write(f'Uptime: {time.time() - start_time:.1f}s\n'.encode())
        
        # Fetch and display API information
        self.wfile.write(b'\n--- Gemstone Daemon Info ---\n')
        system_info = fetch_api_data('system')
        if system_info and system_info.get('success'):
            data = system_info.get('data', {})
            self.wfile.write(f'Version: {data.get("version", "unknown")}\n'.encode())
            self.wfile.write(f'Process Count: {data.get("process_count", 0)}\n'.encode())
            if 'system_stats' in data:
                stats = data['system_stats']
                cpu_percent = stats.get("cpu_percent", 0)
                self.wfile.write(f'CPU Usage (system): {cpu_percent:.1f}%\n'.encode())
                
                # Memory in GB/MB
                mem_total = stats.get("memory_total", 0)  # bytes
                mem_used = stats.get("memory_used", 0)    # bytes
                
                def format_bytes(bytes_val):
                    if bytes_val >= 1024**3:  # GB
                        return f"{bytes_val / 1024**3:.1f} GB"
                    elif bytes_val >= 1024**2:  # MB
                        return f"{bytes_val / 1024**2:.1f} MB"
                    else:  # KB
                        return f"{bytes_val / 1024:.1f} KB"
                
                self.wfile.write(f'Memory Usage: {format_bytes(mem_used)} / {format_bytes(mem_total)}\n'.encode())
                
                # Disk usage
                disk_total = stats.get("disk_total", 0)
                disk_used = stats.get("disk_used", 0)
                disk_percent = stats.get("disk_percent", 0)
                self.wfile.write(f'Disk Usage: {format_bytes(disk_used)} / {format_bytes(disk_total)} ({disk_percent:.1f}%)\n'.encode())
                
                # Load average
                load_avg = stats.get("load_average", [])
                if load_avg and len(load_avg) >= 3:
                    self.wfile.write(f'Load Average: {load_avg[0]:.2f}, {load_avg[1]:.2f}, {load_avg[2]:.2f}\n'.encode())
                
                # System uptime
                sys_uptime = stats.get("uptime", 0)
                def format_uptime(seconds):
                    days = seconds // 86400
                    hours = (seconds % 86400) // 3600
                    minutes = (seconds % 3600) // 60
                    if days > 0:
                        return f"{days}d {hours}h {minutes}m"
                    elif hours > 0:
                        return f"{hours}h {minutes}m"
                    else:
                        return f"{minutes}m"
                
                self.wfile.write(f'System Uptime: {format_uptime(sys_uptime)}\n'.encode())
                
                # Timestamp
                timestamp = stats.get("timestamp", "")
                if timestamp:
                    # Parse and format timestamp
                    try:
                        from datetime import datetime
                        # Remove microseconds and timezone for simpler parsing
                        simple_ts = timestamp.split('.')[0]  # Remove microseconds
                        if '+' in simple_ts:
                            simple_ts = simple_ts.split('+')[0]
                        elif '-' in simple_ts and simple_ts.count('-') > 2:
                            # Handle timezone offset
                            parts = simple_ts.rsplit('-', 1)
                            simple_ts = parts[0]
                        
                        dt = datetime.fromisoformat(simple_ts)
                        formatted_time = dt.strftime('%Y-%m-%d %H:%M:%S')
                        self.wfile.write(f'Last Updated: {formatted_time}\n'.encode())
                    except Exception as e:
                        logger.warning(f"Failed to parse timestamp {timestamp}: {e}")
                        # Fallback: just show the date/time part
                        if 'T' in timestamp:
                            date_part = timestamp.split('T')[0]
                            time_part = timestamp.split('T')[1].split('.')[0]
                            self.wfile.write(f'Last Updated: {date_part} {time_part}\n'.encode())
                        else:
                            self.wfile.write(f'Last Updated: {timestamp}\n'.encode())
        else:
            self.wfile.write(b'Unable to fetch system info from API\n')
        
        self.wfile.write(b'\n--- Managed Processes ---\n')
        processes = fetch_api_data('processes')
        if processes and processes.get('success'):
            proc_list = processes.get('data', [])
            if proc_list:
                for proc in proc_list:
                    self.wfile.write(f'ID: {proc.get("id", "unknown")}, Name: {proc.get("name", "unknown")}, Status: {proc.get("status", "unknown")}\n'.encode())
            else:
                self.wfile.write(b'No processes currently managed\n')
        else:
            self.wfile.write(b'Unable to fetch processes from API\n')

    def log_message(self, format, *args):
        # Suppress default HTTP server logs
        pass

def run_server():
    try:
        server = HTTPServer(('localhost', 8080), TestHandler)
        logger.info("HTTP server started on http://localhost:8080")
        server.serve_forever()
    except Exception as e:
        logger.error(f"HTTP server error: {e}")

def simulate_work():
    data = []
    while True:
        for _ in range(10000):
            random.random() ** 2

        if random.random() < 0.1:
            data.append([random.random() for _ in range(1000)])
            if len(data) > 10:
                data.pop(0)

        time.sleep(1)

def main():
    logger.info("Gemstone test app started")
    logger.info(f"PID: {os.getpid()}")

    server_thread = threading.Thread(target=run_server, daemon=True)
    server_thread.start()

    work_thread = threading.Thread(target=simulate_work, daemon=True)
    work_thread.start()

    counter = 0
    while True:
        counter += 1

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