package util

import "testing"

func TestParseTrackerAddressFromUrls(t *testing.T) {
	urls := []string{"udp://open.stealth.si:80/announcee", "udp://tracker.tiny-vps.com:6969/announcee"}
	expectedAddresses := []string{"open.stealth.si:80", "tracker.tiny-vps.com:6969"}
	trackerAddresses := ParseTrackerAddressFromUrls(urls)

	for i := range expectedAddresses {
		if trackerAddresses[i] != expectedAddresses[i] {
			t.Errorf("expected url: %s to parse to: %s but found: %s", urls[0], expectedAddresses[i], trackerAddresses[i])
		}
	}
}
