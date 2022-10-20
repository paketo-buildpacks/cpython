from http.server import BaseHTTPRequestHandler, HTTPServer
import sys

class Server(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header("Content-type", "text/plain")
        self.end_headers()
        self.wfile.write(bytes("Hello world!", "utf-8"))

if __name__ == "__main__":
    webServer = HTTPServer(('', int(sys.argv[1])), Server)
    webServer.serve_forever()
