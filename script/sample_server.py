from http.server import BaseHTTPRequestHandler, HTTPServer

class SimpleHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        # Send response status code
        self.send_response(200)
        # Send headers
        self.send_header('Content-type', 'text/plain')
        self.end_headers()
        # Send response body
        self.wfile.write(b"Hello! You connected successfully.\n")

if __name__ == "__main__":
    server_address = ("0.0.0.0", 3306)
    httpd = HTTPServer(server_address, SimpleHandler)
    print("Serving on port 3306...")
    httpd.serve_forever()
