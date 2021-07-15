package main

import (
	"archive/zip"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

type Utilities struct{}

func (utilities* Utilities) SetupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Infoln("Closing Bridge.")
		os.Exit(0)
	}()
}

func (utilities *Utilities) CreateFile(filename string, data []byte) error {
	file, err := os.OpenFile(filename, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0755); if err != nil {
		return errors.New("failed creating file: " + filename + ". error: " + err.Error())
    }

    _, err = file.Write([]byte(data)); if err != nil {
		return errors.New("failed writing to file: " + filename + ". error: " + err.Error())
    }; file.Close()
	
	return nil
}

func (utilities *Utilities) DownloadFile(url string, path string) error {
	response, err := http.Get(url); if err != nil {
		return errors.New("failed getting url: " + url + ". error code: " + err.Error())
	}; defer response.Body.Close()

	out, err := os.Create(path); if err != nil {
		return errors.New("failed creating file: " + path + ". error: " + err.Error())
	}; defer out.Close()

	_, err = io.Copy(out, response.Body); if err != nil {
		return errors.New("failed copying io-data to io-reader: " + err.Error())
	}

	return nil
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
		}; defer outFile.Close()

		rc, err := file.Open(); if err != nil {
			return filenames, err
		}; defer rc.Close()

		_, err = io.Copy(outFile, rc); if err != nil {
			return filenames, err
		}
	}

	return filenames, err
}

func (utilities *Utilities) DeleteDirectory(dir string) error {
    path, err := os.Open(dir); if err != nil {
        return errors.New("failed opening directory: " + dir + ". error: " + err.Error())
    }; defer path.Close()

    names, err := path.Readdirnames(-1); if err != nil {
        return errors.New("failed getting files in directory: " + dir + ". error: " + err.Error())
    }

    for _, name := range names {
        err = os.RemoveAll(filepath.Join(dir, name)); if err != nil {
            return errors.New("failed removing files in directory: "  + dir + ". error: " + err.Error())
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