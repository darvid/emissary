package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"

	"github.com/golang/glog"
)

// An UpstreamRule is a regex pattern associated with a remote address.
type UpstreamRule struct {
	pattern *regexp.Regexp
	addr    *net.TCPAddr
}

// NewUpstreamRule creates a new UpstreamRule containing a compiled regular
// expression, and resolved remote address.
func NewUpstreamRule(upstream string) (*UpstreamRule, error) {
	upstreamPattern, _ := regexp.Compile(`^/(.+?)/:(.+?:\d+)`)
	result := upstreamPattern.FindStringSubmatch(upstream)
	if len(result) != 3 {
		return nil, errors.New("invalid upstream specifier")
	}
	rulePattern, err := regexp.Compile(result[1])
	if err != nil {
		return nil, err
	}
	addr, err := net.ResolveTCPAddr("tcp", result[2])
	if err != nil {
		return nil, err
	}
	return &UpstreamRule{rulePattern, addr}, nil
}

// UpstreamRuleList represents a list of UpstreamRules.
type UpstreamRuleList []*UpstreamRule

// FindMatch attempts to find a matching UpstreamRule given a byte array.
func (rules *UpstreamRuleList) FindMatch(buf *[]byte) *UpstreamRule {
	for _, rule := range *rules {
		if !rule.pattern.Match(*buf) {
			continue
		}
		return rule
	}
	return nil
}

// HandleConn performs a Read on the given net.Conn, and serves as a
// reverse proxy to the matching remote upstream, if one was found that
// matched on the Read.
func (rules *UpstreamRuleList) HandleConn(
	conn net.Conn,
	bufSize int) (*UpstreamRule, error) {
	glog.Infof("handling connection from %s", conn.RemoteAddr().String())
	buf := make([]byte, bufSize)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	rule := rules.FindMatch(&buf)
	if rule == nil {
		glog.Warningf("no upstream found for %s", conn.RemoteAddr().String())
		conn.Close()
		glog.Infoln("closed connection from %s", conn.RemoteAddr().String())
		return nil, nil
	}
	glog.Infof("found matching upstream: %s", rule.addr.String())
	dest, err := net.Dial("tcp", rule.addr.String())
	if err != nil {
		return nil, err
	}
	go func(source net.Conn, dest net.Conn) {
		go dest.Write(buf[:n])
		go io.Copy(dest, conn)
		go io.Copy(conn, dest)
	}(conn, dest)
	return rule, nil
}

// Set parses a UpstreamRule specification from the given string and
// appends it to the UpstreamRuleList.
//
// UpstreamRule specifiers must be provided in the following format:
//
// 	'/pattern/:addr:port'
//
// Pattern is a regular expression, and is matched from the start of any
// incoming data stream. Address may be an IPv4 address or a hostname.
func (rules *UpstreamRuleList) Set(value string) error {
	upstreamRule, err := NewUpstreamRule(value)
	if err != nil {
		return err
	}
	*rules = append(*rules, upstreamRule)
	return nil
}

func (rules *UpstreamRuleList) String() string {
	return fmt.Sprint(*rules)
}
