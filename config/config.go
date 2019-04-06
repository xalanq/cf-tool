package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
}

// New an empty config
func New() *Config {
	return &Config{
		Username: "",
		Password: "",
		Template: []CodeTemplate{},
	}
}

func createHash(key string) []byte {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hasher.Sum(nil)
}

func encrypt(username, password string) string {
	block, _ := aes.NewCipher(createHash("glhf" + username + "233"))
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	text := gcm.Seal(nonce, nonce, []byte(password), nil)
	return hex.EncodeToString(text)
}

func decrypt(username, password string) string {
	data, err := hex.DecodeString(password)
	if err != nil {
		log.Fatal("Cannot decode password")
	}
	block, _ := aes.NewCipher(createHash("glhf" + username + "233"))
	gcm, _ := cipher.NewGCM(block)
	nonceSize := gcm.NonceSize()
	nonce, text := data[:nonceSize], data[nonceSize:]
	plain, err := gcm.Open(nil, nonce, text, nil)
	if err != nil {
		panic(err.Error())
	}
	return string(plain)
}

// DecryptPassword get real password
func (c *Config) DecryptPassword() string {
	return decrypt(c.Username, c.Password)
}

// Load from path
func Load(path string) *Config {
	file, err := os.Open(path)
	if err != nil {
		return New()
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)

	c := &Config{}
	json.Unmarshal(bytes, c)

	return c
}

// Save file to path
func (c *Config) Save(path string) {
	data, _ := json.MarshalIndent(c, "", "  ")
	err := ioutil.WriteFile(path, data, 0644)
	if err != nil {
		fmt.Printf("Cannot write config to %v\n%v", path, err.Error())
	}
}

// Login configurate
func (c *Config) Login() {
	fmt.Println("Config username(email) and password(encrypt)")
	fmt.Printf("username: ")
	var username string
	fmt.Scanln(&username)
	fmt.Printf("password: ")
	bytes, _ := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytes)
	fmt.Println()

	err := client.New().Login(username, password)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Invalid username and password")
	} else {
		c.Username = username
		c.Password = encrypt(username, password)
		fmt.Println("Succeed!")
	}
}

// Add template
func (c *Config) Add() {
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
		if err != nil {
			fmt.Printf("%v is invalid. Please input again: ", path)
		} else if _, err := os.Stat(path); err != nil {
			fmt.Printf("%v is invalid. Please input again: ", path)
		} else {
			break
		}
	}

	fmt.Print("Match suffix(e.g. cpp cxx): ")
	var sf string
	fmt.Scanln(&sf)
	suffix := strings.Fields(sf)

	c.Template = append(c.Template, CodeTemplate{lang, path, suffix})
	return
}
