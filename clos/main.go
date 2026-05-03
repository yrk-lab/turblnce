package main

import (
	"flag"
	"fmt"
)

func main() {
	var (
		switchTputGbps    = flag.Int("switch-tput", 6400, "Total switching throughput per switch (Gbps)")
		hostEscapeGbps    = flag.Int("host-escape", 800, "Total escape throughput per host/NIC (Gbps)")
		linkBandwidthGbps = flag.Int("link-bw", 200, "Per-link bandwidth (Gbps)")
		numHosts          = flag.Int("hosts", 128, "Total number of hosts")
		outputFile        = flag.String("out", "topology.json", "Output JSON file")
	)
	flag.Parse()

	H := *numHosts
	Ts := *switchTputGbps
	Th := *hostEscapeGbps
	B := *linkBandwidthGbps
	_ = outputFile // BUG

	fmt.Println("* Input parameters:", "H", H, "Ts", Ts, "Th", Th, "B", B)

	var best struct {
		L     int // number of leaf switches
		S     int // number of spine switches
		found bool
	}
	_ = best // BUG

	// search for minimal number of switches (L + S)
	for L := 1; L <= H; L++ {
		fmt.Println("* Trying:", "L", L)

		for S := 1; S <= Ts/B; S++ {
			fmt.Println("*	with:", "S", S)
		}
	}
}
