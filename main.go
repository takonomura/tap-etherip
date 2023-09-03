package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/songgao/water"
	"golang.org/x/sync/errgroup"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: tap-etherip TAP LOCAL REMOTE\n")
		os.Exit(1)
	}

	tapName := os.Args[1]
	localAddr := os.Args[2]
	remoteAddr := os.Args[3]
	remote := &net.IPAddr{IP: net.ParseIP(remoteAddr)}

	c, err := net.ListenPacket("ip6:97", localAddr)
	if err != nil {
		log.Fatalf("listening socket: %v", err)
	}
	defer c.Close()
	tap, err := water.New(water.Config{
		DeviceType: water.TAP,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name:    tapName,
			Persist: true,
		},
	})
	if err != nil {
		log.Fatalf("opening tap: %v", err)
	}
	defer tap.Close()

	log.Printf("TAP: %s LOCAL: %s REMOTE: %s", tapName, localAddr, remote.String())

	eg := new(errgroup.Group)
	eg.Go(func() error {
		buf := make([]byte, 2000)
		for {
			n, addr, err := c.ReadFrom(buf)
			if err != nil {
				return fmt.Errorf("reading from socket: %w", err)
			}
			if n < 2 {
				log.Printf("too short packet: %d bytes", n)
				continue
			}
			if buf[0] != 0x30 || buf[1] != 0x00 {
				log.Printf("invalid EtherIP header: %x", buf[0:2])
				continue
			}
			if ip, ok := addr.(*net.IPAddr); !ok || !ip.IP.Equal(remote.IP) {
				log.Printf("unknown remote: %q", addr.String())
				continue
			}
			n, err = tap.Write(buf[2:n])
			if err != nil {
				return fmt.Errorf("writing to tap: %w", err)
			}
		}
	})
	eg.Go(func() error {
		buf := make([]byte, 2000)
		buf[0] = 0x30
		buf[1] = 0x00
		for {
			n, err := tap.Read(buf[2:])
			if err != nil {
				return fmt.Errorf("reading from tap: %w", err)
			}
			n, err = c.WriteTo(buf[:n+2], remote)
			if err != nil {
				return fmt.Errorf("writing to socket: %w", err)
			}
		}
	})

	log.Fatal(eg.Wait())
}
