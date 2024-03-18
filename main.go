package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"os/exec"

	"github.com/bogem/id3v2/v2"
	"github.com/go-flac/flacpicture"
	"github.com/go-flac/go-flac"
	"golang.org/x/image/draw"

	_ "image/gif"
	"image/jpeg"
	_ "image/png"
)

const (
	TIMER      = 2
	IMAGE_SIZE = 700
)

func main() {
	prevFilepath := ""
	for {
		c := exec.Command("cmus-remote", "-Q")
		o, _ := c.Output()
		remoteResp := strings.Split(string(o), "\n")

		path, err := getAttribute(remoteResp, "file ")
		if err != nil {
			log.Fatalln(err.Error())
		}

		if path != prevFilepath {
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

			// Image
			img := make([]byte, 0)

			switch {
			case hasFlacTrackArt(path):
				img, err = getFlacArt(path)
				if err != nil {
					img, err = getCoverJpg(path)
					if err != nil {
						log.Println(err.Error())
						img = getDefaultArt()
					}
				}
			case hasMp3TrackArt(path):
				img, err = getMP3Art(path)
				if err != nil {
					img, err = getCoverJpg(path)
					if err != nil {
						log.Println(err.Error())
						img = getDefaultArt()
					}
				}
			case hasCoverJpg(path):
				img, err = getCoverJpg(path)
				if err != nil {
					log.Println(err.Error())
					img = getDefaultArt()
				}
			default:
				img = getDefaultArt()
			}

			imgBuff := bytes.NewBuffer(img)

			imgOrig, _, err := image.Decode(imgBuff)
			if err != nil {
				log.Fatalln(err.Error())
			}
			imgOut := image.NewRGBA(image.Rect(0, 0, IMAGE_SIZE, IMAGE_SIZE))

			draw.BiLinear.Scale(imgOut, imgOut.Rect, imgOrig, imgOrig.Bounds(), draw.Over, nil)

			err = jpeg.Encode(imgBuff, imgOut, nil)
			if err != nil {
				log.Fatalln(err.Error())
			}

			writeTxt("SongAlbum", album)
			writeTxt("SongArtist", artist)
			writeTxt("SongTitle", title)
			writeJpg("AlbumArt", img)

		}
		prevFilepath = path
		time.Sleep(TIMER * time.Second)
	}
}

// has

func hasFlacTrackArt(s string) bool {
	// dumb check
	if !strings.HasSuffix(s, ".flac") {
		return false
	}

	//  is parsable
	f, err := flac.ParseFile(s)
	if err != nil {
		return false
	}

	// has any frames
	if len(f.Meta) == 0 {
		return false
	}

	// has any pictures
	for _, metadata := range f.Meta {
		if metadata.Type == flac.Picture {
			return true
		}
	}

	// no pictures
	return false
}

func hasMp3TrackArt(s string) bool {
	// dumb check
	if !strings.HasSuffix(s, ".mp3") {
		return false
	}

	// is parsable
	m, err := id3v2.Open(s, id3v2.Options{Parse: true})
	if err != nil {
		return false
	}
	defer m.Close()

	// has a picture
	pic := m.GetFrames(m.CommonID("Attached picture"))
	if pic != nil {
		return true
	}

	// no pictures
	return false
}

func hasCoverJpg(s string) bool {
	exts := []string{".jpg", ".jpeg", ".png", ".gif"} // 'hasCoverJpg' is a misnomer, check all sane image types
	dir := filepath.Dir(s)
	file, err := os.Open(dir)
	if err != nil {
		log.Printf("can't open %v; %v\n", dir, err.Error())
		return false
	}
	defer file.Close()
	indexes, _ := file.ReadDir(0)
	for i := range indexes {
		for _, key := range exts {
			if "cover"+key == indexes[i].Name() {
				return true
			}
		}
	}

	return false
}

// get

func getFlacArt(s string) ([]byte, error) {
	f, err := flac.ParseFile(s)
	if err != nil {
		return nil, errors.New("can't open file")
	}
	for _, metadata := range f.Meta {
		if metadata.Type == flac.Picture {
			// do not care if it has multiple pictures, pick the first one.
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
		// do not care if it has multiple pictures, pick the first one.
		return pic[0].(id3v2.PictureFrame).Picture, nil
	}
	return nil, errors.New("no image found")
}

func getCoverJpg(s string) ([]byte, error) {
	exts := []string{".jpg", ".jpeg", ".png", ".gif"}
	dir := filepath.Dir(s)
	file, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("can't open %v; %v", dir, err.Error())
	}
	defer file.Close()
	indexes, _ := file.ReadDir(0)
	for i := range indexes {
		for _, key := range exts {
			if "cover"+key == indexes[i].Name() {
				cover, err := os.Open(dir + "/" + indexes[i].Name())
				if err != nil {
					return nil, fmt.Errorf("can't open %v", indexes[i].Name())
				}
				defer cover.Close()
				info, err := indexes[i].Info()
				if err != nil {
					return nil, fmt.Errorf("cant stat %v; %v", indexes[i].Name(), err.Error())
				}
				art := make([]byte, info.Size())

				cover.Read(art)
				return art, nil
			}
		}
	}
	return nil, errors.New("cover not found")
}

func getDefaultArt() []byte {
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

// write

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
