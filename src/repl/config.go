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

type Config struct {
	Peers    []string
	PeerID   string
	Password string
	Remotes  map[string]*Remote
}

func NewConfig() (c *Config) {
	c = &Config{
		Peers:   make([]string, 0),
		PeerID:  genMac(),
		Remotes: make(map[string]*Remote),
	}
	c.Remotes["bob"] = &Remote{}
	return c
}

func LoadConfig(path string) (c *Config) {
	if _, err := toml.DecodeFile(path, &c); err != nil {
		fmt.Println("NO CONFIG, generate empty")
		c = NewConfig()
		c.Save(path)
	}
	return
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
