计算机网络课程设计

---

##简单Web Server程序的设计与实现

###相关知识

> WWW的工作基于客户机/服务器计算模型，由Web浏览器(客户机)和Web服务器(服务器)构成，两者之间采用超文本传送协议（HTTP）进行通信，HTTP协议的作用原理包括四个步骤：连接，请求，应答,关闭应答。在课程设计中，系统开发平台为Ubuntu，程序设计语言采用Python。
> 
> Python（英语发音：/ˈpaɪθən/）, 是一种面向对象、解释型计算机程序设计语言，由Guido van
> Rossum于1989年底发明，第一个公开发行版发行于1991年。Python语法简洁而清晰，具有丰富和强大的类库。它常被昵称为胶水语言，它能够把用其他语言制作的各种模块（尤其是C/C++）很轻松地联结在一起。常见的一种应用情形是，使用Python快速生成程序的原型（有时甚至是程序的最终界面），然后对其中有特别要求的部分，用更合适的语言改写，比如3D游戏中的图形渲染模块，性能要求特别高，就可以用C++重写。

> HTTP协议是应用层的协议，定义了服务器端和客户端之间文件传输的沟通方式。HTTP协议用于从WWW服务器传输超文本到本地浏览器的传送协议。由于HTTP协议支持的服务不限于WWW，还可以是其它服务，它允许用户在统一的界面下，采用不同的协议访问不同的服务。它可以使浏览器更加高效，使网络传输减少。它不仅保证计算机正确快速地传输超文本文档，还确定传输文档中的哪一部分，以及哪部分内容首先显示等[3]

###设计原理
根据上述HTTP协议的作用原理，实现GET请求的Web服务器程序的方法如下：  

 1. 创建ServerSocket类对象，监听端口8080。这是为了区别于HTTP的标准TCP/IP端口80而取的 
    等待、接受客户机连接到端口8080，得到与客户机连接的socket
 2. 创建与socket字相关联的输入流instream和输出流outstream
 3. 从与socket关联的输入流instream中读取一行客户机提交的请求信息，请求信息的格式为：GET 路径/文件名 HTTP/1.1
 4. 从请求信息中获取请求类型。如果请求类型是GET，则从请求信息中获取所访问的HTML文件名。没有HTML文件名时，则以index.html作为文件名
 5. 如果HTML文件存在，则打开HTML文件，把HTTP头信息和HTML文件内容通过socket传回给Web浏览器，然后关闭文件。否则发送错误信息给Web浏览器
 6. 关闭与相应Web浏览器连接的socket字,如果浏览器支持KeepAlive,则保持连接直到超时

###系统实现&功能特性
- 项目名称 pythonHTTPServer
- 实现语言 python
- 跨平台 linux/win/arm...
- 使用方式: `python server.py  --port 10001 --www "./"`
- 默认端口: `:8000` 
- 默认www目录: 当前目录
- 支持协议版本为HTTP1.1
- 目前支持GET.
- 支持识别index.html index.htm
- 支持列目录
- 支持404 400等错误
- 支持识别url中的 #,?
- 防止目录遍历漏洞.(通过 ../../之类的url)
- 支持中文目录名和文件名

###核心代码

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


    HTTP_STATUS_MSG = BaseHTTPServer.BaseHTTPRequestHandler.responses

    www_directory = './'


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
            print("GET "+self.path)
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




    def server(port=8000,_www_directory="./"):
        global www_directory
        www_directory = _www_directory
        server_address = ('', port)
        httpd = HTTPServer(server_address, SimpleHTTPRequestsHandler)
        sa = httpd.socket.getsockname()
        print("Serving HTTP on %s, port %s ..." % (sa[0], sa[1]))
        print("open http://127.0.0.1:"+str(port))
        httpd.serve_forever()

    if __name__ == '__main__':
        import argparse
        parser = argparse.ArgumentParser()
        parser.add_argument('--port', help='',default=8000,type=int)
        parser.add_argument('--www', help='',default='./')
        args = parser.parse_args()
        server(args.port,args.www)


------------


##网络代理服务器的设计与实现

###功能需求

> 实现一个简易的proxy程序。proxy程序的功能：能够做“二传手”的工作。它自身处在能同时连通外界目标服务器和我的机器的位置上。我的机器把请求发送给它，它接受请求，把请求原封不动的抄下来发送给外界目标服务器；外界目标服务器响应了请求，把回答发送给它，它再接受回答，把回答原封不动的抄下来发送给我的机器。这样，我的机器实际上是把它当作了目标服务器（由于是原封不动的转抄，请求和回答没有被修改）。而它则是外界目标服务器的客户端。

