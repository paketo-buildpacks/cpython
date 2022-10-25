from http.server import BaseHTTPRequestHandler, HTTPServer
import os
import sys

class Server(BaseHTTPRequestHandler):
  def do_GET(self):
    self.send_response(200)
    self.send_header("Content-type","text/plain")
    self.end_headers()

    self.wfile.write(bytes("hello world\n", "utf8"))

    prefix = os.getenv("PYTHONPYCACHEPREFIX")
    self.wfile.write(bytes(f'PYTHONPYCACHEPREFIX={prefix}\n', "utf8"))

    version = sys.version
    self.wfile.write(bytes(f'Python version: {version}\n', "utf8"))

  def do_HEAD(self):
    self.send_response(200)
    self.send_header("Content-type","text/plain")
    self.end_headers()

port = int(os.getenv("PORT", "8080"))
server_address = ("", port)
httpd = HTTPServer(server_address, Server)
print("server is listening on port " + str(port))
httpd.serve_forever()
