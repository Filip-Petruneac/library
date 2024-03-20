from http.server import BaseHTTPRequestHandler, HTTPServer
import mysql.connector

conn = mysql.connector.connect(
    # host="192.168.1.44",
    host="docker_db_1",
    user="root",
    port = "3307",
    password="password",
    database="student")

cursor = conn.cursor()

print("Connected to MySql server!")

hostName= "localhost"
PORT = 8080


class WelcomeHandler(BaseHTTPRequestHandler):

    def do_GET(self):
        cursor.execute("SELECT * FROM students")
        result = cursor.fetchall()
        print(result)
        # Send an HTTP response with header and body
        self.send_response(200)  # OK status code
        self.send_header("Content-type", "text/html")
        self.end_headers()
        self.wfile.write(bytes("<html><head><title>Docker Compose</title></head>", encoding='utf-8'))
        self.wfile.write(bytes("<p>Users: %s</p>" % result, encoding='utf-8'))
        self.wfile.write(bytes("</body></html>", encoding='utf-8'))
        self.wfile.close()

with HTTPServer(("", PORT), WelcomeHandler) as httpd:
    print(f"Server listening on port {PORT}")
    httpd.serve_forever()