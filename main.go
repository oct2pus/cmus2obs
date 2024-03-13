package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"os/exec"
)

const TIMER = 2

func main() {
	c := exec.Command("cmus-remote", "-Q")
	o, _ := c.Output()
	remoteResp := strings.Split(string(o), "\n")
	//Album
	album, err := getAttribute(remoteResp, "tag album ")
	if err != nil {
		album = "Unknown"
	}
	//artist
	artist, err := getAttribute(remoteResp, "tag artist ")
	if err != nil {
		artist = "Unknown"
	}
	// Title
	title, err := getAttribute(remoteResp, "tag title ")
	if err != nil {
		title = "Unknown"
	}

	writeTxt("SongAlbum", album)
	writeTxt("SongArtist", artist)
	writeTxt("SongTitle", title)
}

func writeTxt(filename, input string) {
	f, err := os.Create("./output/" + filename + ".txt")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()

	f.Write([]byte(input))
}

func getAttribute(input []string, prefix string) (string, error) {
	var attr string
	for i := range input {
		has := strings.HasPrefix(input[i], prefix)
		if has {
			attr = input[i]
		}
	}
	attr, b := strings.CutPrefix(attr, prefix)
	if !b {
		return "", fmt.Errorf("did not find prefix \"%v\"", prefix)
	}
	return attr, nil
}
