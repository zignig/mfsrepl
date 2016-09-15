package mfs

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	api      = "/api/v0/"
	ipfsHost = "localhost:5001"
)

type Update struct {
	PeerName string
	Path     string
	NewHash  string
	OldHash  string
	Stamp    time.Time
}

//Share : file system ROfs interface
type Share struct {
	Path    string
	Source  string
	watch   map[string]string
	paths   map[string]string
	updates chan Update
	logger  *logging.Logger
}

func init() {
	http.DefaultClient.Transport = &http.Transport{DisableKeepAlives: true}
}

func NewShare(bind map[string]*Share, logger *logging.Logger) (fs *Share) {
	fs = &Share{}
	fs.watch = make(map[string]string)
	fs.paths = make(map[string]string)
	fs.updates = make(chan Update, 50)
	fs.logger = logger
	for i, j := range bind {
		fs.paths[i] = j.Source
		fs.logger.Debugf("%v", fs.paths)
		fs.watch[i] = ""
	}
	return fs
}

type Stat struct {
	Hash           string
	Size           int
	CumulativeSize int
	Blocks         int
	Type           string
}

func (fs *Share) UpdateChannel() (c chan Update) {
	return fs.updates
}

func (fs *Share) Watch(interval int) {
	c := time.Tick(time.Duration(interval) * time.Second)
	for {
		select {
		case <-c:
			fs.logger.Debug("WATCH")
			fs.CheckChanges()
		}
	}
}

func (fs *Share) CheckChanges() {
	for i, j := range fs.paths {
		fs.logger.Debugf("Check changes %v , %v ", i, j)
		if fs.Stat() {
			stat := fs.Mfs(j)
			fs.logger.Debugf("STAT %v", stat)
			if fs.watch[i] != stat.Hash {
				update := Update{
					Path:    i,
					OldHash: fs.watch[i],
					NewHash: stat.Hash,
					Stamp:   time.Now(),
				}
				fs.updates <- update
				fs.watch[i] = stat.Hash
			}

		}
	}
}

func (fs *Share) Backup(path string) {
	const layout = "/2006/01/02/15/04/"
	n := time.Now()
	dateBack := n.Format(layout)
	fs.Mkdir(dateBack, true)
}

func (fs *Share) Mkdir(path string, parents bool) (err error) {
	val := url.Values{}
	val.Set("arg", path)
	if parents {
		val.Set("p", "true")
	}
	_, err = fs.Request("files/mkdir", val)
	if err != nil {
		fs.logger.Error(err)
		return err
	}
	return nil
}

func (fs *Share) Mfs(path string) (s *Stat) {
	val := url.Values{}
	val.Set("arg", path)
	htr, _ := fs.Request("files/stat", val)
	data, _ := ioutil.ReadAll(htr.Body)
	merr := json.Unmarshal(data, &s)
	if merr != nil {
		fmt.Println("FAIL", merr)
	}
	//fmt.Println(s)
	return s
}

//Stat : Check if the file system exist
func (fs *Share) Stat() (stat bool) {
	_, err := fs.Request("id", nil)
	if err != nil {
		return false
	}
	return true
}

func (fs *Share) Request(path string, val url.Values) (resp *http.Response, err error) {
	u := url.URL{}
	u.Scheme = "http"
	u.Host = ipfsHost
	u.Path = api + path
	if val == nil {
		val = url.Values{}
	}
	val.Set("encoding", "json")
	u.RawQuery = val.Encode()
	fs.logger.Debugf("url request -> %s", u.String())
	resp, err = http.Get(u.String())
	if resp == nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return resp, errors.New(resp.Status)
	}
	if err != nil {
		return resp, err
	}
	return resp, err
}
