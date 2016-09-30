package mfs

import (
	"encoding/json"
	"errors"
	"github.com/op/go-logging"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	api      = "/api/v0/"
	ipfsHost = "localhost:5001"
)

var logger = logging.MustGetLogger("mfs")

type Update struct {
	Path     string
	PeerName string
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
	Updates chan Update
	lock    sync.Mutex
}

func init() {
	http.DefaultClient.Transport = &http.Transport{DisableKeepAlives: true}
}

func NewShare(bind map[string]*Share) (fs *Share) {
	fs = &Share{}
	fs.watch = make(map[string]string)
	fs.paths = make(map[string]string)
	fs.Updates = make(chan Update, 50)
	for i, j := range bind {
		fs.paths[i] = j.Source
		logger.Debugf("%v", fs.paths)
		fs.watch[i] = ""
		fs.Mkdir("/"+i, true)
	}
	return fs
}

type Stat struct {
	Hash           string
	Size           int
	CumulativeSize int
	Blocks         int
	Type           string
	// in case of error
	Message string
	Code    int
}

func (fs *Share) UpdateChannel() (c chan Update) {
	return fs.Updates
}

func (fs *Share) Watch(interval int) {
	c := time.Tick(time.Duration(interval) * time.Second)
	for {
		select {
		case <-c:
			logger.Debug("WATCH")
			fs.CheckChanges()
		}
	}
}

func (fs *Share) CheckChanges() {
	if fs.Stat() {
		for i, j := range fs.paths {
			logger.Debugf("Check changes %v , %v ", i, j)
			stat, err := fs.Mfs(j)
			if err != nil {
				logger.Errorf("file does not exists , %v", err)
				continue
			}
			logger.Debugf("STAT %v", stat)
			if fs.watch[i] != stat.Hash {
				update := Update{
					Path:    i,
					OldHash: fs.watch[i],
					NewHash: stat.Hash,
					Stamp:   time.Now(),
				}
				fs.Updates <- update
				fs.watch[i] = stat.Hash
				logger.Info("HASH has changed! %v", update)
			}
		}
	}
}

func (fs *Share) SubmitUpdate(u Update) (err error) {
	if fs.Stat() {
		fs.lock.Lock()
		defer func() {
			logger.Infof("UNLOCK")
			fs.lock.Unlock()
		}()
		logger.Infof("LOCK")
		logger.Infof("%v", u)
		// do we have this share
		_, ok := fs.watch[u.Path]
		if ok {
			// Make the target backup
			backupPath := fs.StampBackup()
			sourcePath := "/" + u.Path + "/" + u.PeerName
			fs.Mkdir(backupPath+"/"+u.Path, true)
			err = fs.Move(sourcePath, backupPath+sourcePath)
			if err != nil {
				logger.Errorf("Move %v", err)
				fs.Mkdir(sourcePath, true)
				return
			}
			err = fs.CopyHash(u.NewHash, sourcePath)
			if err != nil {
				logger.Errorf("Copy %v", err)
				return
			}
		}
	}
	return err
}

func (fs *Share) StampBackup() string {
	const layout = "/2006/01/02/15/04/"
	n := time.Now()
	dateBack := n.Format(layout)
	logger.Info("Backup ", dateBack)
	fs.Mkdir(dateBack, true)
	return dateBack
}

func (fs *Share) Move(source, destination string) (err error) {
	val := url.Values{}
	val.Set("arg", source)
	val.Add("arg", destination)
	_, err = fs.Request("files/mv", val)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (fs *Share) CopyHash(source, destination string) (err error) {
	return fs.Copy("/ipfs/"+source, destination)
}

func (fs *Share) Copy(source, destination string) (err error) {
	val := url.Values{}
	val.Set("arg", source)
	val.Add("arg", destination)
	_, err = fs.Request("files/cp", val)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (fs *Share) Mkdir(path string, parents bool) (err error) {
	val := url.Values{}
	val.Set("arg", path)
	if parents {
		val.Set("p", "true")
	}
	_, err = fs.Request("files/mkdir", val)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (fs *Share) Mfs(path string) (s *Stat, err error) {
	val := url.Values{}
	val.Set("arg", path)
	htr, err := fs.Request("files/stat", val)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer htr.Body.Close()
	data, _ := ioutil.ReadAll(htr.Body)
	merr := json.Unmarshal(data, &s)
	if merr != nil {
		logger.Error(err)
		return nil, err
	}
	return s, err
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
	logger.Debugf("url request -> %s", u.String())
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
