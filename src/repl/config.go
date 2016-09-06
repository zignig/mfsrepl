package main

// config object
import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
)

type Remote struct {
	Pin       bool
	Replicate bool
}

type Share struct {
	Path string
}

type Config struct {
	Listen    string
	Nickname  string
	Discovery bool
	Peers     []string
	Shares    map[string]*Share
	PeerID    string
	Password  string
	Remotes   map[string]*Remote
	Channel   string
}

func NewConfig(peer, password string) (c *Config) {
	c = &Config{
		Peers:    make([]string, 0),
		PeerID:   genMac(),
		Remotes:  make(map[string]*Remote),
		Shares:   make(map[string]*Share),
		Listen:   "0.0.0.0:6783",
		Channel:  "default",
		Nickname: mustHostname(),
	}
	if peer != "" {
		c.Peers = append(c.Peers, peer)
	}
	if password != "" {
		c.Password = password
	}
	c.Remotes["bob"] = &Remote{}
	c.Shares["share"] = &Share{Path: "/share"}
	return c
}

func LoadConfig(path, peer, password string) (c *Config) {
	if _, err := toml.DecodeFile(path, &c); err != nil {
		fmt.Println("NO CONFIG, generate empty")
		c = NewConfig(peer, password)
		c.Save(path)
	}
	return
}

func (c *Config) Print() {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(buf.String())
}

func (c *Config) Save(path string) {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c)
	if err != nil {
		fmt.Println(err)
	}
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
	}
	f.Write(buf.Bytes())
}