###相关知识

> SOCKS5 是一个代理协议，它在使用
> TCP/IP协议通讯的前端机器和服务器机器之间扮演一个中介角色，使得内部网中的前端机器变得能够访问Internet网中的服务器，或者使通讯更加安全。SOCKS5
> 服务器通过将前端发来的请求转发给真正的目标服务器，
> 模拟了一个前端的行为。在这里，前端和SOCKS5之间也是通过TCP/IP协议进行通讯，前端将原本要发送给真正服务器的请求发送给SOCKS5服务器，然后SOCKS5服务器将请求转发给真正的服务器。
> 
> Go语言是由Google开发的一个开源项目，目的之一为了提高开发人员的编程效率。
> Go语言语法灵活、简洁、清晰、高效。它对的并发特性可以方便地用于多核处理器
> 和网络开发，同时灵活新颖的类型系统可以方便地编写模块化的系统。go可以快速编译，
> 同时具有垃圾内存自动回收功能，并且还支持运行时反射。Go是一个高效、静态类型， 但是又具有解释语言的动态类型特征的系统级语法。

Socks5协议中文文档（RFC1928） 
http://blog.chinaunix.net/uid-26548237-id-3434356.html

socks5代理默认支持多种身份验证机制. 由于课程设计时间限制,故目前只实现了无验证需求
经过测试可以支持作为chrome浏览器的代理.并支持chrome的通过代理发送的DNS查询请求.

###设计原理

 1. 客户向代理方服务器发出请求信息
 2. 代理方应答
 3. 需要代理方接到应答后发送向代理方发送目的ip和端口
 4. 代理方与目的连接
 5. 代理方将需要代理方发出的信息传到目的方，将目的方发出的信息传到需要代理方。代理完成。
 
