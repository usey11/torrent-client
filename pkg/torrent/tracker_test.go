package torrent

// func TestAnnounceRequest(t *testing.T) {
// 	fileName := "C:\\Users\\usa_m\\Downloads\\Solus-4.4-Budgie.torrent"

// 	tf, err := ParseTorrentFile(fileName)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	ih, err := tf.Info.CalcInfoHash()
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	r := AnnounceRequest{
// 		InfoHash: ih,
// 		Left:     100,
// 		NumWant:  -1,
// 		Port:     6881,
// 	}

// 	addresses := tf.GetAllTrackerAddresses()
// 	var trackerCon *TrackerConn
// 	for _, addr := range addresses {
// 		trackerCon, err = NewUDPTrackerConn(addr)
// 		if err != nil {
// 			continue
// 		}

// 		resp, err := trackerCon.Announce(r)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		fmt.Println(len(resp.Peers))
// 		// for _, peer := range resp.Peers {
// 		// 	fmt.Println(peer.IpAddr)
// 		// }
// 	}
// }
