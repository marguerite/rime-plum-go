package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cavaliercoder/grab"
)

func (r Recipe) download(pkg Package) {
	if len(r.DownloadFiles) == 0 {
		return
	}
	downloadFiles(r.DownloadFiles, pkg.WorkingDirectory)
}

func downloadFiles(str, dir string) {
	arr := splitFiles(str)
	for _, url := range arr {
		idx := strings.Index(url, "::")
		var filename string
		if idx < 0 {
			filename = filepath.Join(dir, filepath.Base(url))
		} else {
			filename = filepath.Join(dir, url[:idx])
			url = url[idx+2:]
		}

		resp, err := CLIENT.Head(url)
		if err != nil {
			panic(err)
		}

		lastModified, _ := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))

		if stat, err := os.Stat(filename); os.IsNotExist(err) {
			fetch(url, filename)
			// need to git commit once, or the later git pull will destroy our file
			commit(dir, filename)
		} else {
			// the update case
			fmt.Printf("fuck1%s\n", stat.ModTime().String())
			if lastModified.After(stat.ModTime()) {
				fetch(url, filename)
				commit(dir, filename)
			} else {
				fmt.Printf("%s is already up-to-date\n", filename)
			}
		}
	}
}

func fetch(src, dst string) {
	fmt.Printf("Downloading %s to %s\n", src, dst)

	client := grab.NewClient()
	req, _ := grab.NewRequest(dst, src)

	fmt.Printf("Initializing download...\n")
	resp := client.Do(req)
	if resp.HTTPResponse.StatusCode != http.StatusOK {
		fmt.Printf("failed to get %s, code %d", src, resp.HTTPResponse.StatusCode)
		os.Exit(1)
	}

	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	if err := resp.Err(); err != nil {
		fmt.Printf("Error downloading %s: %v\n", src, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully downloaded to %s\n", dst)
}
