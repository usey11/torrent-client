package torrent

import (
	"os"
	"tor/pkg/bencode"
)

func ParseTorrentFile(fileName string) (Torrent, error) {
	f, err := os.ReadFile(fileName)

	if err != nil {
		return Torrent{}, err
	}

	contents, err := bencode.Decode(f)
	if err != nil {
		panic(err)
	}

	fileDict, ok := contents.(map[string]interface{})
	if !ok {
		panic("Couldn't cast the contents")
	}

	info := fileDict["info"].(map[string]interface{})

	tf := Torrent{
		Announce: string(fileDict["announce"].([]byte)),
		Info:     *NewTorrentInfoFromBencodedDict(info),
	}

	// Optionals
	// Announce list
	if _, ok := fileDict["announce-list"]; ok {

		announcel := fileDict["announce-list"].([]interface{})
		for _, announce := range announcel {
			a := announce.([]interface{})[0]
			tf.AnnounceList = append(tf.AnnounceList, string(a.([]byte)))
		}
	}

	// Url list
	// if _, ok := fileDict["url-list"]; ok {
	// 	tf.UrlList = string(fileDict["url-list"].([]byte))
	// }

	return tf, nil
}
