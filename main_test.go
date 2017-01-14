package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"regexp"
	"testing"
)

var listenAddrPattern = regexp.MustCompile(`listening on (.+?:\d+)`)

func TestMain(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"emissary"}

	_, stderrWriter, _ := os.Pipe()
	oldStderr := os.Stderr
	os.Stderr = stderrWriter
	defer func() { os.Stderr = oldStderr }()

	t.Log("Testing main without args...")
	main()

	t.Log("Testing main with single upstream...")
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Errorf("Failed to start listener: %s", err)
	}

	stderrReader, stderrWriter, _ := os.Pipe()
	go func(f *os.File) {
		oldStderr := os.Stderr
		os.Stderr = f
		defer func() { os.Stderr = oldStderr }()

		os.Args = []string{
			"emissary",
			"-alsologtostderr",
			"-bind",
			"localhost:0",
			"-upstream",
			fmt.Sprintf("/^GET/:%s", listener.Addr().String()),
		}
		main()
	}(stderrWriter)
	reader := bufio.NewReader(stderrReader)
	line, _, err := reader.ReadLine()
	if err != nil {
		t.Errorf("Failed to read output from cli: %s", err)
	}
	matchAddr := listenAddrPattern.FindStringSubmatch(string(line))
	if len(matchAddr) != 2 {
		t.Errorf("Unexpected output line from cli: %s", line)
	}
	server, err := net.Dial("tcp", matchAddr[1])
	defer server.Close()
	if err != nil {
		t.Errorf("Failed to connect to emissary: %s", err)
	}
	server.Write([]byte("GET /\n"))
	client, err := listener.Accept()
	defer client.Close()
	if err != nil {
		t.Errorf("Failed to accept connection: %s", err)
	}
	// buf := make([]byte, 6)
	// client.Read(buf)
	response := []byte("HTTP/1.0 740 Computer says no")
	client.Write(response)
	buf := make([]byte, len(response))
	server.Read(buf)
	if bytes.Compare(buf, response) != 0 {
		t.Error("Expected responses to match")
		t.Errorf("`%s' != `%s'", response, buf)
	}
}

func TestMainVersion(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"emissary", "-version"}

	stdoutReader, stdoutWriter, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = stdoutWriter
	defer func() { os.Stdout = oldStdout }()

	t.Log("Testing main version flag...")
	main()

	buf := make([]byte, 46)
	stdoutReader.Read(buf)

	expectedOutput := []byte(`emissary version: unknown
build time: unknown
`)
	if bytes.Compare(expectedOutput, buf) != 0 {
		t.Errorf("Unexpected output from `emissary -version'")
	}
}
