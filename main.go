package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"log"
	"os"
	"strings"
	"time"

	"os/exec"

	"github.com/bogem/id3v2/v2"
	"github.com/go-flac/flacpicture"
	"github.com/go-flac/go-flac"
	"golang.org/x/image/draw"

	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
)

const (
	TIMER      = 2
	IMAGE_SIZE = 700
)

func main() {
	for {
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
			img, err = getMP3Art(filepath)
			if err != nil {
				img = defaultArt()
			}
		} else {
			img = defaultArt()
		}
		imgBuff := bytes.NewBuffer(img)

		imgOrig, _, err := image.Decode(imgBuff)
		if err != nil {
			log.Fatal(err.Error())
		}
		imgOut := image.NewRGBA(image.Rect(0, 0, IMAGE_SIZE, IMAGE_SIZE))

		draw.BiLinear.Scale(imgOut, imgOut.Rect, imgOrig, imgOrig.Bounds(), draw.Over, nil)

		jpeg.Encode(imgBuff, imgOut, nil)

		writeTxt("SongAlbum", album)
		writeTxt("SongArtist", artist)
		writeTxt("SongTitle", title)
		writeJpg("AlbumArt", img)

		time.Sleep(TIMER * time.Second)
	}
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

func getMP3Art(s string) ([]byte, error) {
	tags, err := id3v2.Open(s, id3v2.Options{Parse: true})
	if err != nil {
		return nil, errors.New("can't open mp3 tags")
	}
	defer tags.Close()
	pic := tags.GetFrames(tags.CommonID("Attached picture"))
	if pic != nil {
		return pic[0].(id3v2.PictureFrame).Picture, nil //i literally do not care if it has multiple pictures, pick the first one.
	}
	return nil, errors.New("no image found")
}

func defaultArt() []byte {
	file, err := os.Open("default.jpg")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		log.Fatal(err.Error())
	}
	art := make([]byte, info.Size())

	file.Read(art)
	return art
}

func writeJpg(filename string, input []byte) {
	os.Mkdir("output", 0755)
	f, err := os.Create("./output/" + filename + ".jpg")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer f.Close()

	f.Write(input)
}

func writeTxt(filename, input string) {
	os.Mkdir("output", 0755)
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
