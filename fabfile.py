from fabric.api import *


def rp():
	local("go build pingScan.go&&sudo  ./pingScan ")
