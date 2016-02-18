package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/AaronO/gogo-proxy"
)

func main() {
	user := ""
	pass := ""

	p, _ := proxy.New(proxy.ProxyOptions{
		Balancer: func(req *http.Request) (string, error) {
			// Auth
			if isV2(req) {
				// V2
				if token := GetToken(getV2Repo(req), user, pass); token != "" {
					req.Header.Set("Authorization", "Bearer "+token)
				}
			} else {
				// V1
				req.SetBasicAuth(user, pass)
			}

			return "https://index.docker.io/", nil
		},
	})

	http.ListenAndServe(":8080", p)
}

func isV2(req *http.Request) bool {
	return strings.HasPrefix(req.URL.String(), "/v2/")
}

func getV2Repo(req *http.Request) string {
	url := req.URL.String()
	parts := strings.Split(url, "/")

	if len(parts) < 4 {
		return ""
	}

	return strings.Join(parts[2:4], "/")
}

type tokenResponse struct {
	Token    string `json:"token"`
	ExpresIn string `json:"expires_in"`
	IssuedAt string `json:"issued_at"`
}

func GetToken(repo, user, pass string) string {
	// URL
	url := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repo)

	// Setup request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	req.SetBasicAuth(user, pass)

	// Do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	// Read body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	// Parse JSON
	tokenR := tokenResponse{}
	if err := json.Unmarshal(data, &tokenR); err != nil {
		return ""
	}

	return tokenR.Token
}
