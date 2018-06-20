package repos

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func url2bytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("non-200 status (%s): %s", url, resp.Status)
		return nil, err
	}

	bbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bbody, nil
}

type URL struct {
	URL string
	Pri int
}

type Checksum struct {
	Kind string
	Data string
}

func (chk Checksum) String() string {
	return fmt.Sprintf("%s:%s", chk.Kind, chk.Data)
}

type Data struct {
	Path string
	Chks []Checksum
	Size int
	TM   time.Time
}

type Snapshot struct {
	URLs   []URL
	Repomd Data
}

func Metalink(url string) (*Snapshot, error) {
	var xmlData struct {
		Timestamp    int64 `xml:"files>file>timestamp"`
		Size         int   `xml:"files>file>size"`
		Verification []struct {
			T string `xml:"type,attr"`
			D string `xml:",chardata"`
		} `xml:"files>file>verification>hash"`
		Maxconnections int `xml:"maxconnections,attr"`
		URLs           []struct {
			URL        string `xml:",chardata"`
			Preference int    `xml:"preference,attr"`
			Protocol   string `xml:"protocol,attr"`
		} `xml:"files>file>resources>url"`
	}

	metalink, err := url2bytes(url)
	if err != nil {
		// fmt.Printf("error: %v", err)
		return nil, err
	}

	// fmt.Println(string(metalink))

	err = xml.Unmarshal(metalink, &xmlData)
	if err != nil {
		// fmt.Printf("error: %v", err)
		return nil, err
	}
	if len(xmlData.URLs) < 1 {
		err = fmt.Errorf("error: No data for metalink")
		fmt.Println(string(metalink))
		return nil, err
	}

	ret := &Snapshot{}

	ret.Repomd.Path = "repodata/repomd.xml"
	ret.Repomd.Size = xmlData.Size
	ret.Repomd.TM = time.Unix(xmlData.Timestamp, 0)
	for i := range xmlData.Verification {
		v := &xmlData.Verification[i]
		ret.Repomd.Chks = append(ret.Repomd.Chks, Checksum{Kind: v.T, Data: v.D})
	}
	for i := range xmlData.URLs {
		v := &xmlData.URLs[i]
		if v.Protocol != "http" {
			continue
		}
		ret.URLs = append(ret.URLs, URL{URL: v.URL, Pri: v.Preference})
	}

	return ret, nil
}

func Baseurl(url string) (*Snapshot, error) {
	ret := &Snapshot{}
	path := "repodata/repomd.xml"
	ret.Repomd.Path = path
	ret.URLs = append(ret.URLs, URL{URL: url + path, Pri: 1})

	return ret, nil
}
