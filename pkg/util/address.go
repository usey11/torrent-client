package util

import (
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"
)

func GetLiveTrackerAddress(addrs []string) string {
	log.Debugf("Got %v addresses \n", len(addrs))
	for _, addr := range addrs {
		if LiveAddress(addr) {
			return addr
		}
	}
	return ""
}

func GetLiveTrackerAddresses(addrs []string) []string {
	log.Tracef("Got %v addresses \n", len(addrs))
	liveAddrs := make([]string, 0)
	var wg sync.WaitGroup
	var mx sync.Mutex

	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			if LiveAddress(addr) {
				mx.Lock()
				liveAddrs = append(liveAddrs, addr)
				mx.Unlock()
			}
		}(addr)

	}
	wg.Wait()
	return liveAddrs
}

func LiveAddress(addr string) bool {
	p := strings.Split(addr, ":")

	log.Debugf("Pinging addr: %s\n", p[0])
	pinger, err := ping.NewPinger(p[0])
	pinger.SetPrivileged(true)
	pinger.Timeout = 200 * time.Millisecond
	if err != nil {
		return false
	}

	pinger.Count = 1
	err = pinger.Run()

	if err == nil && pinger.Statistics().PacketsRecv > 0 {
		return true
	}
	return false
}

func ParseTrackerAddressFromUrls(urls []string) []string {
	addresses := make([]string, len(urls))

	for i, url := range urls {
		addresses[i] = ParseTrackerAddressFromUrl(url)
	}
	return addresses
}

func ParseTrackerAddressFromUrl(url string) string {
	s, _ := strings.CutPrefix(url, "udp://")
	p := strings.Split(s, ":")
	p2 := strings.Split(p[1], "/")

	return p[0] + ":" + p2[0]
}

func GetLiveTrackerAddressesFromUrls(urls []string) []string {
	return GetLiveTrackerAddresses(ParseTrackerAddressFromUrls(urls))
}
