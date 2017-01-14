package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/golang/glog"
)

var (
	version   = "unknown"
	buildTime = "unknown"

	bindAddr      string
	bufferSize    int
	upstreamRules UpstreamRuleList
	showVersion   bool
)

func init() {
	flag.IntVar(&bufferSize, "buffersize", 4096, "buffer size for first read")
	flag.StringVar(&bindAddr, "bind", "localhost:1080", "bind address")
	flag.Var(&upstreamRules, "upstream", "list of upstream rules")
	flag.BoolVar(&showVersion, "version", false, "show version")
}

func main() {
	flag.Parse()
	if showVersion {
		fmt.Printf("emissary version: %s\n", version)
		fmt.Printf("build time: %s\n", buildTime)
	} else {
		if len(upstreamRules) == 0 {
			flag.PrintDefaults()
		} else {
			listener, err := net.Listen("tcp", bindAddr)
			if err != nil {
				glog.Fatalln(err)
			}
			defer listener.Close()
			glog.Infof("listening on %s", listener.Addr().String())

			for {
				conn, err := listener.Accept()
				defer conn.Close()
				if err != nil {
					panic(err)
				}
				go upstreamRules.HandleConn(conn, bufferSize)
			}
		}
	}
	return
}
