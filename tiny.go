package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const preserveOriginals = false
const usage = `
USAGE
	tiny cat.jpg dog.png
	tiny .
`

func main() {
	if len(os.Args) < 2 ||
		len(os.Args) > 2 && os.Args[1] == "." {
		exitUsage()
	}

	var err error

	var paths []string
	if os.Args[1] == "." {
		paths, err = findImageFiles()
		if err != nil {
			log.Fatalf("Couldn't find image files: %s\n", err)
		}
	} else {
		for _, path := range os.Args[1:] {
			paths = append(paths, path)
		}
	}

	// pathStrs := strings.Join(paths, "\n")
	// fmt.Println("Paths:")
	// fmt.Println(pathStrs)
	// os.Exit(0)

	for _, path := range paths {
		fmt.Printf(`Shrinking "%s"... `, path)

		if preserveOriginals {
			if err = makeBackup(path); err != nil {
				fmt.Printf("Failed to back up: %s. Skipping.\n", err)
				continue
			}
		}

		delayMs := 501 + rand.Intn(500)
		time.Sleep(time.Duration(delayMs) * time.Millisecond)

		shrinkResp, err := postShrink(path)
		if err != nil {
			fmt.Printf("Failed to shrink: %s\n", err)
			continue
		}

		if err = download(shrinkResp.Output.Url, path); err != nil {
			fmt.Printf("Failed to download %s: %s\n", shrinkResp.Output.Url, err)
			continue
		}

		savedPcn := int(100 * (1 - shrinkResp.Output.Ratio))
		fmt.Printf("OK, saved %d%%!\n", savedPcn)
	}
}

func exitUsage() {
	fmt.Println(strings.TrimSpace(usage))
	os.Exit(0)
}

var imageExts = []string{".jpg", ".jpeg", ".png"}

func findImageFiles() (paths []string, err error) {
	dirEntries, err := os.ReadDir(".")
	if err != nil {
		return paths, err
	}

	for _, entry := range dirEntries {
		if entry.Type().IsRegular() && slices.Contains(imageExts, filepath.Ext(entry.Name())) {
			paths = append(paths, entry.Name())
		}
	}

	return paths, nil
}

func makeBackup(path string) (err error) {
	file, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf(`can't open "%s": %s`, path, err)
		return err
	}
	defer file.Close()

	ext := filepath.Ext(path)
	pathNoExt := strings.TrimSuffix(path, ext)
	backupPath := pathNoExt + "_original" + ext

	backupFile, err := os.Create(backupPath)
	if err != nil {
		err = errors.New(fmt.Sprintf(`can't open backup "%s": %s`, path, err))
		return err
	}
	defer backupFile.Close()

	_, err = io.Copy(backupFile, file)
	if err != nil {
		err = errors.New(fmt.Sprintf(`can't back up "%s" to "%s": %s`, path, backupPath, err))
		return err
	}

	return nil
}

type shrinkResp struct {
	Output struct {
		Url   string  `json:"url"`
		Ratio float32 `json:"ratio"`
	} `json:"output"`
}

func postShrink(path string) (resp shrinkResp, err error) {
	// "curl",
	// "https://tinyjpg.com/backend/opt/shrink",
	// "-X", "'POST'",
	// "-H", "'authority: tinyjpg.com'",
	// "-H", "'accept: */*'",
	// "-H", "'accept-language: en-US,en;q=0.9,fr-SN;q=0.8,fr;q=0.7,lv;q=0.6'",
	// "-H", "'content-type: image/jpeg'",
	// "-H", "'cookie: __stripe_mid=e9c518c7-f65d-41ce-a024-362f824bff73342cbb; __stripe_sid=da8bec2b-cf29-46d0-8666-fca9dd357310589efd'",
	// "-H", "'origin: https://tinyjpg.com'",
	// "-H", "'referer: https://tinyjpg.com/'",
	// `-H`, `'sec-ch-ua: "Google Chrome";v="119", "Chromium";v="119", "Not?A_Brand";v="24"'`,
	// "-H", "'sec-ch-ua-mobile: ?0'",
	// `-H`, `'sec-ch-ua-platform: "macOS"'`,
	// "-H", "'sec-fetch-dest: empty'",
	// "-H", "'sec-fetch-mode: cors'",
	// "-H", "'sec-fetch-site: same-origin'",
	// "-H", "'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36'",
	// "--compressed",
	// "--data-binary",
	// fmt.Sprintf(`"@%s"`, path),
	//
	// returns:
	// {"input":{"size":30761,"type":"image/jpeg"},"output":{"size":25691,"type":"image/jpeg","width":400,"height":400,"ratio":0.8352,"url":"https://tinyjpg.com/backend/opt/output/5efap4fwyn0kzrjqzjs4z09tm4khcnwp"}}
	// we want to then download the $.output.url

	file, err := os.Open(path)
	if err != nil {
		return resp, err
	}
	req, err := http.NewRequest("POST", "https://tinyjpg.com/backend/opt/shrink", file)
	if err != nil {
		return resp, err
	}
	req.Header.Set("authority", "tinyjpg.com")
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9,fr-SN;q=0.8,fr;q=0.7,lv;q=0.6")

	contentType := "image/jpeg"
	ext := filepath.Ext(path)
	if ext == ".png" {
		contentType = "image/png"
	}
	req.Header.Set("content-type", contentType)

	req.Header.Set("origin", "https://tinyjpg.com")
	req.Header.Set("referer", "https://tinyjpg.com")
	// skipping sec-* headers
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer httpResp.Body.Close()

	err = json.NewDecoder(httpResp.Body).Decode(&resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func download(fromUrl, toLocalPath string) (err error) {
	resp, err := http.Get(fromUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(toLocalPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
