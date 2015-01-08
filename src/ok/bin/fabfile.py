from fabric.api import *

def build_all():
	cmd = "go build"
	clear()
	all(cmd)

def cross_all():
	cmd = 'gox -osarch="linux/amd64 windows/386"'
	clear()
	all(cmd)

def test_all():
	cmd = "go test"
	clear()
	all(cmd)

def all(cmd):
	ping2(cmd)
	pingScan(cmd)
	socks5Proxy(cmd)

def ping2(cmd="go build"):
	local(cmd+" cs/ok/ping2")

def pingScan(cmd="go build"):
	local(cmd+" cs/ok/pingScan")

def socks5Proxy(cmd="go build"):
	local(cmd+" cs/ok/socks5Proxy")

def clear():
	local("find . -maxdepth 1 -type f -perm /111 -exec rm {} \;")

