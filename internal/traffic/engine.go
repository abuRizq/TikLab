package traffic

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
)

// Engine runs traffic generation across namespaces with 40/45/15 distribution.
type Engine struct {
	Gateway    string
	HTTPPort   int
	IperfPort  int
	Namespaces []string
	wg         sync.WaitGroup
}

// NewEngine creates a traffic engine for the given gateway and namespace list.
func NewEngine(gateway string, namespaces []string) *Engine {
	return &Engine{
		Gateway:    gateway,
		HTTPPort:   8080,
		IperfPort:  5201,
		Namespaces: namespaces,
	}
}

// Start begins traffic generation. Blocks until Stop is called.
func (e *Engine) Start(ctx context.Context) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("traffic engine requires Linux")
	}
	// Start HTTP server for browsing traffic
	httpAddr := fmt.Sprintf("%s:%d", e.Gateway, e.HTTPPort)
	if err := e.startHTTPServer(); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	// Start iperf3 server for heavy traffic (runs on host, reachable via bridge)
	go e.runIperfServer()

	// Assign types and run traffic in each namespace (40/45/15)
	n := len(e.Namespaces)
	for i, nsName := range e.Namespaces {
		nsExec := e.nsExec(nsName)
		t := AssignType(i, n)
		e.wg.Add(1)
		go func(ns string, typ Type) {
			defer e.wg.Done()
			switch typ {
			case TypeIdle:
				_ = RunIdle(ctx, nsExec, e.Gateway)
			case TypeBrowsing:
				url := "http://" + httpAddr + "/"
				if e.k6Available(nsExec) {
					_ = RunBrowsing(ctx, nsExec, url)
				} else {
					_ = RunBrowsingCurl(ctx, nsExec, url)
				}
			case TypeHeavy:
				_ = RunHeavy(ctx, nsExec, e.Gateway, e.IperfPort)
			}
		}(nsName, t)
	}
	e.wg.Wait()
	return nil
}

func (e *Engine) nsExec(nsName string) func(string, ...string) *exec.Cmd {
	return func(name string, args ...string) *exec.Cmd {
		all := append([]string{"netns", "exec", nsName, name}, args...)
		return exec.Command("ip", all...)
	}
}

func (e *Engine) k6Available(nsExec func(string, ...string) *exec.Cmd) bool {
	return nsExec("k6", "version").Run() == nil
}

func (e *Engine) startHTTPServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	addr := fmt.Sprintf("%s:%d", e.Gateway, e.HTTPPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go http.Serve(listener, mux)
	return nil
}

func (e *Engine) runIperfServer() {
	cmd := exec.Command("iperf3", "-s", "-p", fmt.Sprintf("%d", e.IperfPort))
	_ = cmd.Run()
}
