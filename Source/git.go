package main

import (
	"encoding/json"
	"net/http"
	"os"
)

type Git struct {
	Version string `json:"tag_name"`
	Assets  []struct {
		DownloadURL string `json:"browser_download_url"`
		Name        string `json:"name"`
	}
	Body string `json:"body"`
}

func (git Git) GetJSONData(url string) (*Git, error) {
	req, err := http.NewRequest("GET", url, nil); if err != nil {
		return nil, err
	}

	req.Header = http.Header{
		"Host":         []string{"api.github.com"},
		"Content-Type": []string{"application/json"},
		"User-Agent":   []string{"PostmanRuntime/7.28.0"},
	}

	// if a token exists then use it (will prevent you from getting limited on api.gitub.com, i need to find a better way)
	auth, exists := os.LookupEnv("GITHUB_TOKEN"); if exists {
		req.Header.Add("Authorization", "token " + auth)
	}

	client := &http.Client{}
	response, err := client.Do(req); if err != nil {
		return nil, err
	}

	// store all the body json into my json struct for accessing data such as filename, version, etc...
	if err = json.NewDecoder(response.Body).Decode(&git); err != nil {
		return nil, err
	}

	return &git, err
}