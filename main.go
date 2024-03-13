package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"os/exec"

	"github.com/go-flac/flacpicture"
	"github.com/go-flac/go-flac"
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

	filepath, err := getAttribute(remoteResp, "file ")
	if err != nil {
		log.Fatal(err.Error())
	}
	img := make([]byte, 0)
	if strings.HasSuffix(filepath, ".flac") {
		img, err = getFlacArt(filepath)
		if err != nil {
			img = defaultArt()
		}
	} else if strings.HasSuffix(filepath, ".mp3") {
		getMP3Art()
	} else {
		defaultArt()
	}

	writeTxt("SongAlbum", album)
	writeTxt("SongArtist", artist)
	writeTxt("SongTitle", title)
	writeJpg("AlbumArt", img)
}

func getFlacArt(s string) ([]byte, error) {
	f, err := flac.ParseFile(s)
	if err != nil {
		return nil, errors.New("can't open file")
	}
	for _, metadata := range f.Meta {
		if metadata.Type == flac.Picture {
			pic, err := flacpicture.ParseFromMetaDataBlock(*metadata)
			return pic.ImageData, err
		}
	}
	return nil, errors.New("no image found")
}

func getMP3Art() {

}

func defaultArt() []byte {
	return nil
}

func writeJpg(filename string, input []byte) {
	f, err := os.Create("./output/" + filename + ".jpg")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()

	f.Write(input)
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
