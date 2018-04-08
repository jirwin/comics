package comics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/fogleman/gg"
)

//"awkward": {
//     "width": 900,
//     "height": 300,
//     "bubbles": [
//       {
//         "x": 22,
//         "y": 17,
//         "width": 255,
//         "height": 95
//       },
//       {
//         "x": 328,
//         "y": 29,
//         "width": 255,
//         "height": 95
//       }
//     ]
//   },

var imageCache = sync.Map{}

type Bubble struct {
	PosX   float64 `json:"x"`
	PosY   float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type Template struct {
	Name     string    `json:"name"`
	Width    float64   `json:"width"`
	Height   float64   `json:"height"`
	Bubbles  []*Bubble `json:"bubbles"`
	ImageURL string    `json:"image_url"`
}

func (t *Template) String() string {
	out, _ := json.Marshal(t)
	return string(out)
}

func (t *Template) getBaseImg() (image.Image, error) {
	var imageBytes []byte

	cacheItem, ok := imageCache.Load(t.ImageURL)
	if ok {
		imageBytes, ok = cacheItem.([]byte)
		if !ok {
			return nil, fmt.Errorf("error: invalid data in image cache")
		}
	} else {
		resp, err := http.Get(t.ImageURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		imageBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		imageCache.Store(t.ImageURL, imageBytes)
	}
	tempFile, err := ioutil.TempFile("/tmp", "comic-base-img")
	if err != nil {
		return nil, err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}

	return gg.LoadImage(tempFile.Name())
}

func (b *Bubble) setFontSize(dc *gg.Context, text string) {
	fontSize := float64(1)
	renderedHeight := float64(0)

Outer:
	for renderedHeight < b.Height {
		//FIXME: How do we load fonts in a portable way?
		dc.LoadFontFace("/Library/Fonts/Arial.ttf", fontSize)
		wrappedText := dc.WordWrap(text, b.Width)
		for _, t := range wrappedText {
			w, _ := dc.MeasureString(t)
			if w > b.Width {
				break Outer
			}
		}
		renderedHeight = dc.FontHeight() * 1.5 * float64(len(wrappedText))
		fontSize++
	}

	dc.LoadFontFace("/Library/Fonts/Arial.ttf", fontSize)
}

func (t *Template) Render(text []string) ([]byte, error) {
	if len(text) < len(t.Bubbles) {
		return nil, fmt.Errorf("error: not enough text for the template")
	}

	baseImg, err := t.getBaseImg()
	if err != nil {
		return nil, err
	}

	dc := gg.NewContextForImage(baseImg)
	dc.SetRGB(0, 0, 0)

	for i, bubble := range t.Bubbles {
		bubble.setFontSize(dc, text[i])
		dc.DrawStringWrapped(text[i], bubble.PosX, bubble.PosY, 0, 0, bubble.Width, 1.5, gg.AlignLeft)
	}

	buf := bytes.NewBuffer(nil)
	err = dc.EncodePNG(buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func NewTemplate(templateUrl string) (*Template, error) {
	resp, err := http.Get(templateUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	templateBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	t := &Template{}
	err = json.Unmarshal(templateBytes, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}
