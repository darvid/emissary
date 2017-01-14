package main

import (
	"bytes"
	"fmt"
	"net"
	"testing"
)

func TestNewUpstreamRule(t *testing.T) {
	t.Log("Creating an upstream rule with invalid regexp...")
	rule, err := NewUpstreamRule("/^(whoops/:localhost:8000")
	if err == nil {
		t.Error("Expected a regexp error to be thrown")
	}

	t.Log("Creating an upstream rule with invalid spec...")
	rule, err = NewUpstreamRule("what is this even")
	if err == nil {
		t.Error("Expected a regexp error to be thrown")
	}

	t.Log("Creating an upstream rule with invalid addr...")
	rule, err = NewUpstreamRule("/^GET/:256.256.256.256:80")
	if err == nil {
		t.Error("Expected an error to be thrown from ResolveTCPAddr")
	}

	t.Log("Creating a valid upstream rule...")
	rule, err = NewUpstreamRule("/^GET/:localhost:8000")
	if rule == nil || err != nil {
		t.Error(err)
	}

	if rule.pattern.String() != "^GET" {
		t.Errorf("Invalid pattern captured: %s (expected '^GET')", rule.pattern)
	}

	if rule.addr.String() != "127.0.0.1:8000" {
		t.Errorf("Invalid addr captured: %s (expected '127.0.0.1:8000')", rule.addr)
	}
}

func TestUpstreamRuleListSet(t *testing.T) {
	var rules UpstreamRuleList

	t.Log("Appending valid upstream to the rules list...")
	err := rules.Set("/^GET/:localhost:8000")
	if len(rules) != 1 || err != nil {
		t.Errorf("Upstream rules list empty after calling Set: %s", err)
	}

	t.Log("Appending invalid upstream to the rules list...")
	err = rules.Set("what is this even")
	if err == nil {
		t.Error("Expected error to be thrown for invalid upstream")
	}
}

func TestUpstreamRuleListFindMatch(t *testing.T) {
	var rules UpstreamRuleList
	rules.Set("/^GET/:localhost:8000")
	rules.Set("/^POST/:localhost:9000")
	b := []byte{'\x05', '\x01'}
	t.Log("Trying to find an upstream for non-matching data...")
	chosenRule := rules.FindMatch(&b)
	if chosenRule != nil {
		t.Error("Expected no rule to match as the rule list is empty")
	}

	rules.Set("/^\x05/:localhost:1080")
	t.Log("Trying to find an upstream for matching data...")
	chosenRule = rules.FindMatch(&b)
	if chosenRule == nil || chosenRule.addr.String() != "127.0.0.1:1080" {
		t.Errorf("Expected an upstream rule to match %v", b)
	}
}

func TestUpstreamHandleConn(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Error(err)
	}

	var (
		rules  UpstreamRuleList
		server net.Conn
	)

	client, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Error(err)
	}

	t.Log("Creating rules list with SOCKS upstream rule...")
	rules.Set(fmt.Sprintf("/^\x05/:%s", listener.Addr().String()))

	t.Log("Sending SOCKS greeting...")
	client.Write([]byte{'\x05', '\x02', '\x00', '\x01'})

	defer listener.Close()
	server, err = listener.Accept()
	if err != nil {
		t.Error(err)
	}
	rule, err := rules.HandleConn(server, 4096)
	if rule == nil || err != nil {
		t.Errorf("Expected upstream rule to match: %s", err)
	}

	t.Log("Sending some data...")
	deadbeef := []byte{'\xFE', '\xEB', '\xDA', '\xED'}
	client.Write(deadbeef)
	b := make([]byte, 4)
	server.Read(b)
	if !bytes.Equal(deadbeef, b) {
		t.Errorf("Expected %v, got %v\n", deadbeef, b)
	}
}
