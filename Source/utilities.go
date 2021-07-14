package main

import (
	"archive/zip"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Utilities struct{}

type Git struct {
	Version string `json:"tag_name"`
	Assets []struct {
		DownloadURL string `json:"browser_download_url"`
		Name string `json:"name"`
	}
}

func (git Git) GetJSONData(url string) (*Git, error) {
	req, err := http.NewRequest("GET", url, nil); if err != nil {
		return nil, err
	};

	req.Header = http.Header{
		"Host": []string{"api.github.com"},
		"Content-Type": []string{"application/json"},
		"User-Agent": []string{"PostmanRuntime/7.28.0"},
	}
	
	auth, exists := os.LookupEnv("GITHUB_TOKEN")
	if exists {
		req.Header.Add("Authorization", auth)
		log.Infoln("Using Github Auth")
	}

	client := &http.Client{}
	response, err := client.Do(req); if err != nil {
		return nil, err
	}

	if err = json.NewDecoder(response.Body).Decode(&git); err != nil {
		return nil, err
	}

	return &git, err
}

func (utilities *Utilities) CreateFile(filename string, data []byte) error {
	file, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0755); if err != nil {
		return err
    }

    _, err = file.Write([]byte(data)); if err != nil {
		return err
    }; file.Close()
	
	return err
}

func (utilities *Utilities) DownloadFile(url string, path string) error {
	response, err := http.Get(url); if err != nil {
		return err
	}; defer response.Body.Close()

	/* Create the File */
	out, err := os.Create(path); if err != nil {
		return err
	}; defer out.Close()

	_, err = io.Copy(out, response.Body)

	return err
}

func (utilities *Utilities) Unzip(path string, dest string) ([]string, error) {
	var filenames[] string

	reader, err := zip.OpenReader(path); if err != nil {
		return filenames, err
	}; defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(dest, file.Name)

		filenames = append(filenames, filePath)
		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, file.Mode()); if err != nil {
			return filenames, err
		}

		rc, err := file.Open(); if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}

	return filenames, err
}

func (utilities *Utilities) DeleteDirectory(dir string) error {
    d, err := os.Open(dir)
    if err != nil {
        return err
    }; defer d.Close()

    names, err := d.Readdirnames(-1); if err != nil {
        return err
    }

    for _, name := range names {
        err = os.RemoveAll(filepath.Join(dir, name)); if err != nil {
            return err
        }
    }

    return nil
}

func (utilities *Utilities) Exists(name string) bool {
    if _, err := os.Stat(name); err != nil {
        if os.IsNotExist(err) {
            return false
        }
    }

    return true
}