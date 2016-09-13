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
		fs.paths[i] = j.Path
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
		//fs.logger.Printf("Check changes %v , %v ", i, j)
		if fs.Stat() {
			stat := fs.Mfs(j)
			//fs.logger.Printf("STAT %v", stat)
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

func (fs *Share) Mfs(path string) (s *Stat) {
	htr, _ := fs.Req("files/stat", "/"+path)
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
	_, err := fs.Req("id", "")
	if err != nil {
		return false
	}
	return true
}

//Req : base request for ipfs access
func (fs *Share) Req(path string, arg string) (resp *http.Response, err error) {
	u := url.URL{}
	u.Scheme = "http"
	u.Host = ipfsHost
	u.Path = api + path
	if arg != "" {
		val := url.Values{}
		val.Set("arg", arg)
		val.Set("encoding", "json")
		u.RawQuery = val.Encode()
	}
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
