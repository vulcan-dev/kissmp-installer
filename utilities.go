package main

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Utilities struct{}

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
    }
    defer d.Close()
    names, err := d.Readdirnames(-1)
    if err != nil {
        return err
    }
    for _, name := range names {
        err = os.RemoveAll(filepath.Join(dir, name))
        if err != nil {
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