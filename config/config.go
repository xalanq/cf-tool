package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"syscall"

	homedir "github.com/mitchellh/go-homedir"

	"github.com/xalanq/codeforces/client"
	"golang.org/x/crypto/ssh/terminal"
)

// CodeTemplate config parse code template
type CodeTemplate struct {
	Lang   string   `json:"lang"`
	Path   string   `json:"path"`
	Suffix []string `json:"suffix"`
}

// Config load and save configuration
type Config struct {
	Username string         `json:"username"`
	Password string         `json:"password"`
	Template []CodeTemplate `json:"template"`
	path     string
}

// New an empty config
func New(path string) *Config {
	c := &Config{path: path}
	if err := c.load(); err != nil {
		return nil
	}
	return c
}

// load from path
func (c *Config) load() (err error) {
	file, err := os.Open(c.path)
	if err != nil {
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, c)
}

// save file to path
func (c *Config) save() (err error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err == nil {
		return
	}
	err = ioutil.WriteFile(c.path, data, 0644)
	return
}

func createHash(key string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hasher.Sum(nil)
}

func encrypt(username, password string) (ret string, err error) {
	block, err := aes.NewCipher(createHash("glhf" + username + "233"))
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}
	text := gcm.Seal(nonce, nonce, []byte(password), nil)
	ret = hex.EncodeToString(text)
	return
}

func decrypt(username, password string) (ret string, err error) {
	data, err := hex.DecodeString(password)
	if err != nil {
		err = errors.New("Cannot decode password")
		return
	}
	block, err := aes.NewCipher(createHash("glhf" + username + "233"))
	if err != nil {
		return
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	nonceSize := gcm.NonceSize()
	nonce, text := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, text, nil)
	if err != nil {
		return
	}
	ret = string(plain)
	return
}

// DecryptPassword get real password
func (c *Config) DecryptPassword() (string, error) {
	return decrypt(c.Username, c.Password)
}

// Login configurate
func (c *Config) Login(path string) (err error) {
	fmt.Println("Config username(email) and password(encrypt)")
	fmt.Printf("username: ")
	var username string
	fmt.Scanln(&username)
	fmt.Printf("password: ")
	bytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	password := string(bytes)
	fmt.Println()

	err = client.New(path).Login(username, password)
	if err != nil {
		return errors.New("Invalid username and password")
	}
	password, err = encrypt(username, password)
	if err == nil {
		c.Username = username
		c.Password = password
		fmt.Println("Succeed!")
		err = c.save()
	}
	return
}

// Add template
func (c *Config) Add() (err error) {
	fmt.Println("Language list:")
	type kv struct {
		K, V string
	}
	langs := []kv{}
	for k, v := range client.Langs {
		langs = append(langs, kv{k, v})
	}
	sort.Slice(langs, func(i, j int) bool { return langs[i].V < langs[j].V })
	for _, t := range langs {
		fmt.Printf("%5v: %v\n", t.K, t.V)
	}
	fmt.Print("Select a language(e.g. 42): ")
	var lang string
	fmt.Scanln(&lang)

	fmt.Print(`Template absolute path(e.g. ~/template/io.cpp): `)
	var path string
	for {
		fmt.Scanln(&path)
		path, err := homedir.Expand(path)
		if err == nil {
			if _, err := os.Stat(path); err == nil {
				break
			}
		}
		fmt.Printf("%v is invalid. Please input again: ", path)
	}

	fmt.Print("Match suffix(e.g. cpp cxx): ")
	var sf string
	fmt.Scanln(&sf)
	suffix := strings.Fields(sf)

	c.Template = append(c.Template, CodeTemplate{lang, path, suffix})
	return c.save()
}
