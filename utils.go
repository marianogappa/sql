package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func println(w io.Writer, s string) {
	printLock.Lock()
	defer printLock.Unlock()
	fmt.Fprintln(w, s)
}

func readInput(r io.Reader) string {
	ls := []string{}
	var err error
	rd := bufio.NewReader(r)

	for {
		var s string
		s, err = rd.ReadString('\n')

		if err == io.EOF {
			return strings.Join(ls, " ")
		}
		s = strings.TrimSpace(s)
		if len(s) == 0 {
			continue
		}
		ls = append(ls, strings.TrimSpace(s))
	}
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
