package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

type threadSafePrintliner struct {
	l sync.Mutex
	w io.Writer
}

func newThreadSafePrintliner(w io.Writer) *threadSafePrintliner {
	return &threadSafePrintliner{w: w}
}

func (p *threadSafePrintliner) println(s string) {
	p.l.Lock()
	fmt.Fprintln(p.w, s)
	p.l.Unlock()
}

func readQuery(r io.Reader) string {
	s, _ := ioutil.ReadAll(r) // N.B. not interested in this error; might as well return an empty string
	return strings.TrimSpace(strings.Replace(string(s), "\n", " ", -1))
}

func trimEmpty(s []string) []string {
	var r = make([]string, 0)
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func awaitSignal(cancel context.CancelFunc) {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	cancel()
}
