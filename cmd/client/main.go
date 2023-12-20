package main

import (
	"encoding/json"
	"tor/pkg/torrent"
	"tor/pkg/util"

	log "github.com/sirupsen/logrus"
)

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func main() {
	// downloadFromFile("C:\\Users\\usa_m\\Downloads\\openttd-13.4-windows-win64.exe.torrent")
	downloadFromMagnet("magnet:?xt=urn:btih:98FF12FB63293C887517917B5CF968431FD96F1A&dn=The.Super.Mario.Bros.Movie.2023.1080p.HDRip.Dual.Audio.X26&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=udp%3A%2F%2Fmovies.zsw.ca%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.dler.org%3A6969%2Fannounce&tr=udp%3A%2F%2Fopentracker.i2p.rocks%3A6969%2Fannounce&tr=udp%3A%2F%2Fopen.stealth.si%3A80%2Fannounce&tr=udp%3A%2F%2Ftracker.0x.tf%3A6969%2Fannounce")
	// metadata()
}

func metadata() {
	log.StandardLogger().SetLevel(log.DebugLevel)

	fileName := "C:\\Users\\usa_m\\Downloads\\Solus-4.4-Budgie.torrent"
	// fileName := "C:\\Users\\usa_m\\Downloads\\Sekiro - Shadows Die Twice [FitGirl Repack](1).torrent"

	// fileName := "C:\\Users\\usa_m\\Downloads\\test.torrent"
	ih, err := util.CalcInfoHash(fileName)
	if err != nil {
		panic(err)
	}

	tf, err := torrent.ParseTorrentFile(fileName)
	if err != nil {
		panic(err)
	}

	pf := torrent.NewTrackersPeerFetcher(ih, util.GetLiveTrackerAddresses(tf.GetTrackerAddresses()))
	ts := torrent.NewTorrentSession(ih, tf.Info, pf)
	ts.GetMetadata()
}

func downloadFromFile(fileName string) {
	// log.StandardLogger().SetLevel(log.DebugLevel)

	// Announce
	// fileName := "C:\\Users\\usa_m\\Downloads\\Solus-4.4-Budgie.torrent"
	// torrentFileName := "C:\\Users\\usa_m\\Downloads\\Sekiro - Shadows Die Twice [FitGirl Repack](1).torrent"

	// fileName := "C:\\Users\\usa_m\\Downloads\\test.torrent"
	ih, err := util.CalcInfoHash(fileName)
	if err != nil {
		panic(err)
	}

	tf, err := torrent.ParseTorrentFile(fileName)
	if err != nil {
		panic(err)
	}

	pf := torrent.NewTrackersPeerFetcher(ih, util.GetLiveTrackerAddresses(tf.GetTrackerAddresses()))

	ts := torrent.NewTorrentSession(ih, tf.Info, pf)
	ts.StartSession()
}

func downloadFromMagnet(uriString string) {
	// uriString = "magnet:?xt=urn:btih:C9523B834E597B4A8926C99E66C84A6AB0B4B520&dn=The+Everything+Solar+Power+For+Beginners+-+2+Books+in+1+-+A+Detailed+Guide+on+How+to+Design+%26amp%3B+install&tr=https%3A%2F%2Finferno.demonoid.is%2Fannounce&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.torrent.eu.org%3A451%2Fannounce&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&tr=udp%3A%2F%2Fexplodie.org%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.moeking.me%3A6969%2Fannounce&tr=udp%3A%2F%2Fexodus.desync.com%3A6969%2Fannounce&tr=udp%3A%2F%2Fipv4.tracker.harry.lu%3A80%2Fannounce&tr=udp%3A%2F%2Fp4p.arenabg.com%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.dler.org%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=http%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&tr=udp%3A%2F%2Fopentracker.i2p.rocks%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969%2Fannounce&tr=udp%3A%2F%2Fcoppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.zer0day.to%3A1337%2Fannounce"
	uri, err := torrent.ParseMagnetUri(uriString)

	ti, err := torrent.GetMetadataFromMagnetUri(uriString)

	if err == nil {
		log.Infof("I got the metadata for: %s", ti.Name)
		// os.Exit(2)
	}
	pf := torrent.NewTrackersPeerFetcher(uri.InfoHash, util.GetLiveTrackerAddresses(util.ParseTrackerAddressFromUrls(uri.Trackers)))
	ts := torrent.NewTorrentSession(uri.InfoHash, *ti, pf)
	ts.StartSession()
}
