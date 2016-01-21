package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	humanize "github.com/dustin/go-humanize"
)

type badgeD struct {
	renderer *template.Template
	urlReq   string

	cache map[string]string
}

const svgTemplate = `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="215" height="20">
	<g shape-rendering="crispEdges">
		<rect x="0" y="0" width="98" height="20" fill="#555"/>
		<rect x="98" y="0" width="117" height="20" fill="#007ec6"/>
	</g>
	<g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
		<text x="50" y="14">{{ .ImageName }}</text>
		<text x="155" y="14">{{ .Size }} / {{ .Layers }}</text>
	</g>
</svg>`

func newBadgeD(apiServer string) *badgeD {
	bd := &badgeD{
		renderer: template.Must(template.New("badge").Parse(svgTemplate)),
		urlReq:   fmt.Sprint(apiServer, "/registry/analyze"),
		cache:    make(map[string]string),
	}
	return bd

}

func (bd badgeD) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	if len(url) <= 1 {
		http.NotFound(w, r)
		return
	}

	if !strings.HasSuffix(url, ".svg") {
		http.NotFound(w, r)
		return
	}

	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(url, "/"), ".svg"), ":")
	if len(parts) != 2 {
		log.Println("error parsing parts for", url)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	image, tag := parts[0], parts[1]

	if str, ok := bd.cache[image+tag]; ok {
		w.Header().Set("Content-type", "image/svg+xml")
		fmt.Fprint(w, str)
		return
	}

	const contentTypeReq = "application/json;charset=UTF-8"
	bodyReq := strings.NewReader(fmt.Sprintf(`{"repos":[{"name":"%s","tag":"%s"}]}`, image, tag))
	resp, err := http.Post(bd.urlReq, contentTypeReq, bodyReq)
	if err != nil {
		log.Println("error posting call:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var data []struct {
		Repo struct {
			Name   string `json:"name"`
			Tag    string `json:"tag"`
			Size   uint64 `json:"size"`
			Count  int    `json:"count"`
			Status int    `json:"status"`
		} `json:"repo"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Println("error parsing JSON response:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(data) == 0 {
		log.Println("can't find image")
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-type", "image/svg+xml")
	var buf bytes.Buffer
	bd.renderer.Execute(&buf, struct {
		ImageName string
		Size      string
		Layers    string
	}{image + ":" + tag, humanize.Bytes(data[0].Repo.Size), fmt.Sprint(data[0].Repo.Count, " layers")})

	bd.cache[image+tag] = buf.String()
	fmt.Fprint(w, buf.String())
}

func main() {
	apiServer := os.Getenv("IMAGE_LAYERS_API")
	if apiServer == "" {
		log.Fatal("please, define environment variable IMAGE_LAYERS_API")
	}

	http.Handle("/", newBadgeD(apiServer))
	http.ListenAndServe(":8080", nil)
}
