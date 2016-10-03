package main

// config object
import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/op/go-logging"
	"mfs"
	"os"
)

var confLogger = logging.MustGetLogger("config")

type Remote struct {
	Pin       bool
	Replicate bool
}

type Config struct {
	Listen    string
	Nickname  string
	Discovery bool
	Peers     []string
	Shares    map[string]*mfs.Share
	PeerID    string
	Password  string
	Remotes   map[string]*Remote
	Channel   string
}

func NewConfig(peer, password, nickname string) (c *Config) {
	c = &Config{
		Peers:    make([]string, 0),
		PeerID:   genMac(),
		Remotes:  make(map[string]*Remote),
		Shares:   make(map[string]*mfs.Share),
		Listen:   "0.0.0.0:6783",
		Channel:  "share",
		Nickname: mustHostname(),
	}
	if peer != "" {
		c.Peers = append(c.Peers, peer)
	}
	if password != "" {
		c.Password = password
	}
	if nickname != "" {
		c.Nickname = nickname
	}
	c.Remotes["bob"] = &Remote{}
	c.Shares["share"] = &mfs.Share{Path: "/share"}
	return c
}

func LoadConfig(path, peer, password, nickname string) (c *Config) {
	if _, err := toml.DecodeFile(path, &c); err != nil {
		fmt.Println(c, err)
		c = NewConfig(peer, password, nickname)
		fmt.Println(c)
		c.Save(path)
	}
	return
}

func (c *Config) Print() {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c)
	if err != nil {
		logger.Criticalf("%v", err)
	}
	fmt.Println(buf.String())
}

func (c *Config) Save(path string) {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c)
	if err != nil {
		confLogger.Errorf("%v", err)
	}
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		confLogger.Errorf("%v", err)
	}
	f.Write(buf.Bytes())
}
