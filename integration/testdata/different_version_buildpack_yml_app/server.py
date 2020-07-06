from http.server import BaseHTTPRequestHandler, HTTPServer
import os

class Server(BaseHTTPRequestHandler):
  def do_GET(self):
        self.send_response(200)
        self.send_header("Content-type","text/plain")
        self.end_headers()
        self.wfile.write(bytes("hello world", "utf8"))
        return
  def do_HEAD(self):
        self.send_response(200)
        self.send_header("Content-type","text/plain")
        self.end_headers()
        return

port = int(os.getenv("PORT", "8080"))
server_address = ("", port)
httpd = HTTPServer(server_address, Server)
print("server is listening on port " + str(port))
httpd.serve_forever()
