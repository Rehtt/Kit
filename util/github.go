package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type FileInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Sha         string `json:"sha"`
	Size        int    `json:"size"`
	Url         string `json:"url"`
	HtmlUrl     string `json:"html_url"`
	GitUrl      string `json:"git_url"`
	DownloadUrl string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	Links       struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		Html string `json:"html"`
	} `json:"_links"`
}

func GetGitHubFileInfo(owner, repo, filePath string) (*FileInfo, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, filePath))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out *FileInfo
	err = json.NewDecoder(resp.Body).Decode(&out)
	return out, err
}
