package main

import (
	"flag"
	"fmt"
	"io"
	"log"
)

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
	_ = outputFile // BUG

	fmt.Println("Input parameters:")
	fmt.Println("	Hosts: ", H)
	fmt.Println("		-hosts", H)
	fmt.Println("	Switch throughput, Gbps (Ts):", Ts)
	fmt.Println("		-switch-tput", Ts)
	fmt.Println("	Host throughput, Gbps (Th):", Th)
	fmt.Println("		-host-escape", Th)
	fmt.Println("	Link bandwidth, Gbps (B):", B)
	fmt.Println("		-link-bw", B)

	var debug *log.Logger
	if *verbose {
		debug = log.Default()
		log.Default().SetFlags(0)	// less clutter
	} else {
		debug = log.New(io.Discard, "", 0)
	}

	var best struct {
		L     int // number of leaf switches
		S     int // number of spine switches
		K     int // number of links per leaf↔spine pair
		found bool
	}

	// search for minimal number of switches (L + S)
	for L := 1; L <= H; L++ {
		debug.Println("* Trying:", "L", L)

		hpl := H / L // hosts per leaf
		debug.Println("*	hpl", hpl)

		if hpl*Th > Ts {
			debug.Println("!	leaf over capacity:", "hpl", hpl, "Th", Th, "hpl*Th", hpl*Th, ">", "Ts", Ts)
			continue
		}

		for S := 1; S <= Ts/B; S++ {
			debug.Println("*	with S", S)
			K := hpl * Th / (S * B) // leaf↔spine links
			if K < 1 {
				debug.Println("!	no leaf↔spine links:", "hpl", hpl, "Th", Th, "S", S, "B", B, "hpl*Th", hpl*Th, "S*B", S*B, "K=hpl*Th / (S*B)", K)
				continue
			}
			debug.Println("*		", "K", K)

			if L*B*K > Ts {
				debug.Println("!	spine over capacity:", "L", L, "B", B, "K", K, "L*B*K", L*B*K, ">", "Ts", Ts)
				continue
			}

			if !best.found || L+S < best.L+best.S {
				debug.Println("*		new best:", "L", L, "+", "S", S, "L+S", L+S, "<", best.L+best.S)
				best.found = true
				best.L, best.S, best.K = L, S, K
			}
		}
	}
	if !best.found {
		log.Fatalf("no 2-layer Clos found with given parameters")
	}
	fmt.Println("Chosen topology:")
	fmt.Println("	Hosts: ", H)
	fmt.Println("	Leaves (L):", best.L)
	fmt.Println("	Spines (S):", best.S)
	fmt.Println("	Links per leaf-spine pair (K):", best.K)
	fmt.Println("	Hosts per leaf (hpl):", H/best.L)
	fmt.Println("	Links per host (lph):", Th/B)
}
