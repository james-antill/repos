package repos

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type Repodata struct {
	Baseurl  string
	Revision int
	Primary  Data
	Files    Data
	GrpRAW   Data
	GrpGZ    Data
	Other    Data
	ModMD    Data
}

func (snap *Snapshot) RepoMD() (*Repodata, error) {
	var xmlData struct {
		Revision int `xml:"revision"`
		Data     []struct {
			T        string `xml:"type,attr"`
			Checksum struct {
				T string `xml:"type,attr"`
				D string `xml:",chardata"`
			} `xml:"checksum"`
			Location struct {
				Href string `xml:"href,attr"`
			} `xml:"location"`
			Timestamp float64 `xml:"timestamp"`
			Size      int     `xml:"size"`
		} `xml:"data"`
	}

	var err error
	var repomd []byte
	var baseurl string

	for i := range snap.URLs {
		repomd, err = url2bytes(snap.URLs[i].URL)
		if err != nil {
			//			fmt.Printf("error: %v", err)
			//			return nil, err
			repomd = nil
			continue
		}

		if !hchks(repomd, snap.Repomd.Chks) {
			err = fmt.Errorf("error: Checksum doesn't match for repomd")
			repomd = nil
			//			return nil, err
			continue
		}

		// FIXME: Pass all urls down
		baseurl = strings.TrimSuffix(snap.URLs[i].URL, "repodata/repomd.xml")

		// fmt.Println(string(repomd))
		break
	}
	if len(repomd) == 0 {
		return nil, err
	}

	err = xml.Unmarshal(repomd, &xmlData)
	if err != nil {
		// fmt.Printf("error: %v", err)
		return nil, err
	}

	ret := &Repodata{Baseurl: baseurl}
	ret.Revision = xmlData.Revision

	for i := range xmlData.Data {
		v := &xmlData.Data[i]
		var d *Data
		switch v.T {
		case "primary":
			d = &ret.Primary
		case "files":
			d = &ret.Files
		case "other":
			d = &ret.Other
		case "group":
			d = &ret.GrpRAW
		case "group_gz":
			d = &ret.GrpGZ
		case "modules":
			d = &ret.ModMD
		case "prestodelta":
			fallthrough
		case "updateinfo":
			fallthrough
		default:
			continue
		}
		d.Path = v.Location.Href
		d.Size = v.Size
		d.TM = time.Unix(int64(v.Timestamp), 0)
		d.Chks = []Checksum{{Kind: v.Checksum.T, Data: v.Checksum.D}}
	}
	return ret, err
}
