package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

// Link structure describes a single link in the topology output
type Link struct {
	Src   int `json:"src"`
	Dst   int `json:"dst"`
	Speed int `json:"speed_gbps"`
}

func main() {
	var (
		switchTputGbps    = flag.Int("switch-tput", 6400, "Total switching throughput per switch (Gbps)")
		hostEscapeGbps    = flag.Int("host-escape", 800, "Total escape throughput per host/NIC (Gbps)")
		linkBandwidthGbps = flag.Int("link-bw", 200, "Per-link bandwidth (Gbps)")
		numHosts          = flag.Int("hosts", 128, "Total number of hosts")
		outputFile        = flag.String("out", "topology.json", "Output JSON file")
		verbose           = flag.Bool("v", false, "Verbose mode (very)")
	)
	flag.Parse()

	H := *numHosts
	Ts := *switchTputGbps
	Th := *hostEscapeGbps
	B := *linkBandwidthGbps
	hl := Th / B

	fmt.Println("Input parameters:")
	fmt.Println("	Hosts: ", H)
	fmt.Println("		-hosts", H)
	fmt.Println("	Switch throughput, Gbps (Ts):", Ts)
	fmt.Println("		-switch-tput", Ts)
	fmt.Println("	Host escape throughput, Gbps (Th):", Th)
	fmt.Println("		-host-escape", Th)
	fmt.Println("	Link bandwidth, Gbps (B):", B)
	fmt.Println("		-link-bw", B)
	fmt.Println("	Output JSON file:", *outputFile)
	fmt.Println("		-out", *outputFile)

	var debug *log.Logger
	if *verbose {
		debug = log.Default()
		log.Default().SetFlags(0) // less clutter
	} else {
		debug = log.New(io.Discard, "", 0)
	}

	// L - number of leaf switches
	// S - number of spine switches
	// K - number of links per leaf↔spine pair
	var L, S, K int
	var found bool

	// search for minimal number of switches (l + s)
	for l := 1; l <= H; l++ {
		hpl := H / l
		debug.Println("* Trying:", "l", l, "hpl", hpl)
		if hpl*Th > Ts {
			debug.Println("!	leaf over capacity:", "hpl", hpl, "Th", Th, "hpl*Th", hpl*Th, ">", "Ts", Ts)
			continue
		}
		for s := 1; s <= Ts/B; s++ {
			k := hpl * Th / (s * B) // leaf↔spine links
			debug.Println("*	with s", s, "k", k)
			if k < 1 {
				debug.Println("!	no leaf↔spine links:", "hpl", hpl, "Th", Th, "s", s, "B", B, "hpl*Th", hpl*Th, "s*B", s*B, "k", k)
				continue
			}

			if l*k*B > Ts {
				debug.Println("!	spine over capacity:", "l", l, "B", B, "k", k, "l*k*B", l*k*B, ">", "Ts", Ts)
				continue
			}

			if !found || l+s < L+S {
				debug.Println("*		new best:", "l", l, "+", "s", s, "l+s", l+s, "<", L+S)
				found = true
				L, S, K = l, s, k
			}
		}
	}
	if !found {
		log.Fatalf("no 2-layer Clos found with given parameters")
	}
	hpl := H / L

	fmt.Println("Chosen topology:")
	fmt.Println("	Hosts: ", H)
	fmt.Println("	Leaves (L):", L, "Spines (S):", S)
	fmt.Println("	Hosts/leaf (hpl):", hpl)
	fmt.Println("	Links/host (hl):", hl)
	fmt.Printf("	Leaf ports: hosts = %d, uplink = %d (%d Gbps)\n", hl*hpl, S*K, hpl*Th)
	fmt.Printf("	Spine ports: %d (%d Gbps)\n", L*K, L*K*B)

	// Render the links
	//
	var links []Link

	leaf0 := H
	spine0 := H + L

	// host↔leaf, aggregated
	for h := 0; h < H; h++ {
		l := h / hpl
		debug.Println("* h↔l:", "host", h, "leaf", leaf0+l, "speed", hl*B, "n", hl)
		links = append(links, Link{Src: h, Dst: leaf0+l, Speed: hl * B})
	}

	// leaf↔spine, aggregated
	for l := 0; l < L; l++ {
		for s := 0; s < S; s++ {
			debug.Println("* l↔s:", "leaf", leaf0+l, "spine", spine0+s, "speed", K*B, "n", K)
			links = append(links, Link{Src: leaf0+l, Dst: spine0+s, Speed: K * B})
		}
	}

	// Write JSON
	//
	f, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ") // readability - remove to produce a smaller file (would read with jq)
	if err := enc.Encode(links); err != nil {
		log.Fatalf("failed to write JSON: %v", err)
	}

	fmt.Printf("\nWrote %d aggregated links to %s\n", len(links), *outputFile)
}