####代理TCP连接
1. 向服务器的1080端口建立tcp连接。
2. 向服务器发送 05 01 00 （此为16进制码，以下同）
3 .如果接到 05 00 则是可以代理
4. 发送 05 01 00 01 + 目的地址(4字节） + 目的端口（2字节），目的地址和端口都是16进制码（不是字符串！！）。 例202.103.190.27 -7201 则发送的信息为：05 01 00 01 CA 67 BE 1B 1C 21 (CA=202 67=103 BE=190 1B=27 1C21=7201)
5. 接受服务器返回的自身地址和端口，连接完成
6. 以后操作和直接与目的方进行TCP连接相同。
####代理UDP连接
1. 向服务器的1080端口建立udp连接
2. 向服务器发送 05 01 00
3. 如果接到 05 00 则是可以代理
4. 发送 05 03 00 01 00 00 00 00 + 本地UDP端口（2字节）
5. 服务器返回 05 00 00 01 +服务器地址+端口
6. 需要申请方发送 00 00 00 01 +目的地址IP（4字节）+目的端口 +所要发送的信息
7. 当有数据报返回时 向需要代理方发出00 00 00 01 +来源地址IP（4字节）+来源端口 +接受的信息
注：此为不需要密码的代理协议，只是socks5的一部分，完整协议请RFC1928

###系统实现&功能特性
备注:题目所述为http代理,http代理实现较为简单, 所以这里选择实现了一个socks5代理

socks5代理默认支持多种身份验证机制. 由于课程设计时间限制,故目前只实现了无验证需求
经过测试可以支持作为chrome浏览器的代理.并支持chrome的通过代理发送的DNS查询请求.

这个程序的开发语言选用的是Google的golang,并没有选择常用的c,c++,java
出于性能考虑没有使用脚本语言. 
为了实现跨平台没有选用c,c++. java由于需要虚拟机也没有选用.

 - 支持不完整的socks5协议(不支持身份验证) 
 - 对大文件和多用户处理良好 
 - 可以跨平台linux/win/arm...

###核心代码
    package main

    import (
        "bufio"
        "fmt"
        "github.com/k0kubun/pp"
        "io"
        "log"
        "net"
        "time"
    )

    var _ = bufio.NewReader
    var _ = fmt.Printf
    var _ = log.Fatalln
    var _ = pp.Fatalln

    type Server struct {
    }

    func (s *Server) Serve(l net.Listener) {
        for {
            conn, err := l.Accept()
            if err != nil {
                log.Println(err)
            } else {
                go s.ServeConn(conn)
            }
        }
    }

    func (s *Server) ListenAndServe(network, addr string) error {
        l, err := net.Listen(network, addr)
        if err != nil {
            return err
        }
        s.Serve(l)
        return nil
    }

    func (s *Server) ServeConn(conn net.Conn) {
        defer func() {
            if p := recover(); p != nil {
                log.Println(p)
            }
        }()
        defer conn.Close()

        rBuf := bufio.NewReader(conn)
        wBuf := bufio.NewWriter(conn)
        VER, err := rBuf.ReadByte()
        NMETHODS, err := rBuf.ReadByte()

        if err != nil {
            log.Println(err)
            return
        }
        if VER != byte(5) {
            log.Println("socks5 VER ERROR")
            return
        }
        log.Printf("Have %d NMETHODS\n", NMETHODS)
        methods, err := readMethods(int(NMETHODS), rBuf)
        if err != nil {
            log.Println("readMethods error", err)
        } else {
            log.Println("methods:", methods)
        }
        for _, method := range methods {
            if method == 0 {
                conn.Write([]byte{5, 0})
                ServerConnAuthSuccess(rBuf, wBuf)
                return
            }
        }
        conn.Write([]byte{5, 0xff})
        log.Println("no accept methods")
    }

    func readMethods(n int, r io.Reader) ([]byte, error) {
        methods := make([]byte, n)
        _, err := io.ReadAtLeast(r, methods, n)
        return methods, err
    }

    type AddrSpec struct {
        FQDN string
        IP   net.IP
        Port int
    }

    const (
        connectCommand   = uint8(1)
        bindCommand      = uint8(2)
        associateCommand = uint8(3)
        ipv4Address      = uint8(1)
        fqdnAddress      = uint8(3)
        ipv6Address      = uint8(4)
    )

    func ServerConnAuthSuccess(rBuf *bufio.Reader, wBuf *bufio.Writer) {
        header := []byte{0, 0, 0}
        rBuf.Read(header)
        if header[0] != 5 {
            log.Println("Unsupported command version")
        }
        addrType, _ := rBuf.ReadByte()
        d := &AddrSpec{}
        switch addrType {
        case ipv4Address:
            addr := make([]byte, 4)
            rBuf.Read(addr)
            d.IP = net.IP(addr)
            log.Println("ipv4")
        case ipv6Address:
            addr := make([]byte, 4)
            rBuf.Read(addr)
            d.IP = net.IP(addr)
            log.Println("ipv6")
        case fqdnAddress:
            addrLen, _ := rBuf.ReadByte()
            fqdn := make([]byte, addrLen)
            rBuf.Read(fqdn)
            d.FQDN = string(fqdn)
            log.Println("host", d.FQDN)
        default:
            log.Println("unknow addr")
            return
        }
        port := []byte{0, 0}
        rBuf.Read(port)
        d.Port = (int(port[0]) << 8) | int(port[1])
        // pp.Println(d, string(d.IP.String()))
        if d.FQDN != "" {
            addr, err := net.ResolveIPAddr("ip", d.FQDN)
            if err != nil {
                log.Println("dns resolve error", d.FQDN, addr.String())
                return
            }
            d.IP = addr.IP
        }
        switch header[1] {
        case connectCommand:
            log.Println("connectCommand")
            handleConnect(rBuf, wBuf, d)
        // case bindCommand:
        // case associateCommand:
        default:
            log.Println("unsupported command")
        }
    }

    const (
        successReply uint8 = iota
        serverFailure
        ruleFailure
        networkUnreachable
        hostUnreachable
        connectionRefused
        ttlExpired
        commandNotSupported
        addrTypeNotSupported
    )

    func sendReply(wBuf *bufio.Writer, resp uint8, addr *AddrSpec) error {
        var addrType uint8
        var addrBody []byte
        var addrPort uint16
        switch {
        case addr == nil:
            addrType = ipv4Address
            addrBody = []byte{0, 0, 0, 0}
            addrPort = 0
        case addr.FQDN != "":
            addrType = fqdnAddress
            addrBody = append([]byte{byte(len(addr.FQDN))}, addr.FQDN...)
            addrPort = uint16(addr.Port)
        case addr.IP.To4() != nil:
            addrType = ipv4Address
            addrBody = []byte(addr.IP.To4())
            addrPort = uint16(addr.Port)
        case addr.IP.To16() != nil:
            addrType = ipv6Address
            addrBody = []byte(addr.IP.To16())
            addrPort = uint16(addr.Port)
        default:
            log.Println("failed to format address")
            return nil
        }
        msg := make([]byte, 6+len(addrBody))
        msg[0] = 5
        msg[1] = resp
        msg[2] = 0
        msg[3] = addrType
        copy(msg[4:], addrBody)
        msg[4+len(addrBody)] = byte(addrPort >> 8)
        msg[4+len(addrBody)+1] = byte(addrPort & 0xff)
        _, err := wBuf.Write(msg)
        wBuf.Flush()
        if err != nil {
            log.Println(err)
        }
        return nil
    }

    func handleConnect(rBuf *bufio.Reader, wBuf *bufio.Writer, dest *AddrSpec) {
        log.Println("handleConnect")
        addr := net.TCPAddr{IP: dest.IP, Port: dest.Port}
        target, err := net.DialTCP("tcp", nil, &addr)
        if err != nil {
            log.Println(err)
        }
        defer target.Close()
        local := target.LocalAddr().(*net.TCPAddr)
        bind := AddrSpec{IP: local.IP, Port: local.Port}
        sendReply(wBuf, successReply, &bind)
        errCh := make(chan error, 2)
        go proxy("target", target, rBuf, errCh)
        go proxy("client", wBuf, target, errCh)
        select {
        case _ = <-errCh:
            return
        }
    }

    func proxy(name string, dst io.Writer, src io.Reader, errCh chan error) {
        n, err := io.Copy(dst, src)
        log.Printf("Copied %d bytes to %s", n, name)
        time.Sleep(10 * time.Millisecond)
        errCh <- err
    }

    func init() {
        log.SetFlags(log.Ltime | log.Lshortfile)
    }


----------

##Tracert与Ping程序设计与实现

###功能需求
> 参照附录2，了解Tracert程序的实现原理，并调试通过。然后参考Tracert程序和教材4.4.2节，编写一个Ping程序，并能测试本局域网的所有机器是否在线，运行界面如图1所示的QuickPing程序。


###系统实现

####Ping2
这是对linux下的工具ping的一个简单模仿

    Usage of ./ping2:
      -host="www.baidu.com": usage: -host=baidu.com
      -times=3: usage: -times=3
实现语言: golang , 支持平台 *nix, win, arm...

注意在 *nix下使用需要root权限运行

#####使用说明

    ➜  bin git:(master) ✗ sudo ./ping2
    2015/01/08 14:52:19 Ping  www.baidu.com
    2015/01/08 14:52:19 IP: 180.97.33.108
    2015/01/08 14:52:19 Runing Ping data  [8 0 247 255] [0 0 0 0] 
    2015/01/08 14:52:19 Recv data  [69 0 0 28] [59 75 64 0] [49 1 221 28] [180 97 33 108] [192 168 155 3] [0 0 255 255] [0 0 0 0] 
    2015/01/08 14:52:19 Time spend ms 97
    2015/01/08 14:52:19 
    2015/01/08 14:52:20 Ping  www.baidu.com
    2015/01/08 14:52:20 IP: 180.97.33.108
    2015/01/08 14:52:20 Runing Ping data  [8 0 247 255] [0 0 0 0] 
    2015/01/08 14:52:20 Recv data  [69 0 0 28] [59 158 64 0] [49 1 220 201] [180 97 33 108] [192 168 155 3] [0 0 255 255] [0 0 0 0] 
    2015/01/08 14:52:20 Time spend ms 91
    2015/01/08 14:52:20 
    2015/01/08 14:52:21 Ping  www.baidu.com
    2015/01/08 14:52:21 IP: 180.97.33.108
    2015/01/08 14:52:21 Runing Ping data  [8 0 247 255] [0 0 0 0] 
    2015/01/08 14:52:21 Recv data  [69 0 0 28] [60 35 64 0] [49 1 220 68] [180 97 33 108] [192 168 155 3] [0 0 255 255] [0 0 0 0] 
    2015/01/08 14:52:21 Time spend ms 114

#####核心代码

package main

    import (
        "bytes"
        "encoding/binary"
        "flag"
        "fmt"
        "log"
        "net"
        "time"
    )

    type ICMP struct {
        Type        uint8
        Code        uint8
        Checksum    uint16
        Identifier  uint16
        SequenceNum uint16
    }

    func CheckSum(data []byte) uint16 {
        length := len(data)
        var sum uint32
        var index int
        for length > 1 {
            sum += uint32(data[index])<<8 + uint32(data[index+1])
            index += 2
            length -= 2
        }
        if length > 0 {
            sum += uint32(data[index])
        }
        return uint16(^sum)
    }

    func ping(host string) {
        log.Println("Ping ", host)
        raddr, err := net.ResolveIPAddr("ip", host)
        if err != nil {
            log.Fatalln("resolve ip addr error", err)
        } else {
            log.Println("IP:", raddr)
        }
        conn, err := net.DialIP("ip4:icmp", nil, raddr)
        if err != nil {
            log.Fatalln(err)
        }
        var icmp ICMP
        icmp.Type = 8
        icmp.Code = 0
        icmp.Checksum = 0
        icmp.Identifier = 0
        icmp.SequenceNum = 0
        var buffer bytes.Buffer
        binary.Write(&buffer, binary.BigEndian, icmp)
        icmp.Checksum = CheckSum(buffer.Bytes())
        buffer.Reset()
        binary.Write(&buffer, binary.BigEndian, icmp)
        log.Println("Runing Ping data ", printByte(buffer.Bytes()))
        conn.Write(buffer.Bytes())
        t_start := time.Now()
        conn.SetReadDeadline((time.Now().Add(time.Second * 5)))
        recv := make([]byte, 100)
        recv_len, err := conn.Read(recv)
        if err != nil {
            log.Fatalln(err)
        }

        // log.Println("Recv data check error", recv, buffer.Bytes())
        // }
        log.Println("Recv data ", printByte(recv[:recv_len]))
        t_end := time.Now()
        dur := t_end.Sub(t_start).Nanoseconds() / 1e6
        log.Println("Time spend ms", dur)
        log.Println("")
        defer conn.Close()
    }

    func printByte(b []byte) (r string) {
        l := len(b)
        for i := 0; i < l; i += 4 {
            r += fmt.Sprint(b[i:i+4], " ")
        }
        return
    }

    var host = flag.String("host", "www.baidu.com", "usage: -host=baidu.com")
    var times = flag.Int("times", 3, "usage: -times=3")

    func init() {
        flag.Parse()
    }

    func main() {
        for i := 0; i < *times; i++ {
            // ping("125.221.232.253")
            ping(*host)
            time.Sleep(time.Second / 2)
        }
    }


---------
####pingScan
这是一个简单快速的在线主机扫描工具

实现语言: golang , 支持平台 *nix, win, arm...

#####功能特性

 - 可以快速探测局域网内的在线主机
 - 拥有极快的扫描速度,可以超过nmap
 - 可以命令行指定ping超时时间,以及没有新结果的等待时间
 - 可以及其方便的指定IP段. 如10.1.1-12.1-255 代表10.1.1.1->10.1.12.255之前的所有IP
 

> 并发扫描采用golang自带的goroutine实现(类似于协程),而不是通常的多线程实现,可以同时发起上万请求而不会向使用线程一样占用大量CPU.
> 在linux系统下通过修改一些参数即可实现万级并发,详见
    http://blog.csdn.net/guowake/article/details/6615728



#####使用说明

    ➜  bin git:(master) ✗ sudo ./pingScan   --help
    Usage of ./pingScan:
      -ip="10.1.12.1-255": 
      -n=255: Number of requests to perform
      -t=3000: Millisecond of ping timeout
      -w=3: Second wait after no ans


效果如下

    ➜  bin git:(master) ✗ time fab pingScan_test
    [localhost] local: go build cs/ok/pingScan
    [localhost] local: sudo ./pingScan  -ip="10.1.8.1-255" -n=300 -t=3000 -w=2
    10.1.8.1-255
    now: 10.1.8.1
    all:255 over:44 pre:0.000000
    all:255 over:44 pre:0.000000
    timeout  2
    ip:10.1.8.1 success:true
    ip:10.1.8.112 success:true
    ip:10.1.8.111 success:true
    ip:10.1.8.110 success:true
    ip:10.1.8.109 success:true
    ip:10.1.8.104 success:true
    ...
    ###########################
    success:44 error:0
    Done.
    fab pingScan_test  0.40s user 0.17s system 20% cpu 2.846 total

#####核心代码

package main

    import (
        "bytes"
        "encoding/binary"
        // "io"
        "errors"
        "flag"
        "fmt"
        "io/ioutil"
        "log"
        "net"
        "os"
        "strconv"
        "strings"
        "time"
    )

    var _ = bytes.MinRead
    var _ = binary.MaxVarintLen16
    var _ = fmt.Scanln
    var _ = log.Println
    var _ = time.ANSIC
    var _ = os.DevNull
    var _ = ioutil.NopCloser
    var _ = flag.ContinueOnError

    type ICMP struct {
        Type        uint8
        Code        uint8
        Checksum    uint16
        Identifier  uint16
        SequenceNum uint16
    }

    type PingReturn struct {
        success bool
        msg     string
        host    string
        err     error
    }

    var PingLogger *log.Logger

    func CheckSum(data []byte) uint16 {
        length := len(data)
        var sum uint32
        var index int
        for length > 1 {
            sum += uint32(data[index])<<8 + uint32(data[index+1])
            index += 2
            length -= 2
        }
        if length > 0 {
            sum += uint32(data[index])
        }
        return uint16(^sum)
    }

    func ping(host string, timeLimit int) (re PingReturn) {
        re.success = false
        re.host = host
        PingLogger.Println("Ping ", host)
        raddr, err := net.ResolveIPAddr("ip", host)
        if err != nil {
            PingLogger.Println("resolve ip addr error", err)
            re.msg = "ip error"
            re.err = err
            return
        } else {
            PingLogger.Println("IP:", raddr)
        }
        conn, err := net.DialIP("ip4:icmp", nil, raddr)
        if err != nil {
            if strings.Index(err.Error(), "operation not permitted") != -1 {
                log.Fatalln("operation not permitted, please run it by sudo")
            }
            fmt.Printf("%+v\n", err.Error())
            PingLogger.Println(err)
            re.msg = "dial error"
            re.err = err
            return
        }
        var icmp ICMP
        icmp.Type = 8
        icmp.Code = 0
        icmp.Checksum = 0
        icmp.Identifier = 0
        icmp.SequenceNum = 0

        var buffer bytes.Buffer

        binary.Write(&buffer, binary.BigEndian, icmp)
        icmp.Checksum = CheckSum(buffer.Bytes())
        buffer.Reset()
        binary.Write(&buffer, binary.BigEndian, icmp)
        PingLogger.Println("Runing Ping data ",
            printByte(buffer.Bytes()))
        conn.Write(buffer.Bytes())
        t_start := time.Now()
        conn.SetReadDeadline((time.Now().Add(
            time.Duration(timeLimit) * time.Millisecond)))
        recv := make([]byte, 100)
        recv_len, err := conn.Read(recv)
        if err != nil {
            re.msg = "read error"
            re.err = err
            PingLogger.Println(err)
            return
        }
        PingLogger.Println("Recv data ", printByte(recv[:recv_len]))
        t_end := time.Now()
        dur := t_end.Sub(t_start).Nanoseconds() / 1e6
        PingLogger.Println("Time spend ms", dur)
        PingLogger.Println("")
        re.success = true
        defer conn.Close()
        return
    }

    func printByte(b []byte) (r string) {
        l := len(b)
        for i := 0; i < l; i += 4 {
            r += fmt.Sprint(b[i:i+4], " ")
        }
        return
    }

    func PingList(hostList []string, waitTime int, timeLimit int) {
        successAlive := make([]PingReturn, 0)
        noRet := make(chan PingReturn, 255)
        var ticker *time.Ticker
        ticker = time.NewTicker(time.Second)
        defer ticker.Stop()
        go func() {
            for {
                select {
                case <-ticker.C:
                    fmt.Printf("all:%d over:%d pre:%f\n", len(hostList), len(successAlive), 0.)
                }
            }
        }()
        for _, v := range hostList {
            go func(v string) {
                r := ping(v, timeLimit)
                // print("*")
                noRet <- r
            }(v)
        }
        func() {
            for {
                select {
                case <-time.After(time.Second * time.Duration(waitTime)):
                    fmt.Println("timeout ", waitTime)
                    return
                case r := <-noRet:
                    successAlive = append(successAlive, r)
                    if len(successAlive) == len(hostList) {
                        return
                    }
                    continue
                }
                break
            }
        }()

        var suc, err int
        for _, v := range successAlive {
            if v.success {
                suc++
                fmt.Printf("ip:%s success:%t\n", v.host, v.success)
            } else {
                err++
                // fmt.Println(v.msg, v.err.Error())
            }
        }
        fmt.Printf("###########################\nsuccess:%d error:%d\n", suc, err)

    }

    func max(a, b int) int {
        if a >= b {
            return a
        }
        return b
    }

    func min(a, b int) int {
        if a <= b {
            return a
        }
        return b
    }

    func parseTwoInt(s string) (l, r int, err error) {
        switch strings.Count(s, "-") {
        case 0:
            n, _ := strconv.Atoi(s)
            return n, n + 1, nil
        case 1:
            sp := strings.SplitN(s, "-", 2)
            l, _ = strconv.Atoi(sp[0])
            r, _ = strconv.Atoi(sp[1])
            return l, r + 1, nil
        default:
            return 0, 1, errors.New("IP interval illegal ")
        }
    }

    func parseIPList(IPInterval string) ([]string, error) {
        ip := strings.SplitN(IPInterval, ".", 4)
        ips := make([]string, 0)
        if len(ip) != 4 {
            return ips, errors.New("IP interval illegal ")
        }
        for l0, r0, err := parseTwoInt(ip[0]); l0 < r0; l0++ {
            if err != nil {
                return ips, errors.New("IP interval illegal ")
            }
            for l1, r1, _ := parseTwoInt(ip[1]); l1 < r1; l1++ {
                if err != nil {
                    return ips, errors.New("IP interval illegal ")
                }
                for l2, r2, _ := parseTwoInt(ip[2]); l2 < r2; l2++ {
                    if err != nil {
                        return ips, errors.New("IP interval illegal ")
                    }
                    for l3, r3, _ := parseTwoInt(ip[3]); l3 < r3; l3++ {
                        if err != nil {
                            return ips, errors.New("IP interval illegal ")
                        }
                        now := fmt.Sprintf("%d.%d.%d.%d", l0, l1, l2, l3)
                        ips = append(ips, now)
                    }
                }
            }
        }
        return ips, nil
    }

    func init() {
        PingLogger = log.New(ioutil.Discard,
            "TRACE: ",
            log.Ldate|log.Ltime|log.Lshortfile)
        flag.Parse()
    }

    var ipInterval = flag.String("ip", "10.1.12.1-255", "")
    var numRequests = flag.Int("n", 255, "Number of requests to perform")
    var timeLimit = flag.Int("t", 3000, "Millisecond of ping timeout")
    var waitTime = flag.Int("w", 3, "Second wait after no ans")

    func main() {
        fmt.Println(*ipInterval)
        ips, err := parseIPList(*ipInterval)
        if err != nil {
            log.Fatal(err)
        }
        every_limit := *numRequests
        for i := 0; i < len(ips); i += every_limit {
            fmt.Println("now:", ips[i])
            PingList(ips[i:min(i+every_limit, len(ips))], *waitTime, *timeLimit)
        }
    }



------
##网络聊天程序的设计与实现

###功能需求
参照附录1，了解Socket通信的原理，在此基础上编写一个聊天程序。

###功能特性:
- 实现语言: golang , 支持平台 *nix, win, arm...
- 可以实现群聊,点对点聊天
- 客户端可以使用netcat
- 使用说明, 运行服务器 `./simpleChat --port 9000`
- 然后 客户端运行 `nc localhost 9000`
- 就会提示输入昵称,确定之后就可以进入聊天室
- 输入/name msg可以发送消息msg给昵称为name的用户
- 其它任意字符则为群发
- 任一用户下线或者上线其它用户都会接收到通知.

服务器运行效果:

    ➜  bin git:(master) ✗ fab simpleChat_test
    [localhost] local: go build cs/ok/simpleChat
    [localhost] local: ./simpleChat
    2015/01/08 15:28:20 simpleChat run on  0.0.0.0:6666
    2015/01/08 15:28:20 please type "nc serverip youport" to connect it
    2015/01/08 15:28:20 for example: nc localhost 6666
    TRACE: 2015/01/08 15:28:23 chat.go:117: New user asd 127.0.0.1:46206
    TRACE: 2015/01/08 15:28:23 chat.go:82: brodCast New user asd 127.0.0.1:46206
    TRACE: 2015/01/08 15:28:25 chat.go:63: asd: das
    TRACE: 2015/01/08 15:28:25 chat.go:82: brodCast asd: das
    TRACE: 2015/01/08 15:28:26 chat.go:63: asd: asd
    TRACE: 2015/01/08 15:28:26 chat.go:82: brodCast asd: asd
    TRACE: 2015/01/08 15:28:26 chat.go:63: asd: asd
    TRACE: 2015/01/08 15:28:26 chat.go:82: brodCast asd: asd

客户端运行效果:

    ➜  ~  nc localhost 6666
    Input you name:asd
    You name:asd
    You ip:127.0.0.1:46206
    type "/name msg" to chat 
    type "msg" to broadcast
    Online user 0:
    das
    asd: das
    asd
    asd: asd
    asd
    asd: asd
    ^C

###核心代码

    package main

    import (
        "bufio"
        "flag"
        "fmt"
        "io"
        "log"
        "net"
        "os"
        "strconv"
        "strings"
    )

    var _ = fmt.Println

    var Log *log.Logger

    type Client struct {
        outgoing chan string
        reader   *bufio.Reader
        writer   *bufio.Writer
        name     string
        ip       string
    }

    func dealBr(s string) string {
        return strings.TrimRight(s, "\n")
    }

    func (client *Client) Read() {
        for {
            line, err := client.reader.ReadString('\n')
            if err != nil {
                if err == io.EOF {
                    return
                }
                Log.Println(err)
                delete(onlineClients, client.name)
                brodCast(fmt.Sprintf("user %s leave\n", client.name))
                return
            }
            if len(line) > 2 && line[0] == '/' {
                msg := line[1:]
                sp := strings.SplitN(msg, " ", 2)
                Log.Println(sp)
                if len(sp) == 2 {
                    name := sp[0]
                    msg := sp[1]
                    msg = fmt.Sprintf("msg from %s to %s : %s\n", client.name, name, msg)
                    Log.Println(msg)
                    to, ok := onlineClients[name]
                    if ok {
                        to.outgoing <- msg
                    } else {
                        msg := "no such user online"
                        Log.Println(msg)
                        client.outgoing <- msg
                    }
                }
            } else {
                msg := fmt.Sprintf("%s: %s\n", client.name, line)
                Log.Println(dealBr(msg))
                brodCast(msg)
            }
        }
    }

    func (client *Client) Write() {
        for data := range client.outgoing {
            client.writer.WriteString(dealBr(data) + "\n")
            client.writer.Flush()
        }
    }

    func (client *Client) Listen() {
        go client.Read()
        go client.Write()
    }

    func brodCast(msg string) {
        Log.Println("brodCast", dealBr(msg))
        for _, v := range onlineClients {
            v.outgoing <- dealBr(msg) + "\n"
        }
    }

    func NewClient(connection net.Conn) *Client {
        writer := bufio.NewWriter(connection)
        reader := bufio.NewReader(connection)
        writer.WriteString("Input you name:")
        writer.Flush()
        name, _ := reader.ReadString('\n')
        name = strings.Replace(name, "\n", "", -1)
        client := &Client{
            outgoing: make(chan string),
            reader:   reader,
            writer:   writer,
            ip:       connection.RemoteAddr().String(),
            name:     name,
        }
        msg := fmt.Sprintf(
            `You name:%s
    You ip:%s
    type "/name msg" to chat 
    type "msg" to broadcast
    Online user %d:
    `,
            name, client.ip, len(onlineClients))
        for _, v := range onlineClients {
            line := fmt.Sprintf("  name: %s , ip: %s\n", v.name, v.ip)
            msg += line
        }
        writer.WriteString(msg)
        writer.Flush()
        msg = fmt.Sprintf("New user %s %s\n", name, client.ip)
        Log.Println(dealBr(msg))
        brodCast(msg)
        client.Listen()
        return client
    }

    var onlineClients map[string](*Client)
    var port = flag.Int("port", 6666, "port=6666 ")

    func main() {
        onlineClients = make(map[string](*Client), 0)
        listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(*port))
        if err != nil {
            Log.Fatalln(err)
        }
        log.Println("simpleChat run on ", "0.0.0.0:"+strconv.Itoa(*port))
        log.Println("please type \"nc serverip youport\" to connect it")
        log.Println("for example: nc localhost 6666")
        for {
            conn, _ := listener.Accept()
            c := NewClient(conn)
            onlineClients[c.name] = c
        }
    }
    func init() {
        Log = log.New(os.Stdin,
            "TRACE: ",
            log.Ldate|log.Ltime|log.Lshortfile)
        flag.Parse()
    }




##心得体会
课程设计是一次奇妙的旅行，每次拿到课程设计指导书，翻开一看什么都不会，但是在通过不断地摸索以及不断对查阅资料进行总结，分析，验证，对自己所选择的题目越来越清晰，直到最后编程，调试，运行成功实现指定的功能的过程中，让我的学习能力，查阅资料和分析资料的能力有很大的提高。同时，在编程的过程中还会遇到很多的错误，在不断调试的过程中，我发现自己的编程能力也有很大的提高。理论应该运用于实践，这样才能够深入了掌握计算机网络这门学科。课程设计是一个提高自己能力的好方法，只要认真对待，分析自己遇到的问题并且尝试独立解决，都会有所收获。

