package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/ringbuf"
)

// Define the same struct as used in your eBPF C program
// For example:
type Event struct {
	SrcIP   [4]byte
	DstIP   [4]byte
	SrcPort uint16
	DstPort uint16
	Len     uint32
}

func main() {
	// Path to your pinned ringbuf map
	const mapPath = "/sys/fs/bpf/xdp_maps/events"

	// Load pinned map
	m, err := ebpf.LoadPinnedMap(mapPath, &ebpf.LoadPinOptions{})
	if err != nil {
		log.Fatalf("failed to load pinned map: %v", err)
	}
	defer m.Close()

	// Open ring buffer reader
	rd, err := ringbuf.NewReader(m)
	if err != nil {
		log.Fatalf("failed to create ringbuf reader: %v", err)
	}
	defer rd.Close()

	log.Println("Listening for events... (Ctrl+C to exit)")

	// Handle Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// Read events
	go func() {
		for {
			record, err := rd.Read()
			if err != nil {
				if err == ringbuf.ErrClosed {
					return
				}
				log.Printf("ringbuf read error: %v", err)
				continue
			}

			var e Event
			err = binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &e)
			if err != nil {
				log.Printf("decode error: %v", err)
				continue
			}

			fmt.Printf("Event: SrcPort=%d DstPort=%d Len=%d\n", e.SrcPort, e.DstPort, e.Len)
		}
	}()

	<-sig
	log.Println("Exiting...")
}
