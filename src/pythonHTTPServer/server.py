import os
import posixpath
import BaseHTTPServer
import urllib
import cgi
import sys
import shutil
import mimetypes

try:
	from cStringIO import StringIO
except ImportError :
	from StringIO import StringIO


www_directory = "/home/wkc/project/ComputerNetworkCourseDesign/src/pythonHTTPServer/v3.bootcss.com"
www_directory = "/home/wkc/project/ComputerNetworkCourseDesign/src/pythonHTTPServer/"

class SimpleHTTPRequestsHandler(BaseHTTPServer.BaseHTTPRequestHandler):
	def do_GET(self):
		f = self.send_head()
		if f:
			try:
				self.copyfile(f,self.wfile)
			finally:
				f.close()

	def do_HEAD(self):
		f = self.send_head()
		if f:
			f.close()

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
		# ctype = self.guess_type(path)
		try:
			f = open(path,"rb")
		except IOError:
			self.send_error(404,"File not found")
			return None
		try:
			self.send_response(200)
			# self.send_header("Content-type",ctype )
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

def test(HandleClass = SimpleHTTPRequestsHandler,
	ServerClass = BaseHTTPServer.HTTPServer):
	BaseHTTPServer.test(HandleClass,ServerClass)



if __name__ == '__main__':
	test()


