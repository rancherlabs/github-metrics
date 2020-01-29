package metrics

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func getAuthHeader(user, token string) string {
	if user != "" && token != "" {
		s := user + ":" + token
		return "Basic " + base64.StdEncoding.EncodeToString([]byte(s))
	}
	return ""
}

func getJSON(url, user, token string, insecure bool, target interface{}) (string, error) {

	start := time.Now()

	log.Debug("Connecting to ", url)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 4 {
				return fmt.Errorf("stopped after 4 redirects")
			}
			req.Header.Add("Authorization", getAuthHeader(user, token))
			return nil
		},
		Timeout: 15 * time.Second,
	}

	if insecure {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		}
		client.Transport = tr
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("Error Creating request: %v", err)
	}
	req.Header.Add("Authorization", getAuthHeader(user, token))
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error Collecting JSON from API: %v", err)
	}
	respFormatted := json.NewDecoder(resp.Body).Decode(target)
	resp.Body.Close()

	log.Debug("Time to get json: ", float64((time.Since(start))/time.Millisecond), " ms")

	return getNext(resp.Header.Get("Link")), respFormatted
}

func getNext(header string) string {
	if len(header) == 0 {
		return ""
	}

	linkFormat, err := regexp.Compile("^ *<([^>]+)> *; *rel=\"next\"")
	if err != nil {
		log.Error("Error checking header format ", err)
		return ""
	}

	for _, line := range strings.Split(header, ",") {
		if linkFormat.MatchString(line) {
			return linkFormat.ReplaceAllString(line, "$1")
		}
	}

	return ""
}
