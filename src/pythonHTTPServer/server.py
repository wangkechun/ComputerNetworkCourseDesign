import os
import posixpath
import BaseHTTPServer
import urllib
import cgi
import sys
import shutil
import mimetypes
import socket
import SocketServer
import time
import mimetools
try:
  from cStringIO import StringIO
except ImportError :
  from StringIO import StringIO


www_directory = "/home/wkc/project/ComputerNetworkCourseDesign/src/pythonHTTPServer/v3.bootcss.com"
www_directory = "/home/wkc/project/ComputerNetworkCourseDesign/src/pythonHTTPServer/"


class StreamRequestHandler(SocketServer.BaseRequestHandler):
  rbufsize = -1
  wbufsize = 0
  timeout = None
  disable_nagle_algorithm = False

  def setup(self):
    self.connection = self.request
    if self.timeout is not None:
      self.connection.settimeout(self.timeout)
    self.rfile = self.connection.makefile('rb',self.rbufsize)
    self.wfile = self.connection.makefile('wb',self.rbufsize)

  def finish(self):
    if not self.wfile.closed:
      try:
        self.wfile.flush()
      except socket.error:
        pass

      self.wfile.close()
      self.rfile.close()


class BaseHTTPRequestHandler(StreamRequestHandler):
  MessageClass = mimetools.Message
  def parse_request(self):
    requestsline = self.raw_requestline
    requestsline = requestsline.rstrip('\r\n')
    self.requestsline = requestsline
    words = requestsline.split()
    if len(words) ==3:
      command,path,version = words
      if version!='HTTP/1.1':
        self.send_error(400,'Bad request version')
        return False
    else:
      self.send_error(400,'Bad request')
      return False
    self.command,self.path,self.request_version = command,path,version
    self.headers = self.MessageClass(self.rfile,0)
    return True

  def handle_one_requests(self):
    try:
      self.raw_requestline = self.rfile.readline(65537)
      if len(self.raw_requestline)>65536:
        self.requestsline =''
        self.command = ''
        self.send_error(414)
        return
      if not self.raw_requestline:
        return
      if not self.parse_request():
        return

      if self.command=='GET':
        self.do_GET()
      elif self.command=='POST':
        self.do_POST()
      self.wfile.flush()
    except socket.timeout as e:
      return

  def handle(self):
    self.handle_one_requests()

  def send_error(self,code,message=None):
    self.end_headers()
    content = message
    self.wfile.write(content)
 
  def send_response(self, code, message=None):
        """Send the response header and log the response code.

        Also send two standard headers with the server software
        version and the current date.

        """
        if message is None:
            if code in self.responses:
                message = self.responses[code][0]
            else:
                message = ''
        if self.request_version != 'HTTP/0.9':
            self.wfile.write("HTTP/1.1 %d %s\r\n" %
                             (code, message))
            # print (self.protocol_version, code, message)
        # self.send_header('Server', self.version_string())
        # self.send_header('Date', self.date_time_string())

  def send_header(self, keyword, value):
      """Send a MIME header."""
      if self.request_version != 'HTTP/0.9':
          self.wfile.write("%s: %s\r\n" % (keyword, value))

      if keyword.lower() == 'connection':
          if value.lower() == 'close':
              self.close_connection = 1
          elif value.lower() == 'keep-alive':
              self.close_connection = 0

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
        self.copyfile(f,self.wfile)
      finally:
        f.close()

  # def do_HEAD(self):
  #   f = self.send_head()
  #   if f:
  #     f.close()

  def send_head(self):
    path = self.translate_path(self.path)
    f = None
    if os.path.isdir(path):
      if not self.path.endswith('/'):
        self.send_response(301)
        self.send_head("Location",self.path+"/")
        self.end_headers()
        return None

      for index in "index.html","index.htm":
        index = os.path.join(path,index)
        if os.path.exists(index):
          path = index 
          break

      else:
        return self.list_directory(path)
    try:
      f = open(path,"rb")

    except IOError:
      self.send_error(404,"File not found")
      return None
    try:
      self.send_response(200)
      fs = os.fstat(f.fileno())
      self.send_header("Content-Length",str(fs[6]))
      self.send_header("Last-Modified",self.date_time_string(fs.st_mtime))
      self.end_headers()
      return f
    except:
      f.close()
      raise

  def list_directory(self,path):
    try:
      list = os.listdir(path)
    except os.error :
      self.send_error(404,"No permission to list directory")
      return None

    list.sort(key=lambda a:a.lower())
    f = StringIO()

    for name in list:
      fullname = os.path.join(path,name)
      displayname = linkname = name
      if os.path.isdir(fullname):
        displayname = name + "/"
        linkname = name + "/"
      if os.path.islink(fullname):
        displayname = name + "@"
      f.write('<li><a href="%s">%s</a></li>\n'%
        (urllib.quote(linkname),cgi.escape(displayname)))
    length = f.tell()
    f.seek(0)
    self.send_response(200)
    self.send_header("Content-type","text/html;charset=%s"%"UTF-8")
    self.send_header("Content-Length",str(length))
    self.end_headers()
    return f

  def translate_path(self,path):
    path = path.split('?',1)[0]
    path = path.split('#',1)[0]
    trailing_slash = path.rstrip().endswith('/')
    path = posixpath.normpath(urllib.unquote(path))
    words = path.split('/')
    words = filter(None, words)
    path = www_directory
    for word in words:
      drive,word = os.path.splitdrive(word)
      head, word = os.path.split(word)
      if word in (os.curdir,os.pardir):
        continue
      path = os.path.join(path,word)
    if trailing_slash:
      path+='/'
    return path

  def copyfile(self,source ,outputfile):
    shutil.copyfileobj(source,outputfile)

  responses = {
        100: ('Continue', 'Request received, please continue'),
        101: ('Switching Protocols',
              'Switching to new protocol; obey Upgrade header'),

        200: ('OK', 'Request fulfilled, document follows'),
        201: ('Created', 'Document created, URL follows'),
        202: ('Accepted',
              'Request accepted, processing continues off-line'),
        203: ('Non-Authoritative Information', 'Request fulfilled from cache'),
        204: ('No Content', 'Request fulfilled, nothing follows'),
        205: ('Reset Content', 'Clear input form for further input.'),
        206: ('Partial Content', 'Partial content follows.'),

        300: ('Multiple Choices',
              'Object has several resources -- see URI list'),
        301: ('Moved Permanently', 'Object moved permanently -- see URI list'),
        302: ('Found', 'Object moved temporarily -- see URI list'),
        303: ('See Other', 'Object moved -- see Method and URL list'),
        304: ('Not Modified',
              'Document has not changed since given time'),
        305: ('Use Proxy',
              'You must use proxy specified in Location to access this '
              'resource.'),
        307: ('Temporary Redirect',
              'Object moved temporarily -- see URI list'),

        400: ('Bad Request',
              'Bad request syntax or unsupported method'),
        401: ('Unauthorized',
              'No permission -- see authorization schemes'),
        402: ('Payment Required',
              'No payment -- see charging schemes'),
        403: ('Forbidden',
              'Request forbidden -- authorization will not help'),
        404: ('Not Found', 'Nothing matches the given URI'),
        405: ('Method Not Allowed',
              'Specified method is invalid for this resource.'),
        406: ('Not Acceptable', 'URI not available in preferred format.'),
        407: ('Proxy Authentication Required', 'You must authenticate with '
              'this proxy before proceeding.'),
        408: ('Request Timeout', 'Request timed out; try again later.'),
        409: ('Conflict', 'Request conflict.'),
        410: ('Gone',
              'URI no longer exists and has been permanently removed.'),
        411: ('Length Required', 'Client must specify Content-Length.'),
        412: ('Precondition Failed', 'Precondition in headers is false.'),
        413: ('Request Entity Too Large', 'Entity is too large.'),
        414: ('Request-URI Too Long', 'URI is too long.'),
        415: ('Unsupported Media Type', 'Entity body in unsupported format.'),
        416: ('Requested Range Not Satisfiable',
              'Cannot satisfy request range.'),
        417: ('Expectation Failed',
              'Expect condition could not be satisfied.'),

        500: ('Internal Server Error', 'Server got itself in trouble'),
        501: ('Not Implemented',
              'Server does not support this operation'),
        502: ('Bad Gateway', 'Invalid responses from another server/proxy.'),
        503: ('Service Unavailable',
              'The server cannot process the request due to a high load'),
        504: ('Gateway Timeout',
              'The gateway server did not receive a timely response'),
        505: ('HTTP Version Not Supported', 'Cannot fulfill request.'),
        }

def test(HandleClass = SimpleHTTPRequestsHandler,
  ServerClass = BaseHTTPServer.HTTPServer):
  BaseHTTPServer.test(HandleClass,ServerClass)



if __name__ == '__main__':
  test()


