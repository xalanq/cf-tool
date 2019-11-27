package util

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// CHA map
const CHA = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandString n is the length. a-z 0-9
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = CHA[rand.Intn(len(CHA))]
	}
	return string(b)
}

// Scanline scan line
func Scanline() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text()
	}
	fmt.Println("\nInterrupted.")
	os.Exit(1)
	return ""
}

// ScanlineTrim scan line and trim
func ScanlineTrim() string {
	return strings.TrimSpace(Scanline())
}

// ChooseIndex return valid index in [0, maxLen)
func ChooseIndex(maxLen int) int {
	color.Cyan("Please choose one (index): ")
	for {
		index := ScanlineTrim()
		i, err := strconv.Atoi(index)
		if err == nil && i >= 0 && i < maxLen {
			return i
		}
		color.Red("Invalid index! Please try again: ")
	}
}

// YesOrNo must choose one
func YesOrNo(note string) bool {
	color.Cyan(note)
	for {
		tmp := ScanlineTrim()
		if tmp == "y" || tmp == "Y" {
			return true
		}
		if tmp == "n" || tmp == "N" {
			return false
		}
		color.Red("Invalid input. Please input again: ")
	}
}

// DebugSave write data to temperory file
func DebugSave(data interface{}) {
	f, err := os.OpenFile("./tmp/body", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if data, ok := data.([]byte); ok {
		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}
	} else {
		if _, err := f.Write([]byte(fmt.Sprintf("%v\n\n", data))); err != nil {
			log.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

// Returns true if a given string is an url
func IsUrl(str string) bool {
	if _, err := url.ParseRequestURI(str); err == nil {
		return true
	}
	return false
}
