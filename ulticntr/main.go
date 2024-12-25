package main

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/http/cgi"
	"net/url"
	"os"
	"time"

	counter "github.com/funcimp/cgibinftw/ulticntr/counter"
)

var (
	//go:embed assets/html.go.html
	htmlTemplate string
	//go:embed assets/images/starbg.gif
	starBG []byte
	//go:embed assets/images/fireworksbg.gif
	fireworksBG []byte
	//go:embed assets/images/alien.gif
	alien []byte
	//go:embed assets/images/disk.gif
	disk []byte
	//go:embed assets/images/spidey.gif
	spidey []byte
	//go:embed assets/images/yinyang.gif
	yinyang []byte
	//go:embed assets/images/peace.gif
	peace []byte
)

func main() {
	http.HandleFunc("/", render)
	if err := cgi.Serve(nil); err != nil {
		log.Fatal(err)
	}
}

type img struct {
	data   []byte
	Class  string
	Width  uint
	Height uint
}

func (i img) URI() template.URL {
	uri := fmt.Sprintf("data:image/gif;base64,%v", base64.StdEncoding.EncodeToString(i.data))
	return template.URL(uri)
}

func (a assets) GetClassName(i int) string {
	return a.Images[a.Arrangement[i]].Class
}

type assets struct {
	Background  img
	Images      []img
	Arrangement []uint
	Counter     uint64
	Diag        bool
}

type option func(*assets)

func withFireworks() option {
	return func(a *assets) {
		a.Background = img{data: fireworksBG}
	}
}

func withDiag() option {
	return func(a *assets) { a.Diag = true }
}

func withRandomImages() option {
	return func(a *assets) {
		rand.Seed(time.Now().UnixNano())
		min := 1
		max := 4
		arr := []uint{0}
		for i := 1; i <= 12; i++ {
			arr = append(arr, uint(rand.Intn(max-min+1)+min))
		}
		a.Arrangement = arr
	}
}

func newAssets(options ...option) assets {
	images := []img{
		{data: spidey, Class: "spidey", Width: 241, Height: 124},
		{data: alien, Class: "alien", Width: 82, Height: 90},
		{data: disk, Class: "disk", Width: 65, Height: 80},
		{data: peace, Class: "peace", Width: 74, Height: 75},
		{data: yinyang, Class: "yinyang", Width: 65, Height: 65},
	}
	arr := []uint{0, 1, 3, 2, 1, 4, 1, 1, 4, 1, 2, 3, 1}
	a := assets{
		Images:      images,
		Arrangement: arr,
		Background:  img{data: starBG},
	}
	for _, opt := range options {
		opt(&a)
	}
	return a
}

func parseOpts(queryString string) (o []option) {
	values, _ := url.ParseQuery(queryString)
	if values.Get("random") == "true" {
		o = append(o, withRandomImages())
	}
	if values.Get("bg") == "fireworks" {
		o = append(o, withFireworks())
	}
	if values.Get("diag") == "true" {
		o = append(o, withDiag())
	}
	return o
}

func render(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("page").Parse(htmlTemplate))
	a := newAssets(parseOpts(os.Getenv("QUERY_STRING"))...)

	c, err := counter.New()
	if err != nil {
		log.Println("oops:", err)
		return
	}

	i, err := c.Count()
	if err != nil {
		log.Println("oops:", err)
		return
	}
	a.Counter = i
	if err := t.Execute(w, a); err != nil {
		log.Println("oops:", err)
		return
	}
}
