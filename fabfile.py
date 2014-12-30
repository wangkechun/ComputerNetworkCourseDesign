from fabric.api import *


def ping2():
	with lcd("bin"):
		local("go build cs/ping2&&sudo ./ping2")

def pingScan():
	local("go build cs/pingScan&&sudo  ./pingScan ")


def traceroute():
	local("go build cs/traceroute&&sudo  ./traceroute baidu.com ")


