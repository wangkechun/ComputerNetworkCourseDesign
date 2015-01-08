计算机网络课程设计

---

##简单Web Server程序的设计与实现
###pythonHTTPServer
- 跨平台 linux/win/arm...
- 这是用python实现的一个简单的HTTP服务器
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

------------

##网络代理服务器的设计与实现

####题目:

> 实现一个简易的proxy程序。proxy程序的功能：能够做“二传手”的工作。它自身处在能同时连通外界目标服务器和我的机器的位置上。我的机器把请求发送给它，它接受请求，把请求原封不动的抄下来发送给外界目标服务器；外界目标服务器响应了请求，把回答发送给它，它再接受回答，把回答原封不动的抄下来发送给我的机器。这样，我的机器实际上是把它当作了目标服务器（由于是原封不动的转抄，请求和回答没有被修改）。而它则是外界目标服务器的客户端。

题目所述为http代理,http代理实现较为简单, 所以这里选择实现了一个socks5代理(的一部分)

> SOCKS5 是一个代理协议，它在使用
> TCP/IP协议通讯的前端机器和服务器机器之间扮演一个中介角色，使得内部网中的前端机器变得能够访问Internet网中的服务器，或者使通讯更加安全。SOCKS5
> 服务器通过将前端发来的请求转发给真正的目标服务器，
> 模拟了一个前端的行为。在这里，前端和SOCKS5之间也是通过TCP/IP协议进行通讯，前端将原本要发送给真正服务器的请求发送给SOCKS5服务器，然后SOCKS5服务器将请求转发给真正的服务器。

Socks5协议中文文档（RFC1928） 
http://blog.chinaunix.net/uid-26548237-id-3434356.html

socks5代理默认支持多种身份验证机制. 由于课程设计时间限制,故目前只实现了无验证需求
经过测试可以支持作为chrome浏览器的代理.并支持chrome的通过代理发送的DNS查询请求.

这个程序的开发语言选用的是Google的golang,并没有选择常用的c,c++,java
出于性能考虑没有使用脚本语言. 
为了实现跨平台没有选用c,c++. java由于需要虚拟机也没有选用.

> Go语言是由Google开发的一个开源项目，目的之一为了提高开发人员的编程效率。
> Go语言语法灵活、简洁、清晰、高效。它对的并发特性可以方便地用于多核处理器
> 和网络开发，同时灵活新颖的类型系统可以方便地编写模块化的系统。go可以快速编译，
> 同时具有垃圾内存自动回收功能，并且还支持运行时反射。Go是一个高效、静态类型， 但是又具有解释语言的动态类型特征的系统级语法。

这份代码主要部分来源于我以前写的一个项目 ,
https://github.com/wangkechun/go-multi-IP-acceleration

以下是该项目的中文说明

> go-multi-IP-acceleration
> go-multi-IP-acceleration是用golang实现的一个socks5代理，它可以把流量随机从多个本地IP发出，在某些网络环境下可以达到加速网络的效果
> 
> 使用说明：
> 
> 下载golang开发环境，然后编译对应平台的二进文件,也可以直接从release中下载。 用nmap等工具扫描出一个空余IP列表
> 然后在本地机器加上多个IP,(windows在设置IP的对话框选择高级，然后加上多个IP)，linux自己搜索相关资料。
> 然后把goproxy.ini放到和二进制文件一个目录，并修改配置文件 配置文件： port 为http代理功能保留，暂空
> port_socks5 代表监听端口 local ip 本机IP地址列表 allowed ip 允许使用代理的IP列表，暂未实现 然后运行，
> 打开迅雷，进入设置->高级设置->代理设置，然后添加代理，服务器 127.0.0.1，
> 端口和port_socks5一致，类型socks5，点测试，如果成功的话就ok了。建议把任务默认属性->原始地线程数调到10
> 
> TODO：
> 
> 身份验证 http代理 自动翻墙 自动检测IP冲突

简而言之就是可以逸夫楼的网络下实现带宽叠加(突破路由器的带宽分配策略),迅雷下载速度轻松突破1MB/s. 

Todo:

- 同时支持HTTP代理和SOCKS5代理
- 自动检测被墙网站,并使用goagent ,shadowsocks作为二级代理
- 支持简单的负载均衡
- 实现一个类似shadowsocks的工具,可以实现简单的翻墙功能
- 实现ssh隧道类似的功能, 可以用于内网端口转发,内网映射(如:在校外访问校内资源)


----------
##Tracert与Ping程序设计与实现

> 参照附录2，了解Tracert程序的实现原理，并调试通过。然后参考Tracert程序和教材4.4.2节，编写一个Ping程序，并能测试本局域网的所有机器是否在线，运行界面如图1所示的QuickPing程序。

------

###Ping2
这是对linux下的工具ping的一个简单模仿

    Usage of ./ping2:
      -host="www.baidu.com": usage: -host=baidu.com
      -times=3: usage: -times=3
实现语言: golang , 支持平台 *nix, win, arm...

注意在 *nix下使用需要root权限运行

####使用效果如下:

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

---------
###pingScan
这是一个简单快速的在线主机扫描工具

实现语言: golang , 支持平台 *nix, win, arm...

功能: 

 - 可以快速探测局域网内的在线主机
 - 拥有极快的扫描速度,可以超过nmap
 - 可以命令行指定ping超时时间,以及没有新结果的等待时间
 - 可以及其方便的指定IP段. 如10.1.1-12.1-255 代表10.1.1.1->10.1.12.255之前的所有IP
 

> 并发扫描采用golang自带的goroutine实现(类似于协程),而不是通常的多线程实现,可以同时发起上万请求而不会向使用线程一样占用大量CPU.
> 在linux系统下通过修改一些参数即可实现万级并发,详见
    http://blog.csdn.net/guowake/article/details/6615728



使用说明:

    ➜  bin git:(master) ✗ sudo ./pingScan   --help
    Usage of ./pingScan:
      -ip="10.1.12.1-255": 
      -n=255: Number of requests to perform
      -t=3000: Millisecond of ping timeout
      -w=3: Second wait after no ans


效果如下:

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


------
##网络聊天程序的设计与实现
###simpleChat
功能:
实现语言: golang , 支持平台 *nix, win, arm...

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

