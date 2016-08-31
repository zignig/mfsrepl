package mfs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"net/http"
	"net/url"
)

const (
	api      = "/api/v0/"
	ipfsHost = "localhost:5001"
)

//IPfsfs : file system ROfs interface
type IPfsfs struct {
	watch map[string]string
}

func init() {
	http.DefaultClient.Transport = &http.Transport{DisableKeepAlives: true}
}

func NewIPfsfs() (fs *IPfsfs) {
	fs = &IPfsfs{}
	fs.watch = make(map[string]string)
	return fs
}

type Stat struct {
	Hash           string
	Size           int
	CumulativeSize int
	Blocks         int
	Type           string
}

func (fs *IPfsfs) Mfs(path string) (s *Stat) {
	htr, _ := fs.Req("files/stat", "/"+path)
	data, _ := ioutil.ReadAll(htr.Body)
	merr := json.Unmarshal(data, &s)
	if merr != nil {
		fmt.Println("FAIL", merr)
	}
	fmt.Println(s)
	return s
}

//Stat : Check if the file system exist
func (fs *IPfsfs) Stat() (stat bool) {
	_, err := fs.Req("id", "")
	if err != nil {
		return false
	}
	return true
}

//Req : base request for ipfs access
func (fs *IPfsfs) Req(path string, arg string) (resp *http.Response, err error) {
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
