from fabric.api import *


def rp():
	local("go build pingScan.go&&sudo  ./pingScan ")


def tr():
	local("go build tracert.go&&sudo  ./tracert ")


def echo():
	local("go build echo.go&&sudo  ./echo ")

def tr2():
	local("go build traceroute.go&&sudo  ./traceroute baidu.com ")
