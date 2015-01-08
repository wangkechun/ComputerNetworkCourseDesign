# coding=utf8


import os
import posixpath
import urllib
import cgi
import sys
import shutil
import select
import socket
import time
import mimetools
import threading
import BaseHTTPServer

try:
    from cStringIO import StringIO
except ImportError:
    from StringIO import StringIO

www_directory = "./"

HTTP_STATUS_MSG = BaseHTTPServer.BaseHTTPRequestHandler.responses


class BaseHTTPRequestHandler:
    rbufsize = -1
    wbufsize = 0
    timeout = 5.

    def __init__(self, request, client_address, server):
        self.request = request
        self.client_address = client_address
        self.server = server
        self.raw_requestline = None
        self.rfile = None
        self.wfile = None
        self.setup()
        try:
            self.handle()
        finally:
            self.finish()

    def setup(self):
        self.connection = self.request
        if self.timeout is not None:
            self.connection.settimeout(self.timeout)
        self.rfile = self.connection.makefile('rb', self.rbufsize)
        self.wfile = self.connection.makefile('wb', self.rbufsize)

    def finish(self):
        if not self.wfile.closed:
            try:
                self.wfile.flush()
            except socket.error:
                pass

            self.wfile.close()
            self.rfile.close()

    def handle(self):
        try:
            self.raw_requestline = self.rfile.readline(65537)
            if len(self.raw_requestline) > 65536:
                return
            if not self.raw_requestline:
                return
            if not self.parse_request():
                return

            if self.command == 'GET':
                self.do_GET()
            elif self.command == 'POST':
                self.do_POST()
            self.wfile.flush()
        except socket.timeout as e:
            return

    def parse_request(self):
        requestsline = self.raw_requestline.rstrip('\r\n')
        self.requestsline = requestsline
        words = requestsline.split()
        if len(words) == 3:
            command, path, version = words
            if version != 'HTTP/1.1':
                self.send_error(400, 'Bad request version')
                return False
        else:
            self.send_error(400, 'Bad request')
            return False
        self.command, self.path, self.request_version = command, path, version
        self.headers = mimetools.Message(self.rfile, 0)
        return True

    def send_error(self, code, message=None):
        self.end_headers()
        if message:
            content = message
        else:
            content = HTTP_STATUS_MSG.get(code, '')
        self.wfile.write(content)

    def send_response(self, code, message=None):
        if message is None:
            if code in HTTP_STATUS_MSG:
                message = HTTP_STATUS_MSG[code][0]
            else:
                message = ''
        self.wfile.write("HTTP/1.1 %d %s\r\n" % (code, message))

    def send_header(self, keyword, value):
        self.wfile.write("%s: %s\r\n" % (keyword, value))

    def end_headers(self):
        self.wfile.write("\r\n")

    weekdayname = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

    monthname = [None,
                 'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
                 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

    def date_time_string(self, timestamp=None):
        if timestamp is None:
            timestamp = time.time()
        year, month, day, hh, mm, ss, wd, y, z = time.gmtime(timestamp)
        s = "%s, %02d %3s %4d %02d:%02d:%02d GMT" % (
            self.weekdayname[wd],
            day, self.monthname[month], year,
            hh, mm, ss)
        return s


class SimpleHTTPRequestsHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        f = self.send_head()
        if f:
            try:
                shutil.copyfileobj(f,self.wfile)
            finally:
                f.close()

    def send_head(self):
        path = self.translate_path(self.path)
        f = None
        if os.path.isdir(path):
            if not self.path.endswith('/'):
                self.send_response(301)
                self.send_header("Location", self.path + "/")
                self.end_headers()
                return None

            for index in "index.html", "index.htm":
                index = os.path.join(path, index)
                if os.path.exists(index):
                    path = index
                    break

            else:
                return self.list_directory(path)
        try:
            f = open(path, "rb")

        except IOError:
            self.send_error(404, "File not found")
            return None

        try:
            self.send_response(200)
            fs = os.fstat(f.fileno())
            self.send_header("Content-Length", str(fs[6]))
            self.send_header("Last-Modified", self.date_time_string(fs.st_mtime))
            self.end_headers()
            return f
        except:
            f.close()
            raise

    def list_directory(self, path):
        try:
            list = os.listdir(path)
        except os.error:
            self.send_error(404, "No permission to list directory")
            return None

        list.sort(key=lambda a: a.lower())
        f = StringIO()

        for name in list:
            fullname = os.path.join(path, name)
            displayname = linkname = name
            if os.path.isdir(fullname):
                displayname = name + "/"
                linkname = name + "/"
            if os.path.islink(fullname):
                displayname = name + "@"
            f.write('<li><a href="%s">%s</a></li>\n' %
                    (urllib.quote(linkname), cgi.escape(displayname)))
        length = f.tell()
        f.seek(0)
        self.send_response(200)
        self.send_header("Content-type", "text/html;charset=%s" % "UTF-8")
        self.send_header("Content-Length", str(length))
        self.end_headers()
        return f

    def translate_path(self, path):
        path = path.split('?', 1)[0]
        path = path.split('#', 1)[0]
        trailing_slash = path.rstrip().endswith('/')
        path = posixpath.normpath(urllib.unquote(path))
        words = path.split('/')
        words = filter(None, words)
        path = www_directory
        for word in words:
            drive, word = os.path.splitdrive(word)
            head, word = os.path.split(word)
            if word in (os.curdir, os.pardir):
                continue
            path = os.path.join(path, word)
        if trailing_slash:
            path += '/'
        return path



def _eintr_retry(func, *args):
    while True:
        try:
            return func(*args)
        except (OSError, select.error) as e:
            if e.args[0] != 4 :#errno.EINTR
                raise


class HTTPServer:
    request_queue_size = 5
    allow_reuse_address = True

    def __init__(self, server_address, RequestHandlerClass, bind_and_activate=True):

        self.server_address = server_address
        self.RequestHandlerClass = RequestHandlerClass
        self.__is_shut_down = threading.Event()
        self.__shutdown_request = False

        self.socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        if bind_and_activate:
            self.server_bind()
            self.socket.listen(self.request_queue_size)

    def server_bind(self):
        if self.allow_reuse_address:
            self.socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self.socket.bind(self.server_address)
        self.server_address = self.socket.getsockname()  #('0.0.0.0', 8000)

        host, port = self.socket.getsockname()[:2]
        self.server_name = socket.getfqdn(host)
        self.server_port = port


    def serve_forever(self, poll_interval=0.5):
        self.__is_shut_down.clear()
        try:
            while not self.__shutdown_request:
                r, w, e = _eintr_retry(select.select, [self], [], [], poll_interval)
                if self in r:
                    self._handle_request_noblock()
        finally:
            self.__shutdown_request = False
            self.__is_shut_down.set()

    def fileno(self):
        return self.socket.fileno()

    def get_request(self):
        return self.socket.accept()

    def handle_request(self):
        timeout = self.socket.gettimeout()
        if timeout is None:
            timeout = self.timeout
        fd_sets = _eintr_retry(select.select, [self], [], [], timeout)
        if not fd_sets[0]:
            return
        self._handle_request_noblock()

    def _handle_request_noblock(self):
        try:
            request, client_address = self.get_request()
        except socket.error:
            return
        try:
            self.RequestHandlerClass(request, client_address, self)
        except Exception as e:
            print(request, client_address, e)
            # import ipdb;ipdb.set_trace()
            raise


www_directory = "./"


def main():
    port = 8000
    server_address = ('', port)
    httpd = HTTPServer(server_address, SimpleHTTPRequestsHandler)
    sa = httpd.socket.getsockname()
    print("Serving HTTP on %s, port %s ..." % (sa[0], sa[1]))
    print("open http://127.0.0.1:8000")
    httpd.serve_forever()

if __name__ == '__main__':
    main()



