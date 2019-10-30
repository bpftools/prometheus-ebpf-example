// Copyright 2017 Louis McCormack
// Copyright 2019 Lorenzo Fontana (for the prometheus modifications)
// This file was initially taken from the iovisor/gobpf repository
// https://github.com/iovisor/gobpf/blob/master/examples/bcc/bash_readline/bash_readline.go
// It was modified to expose the extracted metrics as a prometheus endpoint on port 8080
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	bpf "github.com/iovisor/gobpf/bcc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const source string = `
#include <uapi/linux/ptrace.h>

struct readline_event_t {
				u32 pid;
				char str[80];
} __attribute__((packed));

BPF_PERF_OUTPUT(readline_events);

int get_return_value(struct pt_regs *ctx) {
	struct readline_event_t event = {};
	u32 pid;
	if (!PT_REGS_RC(ctx)) {
		return 0;
	}
	pid = bpf_get_current_pid_tgid();
	event.pid = pid;
	bpf_probe_read(&event.str, sizeof(event.str), (void *)PT_REGS_RC(ctx));
	readline_events.perf_submit(ctx, &event, sizeof(event));

	return 0;
}
`

// This is our userspace struct 1:1 with the struct readline_event_t in the eBPF program.
type readlineEvent struct {
	Pid uint32
	Str [80]byte
}

var (
	readlineProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "commands_count",
		Help: "The number of times a command is invoked via bash",
	}, []string{"command", "pid", "nodename"})
)

func main() {
	// URETPROBE_BINARY is the path of the binary (or library) we want to analyze
	binaryName := os.Getenv("URETPROBE_BINARY")

	// Get the current node hostname
	nodeName := os.Getenv("NODENAME")

	if len(nodeName) == 0 {
		nodeName = os.Getenv("HOSTNAME")
	}

	if len(nodeName) == 0 {
		nodeName = "unknown"
	}

	if len(binaryName) == 0 {
		binaryName = "/host/lib/libreadline.so"
	}

	// This creates a new module to compile our eBPF code asynchronously
	m := bpf.NewModule(source, []string{})
	defer m.Close()

	// This loads the uprobe program and sets the "get_return_value" as entrypoint
	readlineUretprobe, err := m.LoadUprobe("get_return_value")
	if err != nil {
		log.Fatalf("Failed to load get_return_value: %v", err)
	}

	// This attaches the uretprobe to the readline function of the passed binary.
	// This will consider every process (old and new) since we didn't specify the pid to look for.
	err = m.AttachUretprobe(binaryName, "readline", readlineUretprobe, -1)
	if err != nil {
		log.Fatalf("Failed to attach return_value: %v", err)
	}

	// This creates a new perf table "readline_events" to look to,
	// this must have the same name as the table defined in the eBPF progrma with BPF_PERF_OUTPUT.
	table := bpf.NewTable(m.TableId("readline_events"), m)

	// This channel will contain our results
	channel := make(chan []byte)

	// Link our channel with the perf table
	perfMap, err := bpf.InitPerfMap(table, channel)
	if err != nil {
		log.Fatalf("Failed to init perf map: %v", err)
	}

	// Defined some handlers ot allow the user to kill the program
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	// Goroutine to handle the events
	go func() {
		var event readlineEvent
		for {

			// Get the current element from the channel
			data := <-channel

			// Read the data and populate the event struct
			err = binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &event)
			if err != nil {
				log.Printf("failed to decode received data: %s", err)
				continue
			}

			// Convert the C string to a Go string
			comm := string(event.Str[:bytes.IndexByte(event.Str[:], 0)])

			readlineProcessed.WithLabelValues(comm, strconv.Itoa(int(event.Pid)), nodeName).Inc()

		}
	}()

	go func() {
		r := prometheus.NewRegistry()
		r.MustRegister(readlineProcessed)
		handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})
		http.Handle("/metrics", handler)
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatalf("error starting the webserver: %v", err)
		}
	}()

	// Start reading
	perfMap.Start()
	// Wait to stop
	<-sig
	// Stop reading
	perfMap.Stop()
}
