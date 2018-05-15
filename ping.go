package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

// Folder with THIS ping.exe wrapper must me FIRST in PATH
// So to find real ping.exe we stop on latest entry
func get_real_pinger_path() string {
	real_ping_filename := "ping"
	path_separator := ":"
	if runtime.GOOS == "windows" {
		real_ping_filename = "ping.exe"
		path_separator = ";"
	}

	real_pinger_path := ""

	var folders []string
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if strings.EqualFold(pair[0], "pAth") {
			folders = strings.Split(pair[1], path_separator)
		}
	}
	for _, folder := range folders {
		test_path := path.Join(folder, real_ping_filename)
		if _, err := os.Stat(test_path); os.IsNotExist(err) {
			continue
		}
		real_pinger_path = test_path // do not stop, go for last in $PATH
	}
	return real_pinger_path

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func main() {

	var stdoutBuf, stderrBuf bytes.Buffer
	var errStdout, errStderr error

	real_pinger := get_real_pinger_path()
	os.Args = os.Args[1:] // skip self
	cmd := exec.Command(real_pinger)
	if len(os.Args) > 0 {
		for i, _ := range os.Args {
			if strings.HasPrefix(os.Args[i], "http") {
				// URL --> DOMAIN
				u, err := url.Parse(os.Args[i])
				if err != nil {
					panic(err)
				}
				host, _, _ := net.SplitHostPort(u.Host) // for case DOMAIN:PORT
				if host == "" {
					host = u.Host
				}
				os.Args[i] = host
			}

		}
		cmd = exec.Command(real_pinger, os.Args...)
	}

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
	stderr := io.MultiWriter(os.Stderr, &stderrBuf)

	// fmt.Println("Redirecting args to ", real_pinger)
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		_, errStdout = io.Copy(stdout, stdoutIn)
	}()

	go func() {
		_, errStderr = io.Copy(stderr, stderrIn)
	}()

	err = cmd.Wait()
	if errStdout != nil || errStderr != nil {
		log.Fatal("Failed to capture stdout!")
	}
}
