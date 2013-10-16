package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	var (
		gmd  = flag.String("gmd", "localhost:5432", "gmd UDP endpoint")
		file = flag.String("file", "", "session file")
	)
	flag.Parse()

	addr, err := net.ResolveUDPAddr("udp", *gmd)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	f, err := os.Open(*file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	last := uint64(0)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		toks := strings.SplitN(line, " ", 2)
		if len(toks) != 2 {
			log.Fatalf("invalid line (split=%d): %s", len(toks), line)
		}

		ts, err := strconv.ParseUint(toks[0], 10, 64)
		if err != nil {
			log.Fatalf("invalid line (parse=%s): %s", err, line)
		}

		if last > 0 {
			time.Sleep(time.Duration(ts-last) * time.Nanosecond)
		}
		last = ts

		conn.Write([]byte(toks[1]))
	}
}
