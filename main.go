package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const repository = "ratelimitpreview/test"
const tokenURL = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + repository + ":pull"
const registryURL = "https://registry-1.docker.io/v2/" + repository + "/manifests/latest"

type dockerHubLimitCollector struct {
	username           string
	password           string
	rateLimit          *prometheus.Desc
	rateLimitRemaining *prometheus.Desc
}

func (collector *dockerHubLimitCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.rateLimit
	ch <- collector.rateLimitRemaining
}

func (collector *dockerHubLimitCollector) Collect(ch chan<- prometheus.Metric) {
	rateLimit, rateLimitRemaining, err := getRegistryLimits(collector.username, collector.password)
	if err != nil {
		log.Printf("couldn't get limits: %s", err)
	}
	ch <- prometheus.MustNewConstMetric(collector.rateLimit, prometheus.GaugeValue, float64(rateLimit))
	ch <- prometheus.MustNewConstMetric(collector.rateLimitRemaining, prometheus.GaugeValue, float64(rateLimitRemaining))
}

func newDockerHubLimitCollector(username string, password string) *dockerHubLimitCollector {
	return &dockerHubLimitCollector{
		username:           username,
		password:           password,
		rateLimit:          prometheus.NewDesc("dockerhub_limit_max_requests_total", "Docker Hub Rate Limit Max Requests", nil, prometheus.Labels{"limit": "max_requests_total"}),
		rateLimitRemaining: prometheus.NewDesc("dockerhub_limit_remaining_requests_total", "Docker Hub Rate Limit Remaining Requests", nil, prometheus.Labels{"limit": "remaining_requests_total"}),
	}
}

type TokenResponse struct {
	Token string `json:"token"`
}

func getToken(username string, password string) (string, error) {
	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return "", err
	}

	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tokenBody, err := ioutil.ReadAll(resp.Body)
	var tokenResponse TokenResponse
	err = json.Unmarshal(tokenBody, &tokenResponse)
	if err != nil {
		return "", err
	}
	return tokenResponse.Token, nil
}

func extractLimit(header string) (int, error) {
	if strings.Contains(header, ";") {
		parts := strings.Split(header, ";")
		limit, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, err
		}
		return limit, nil
	}
	if header != "" {
		limit, err := strconv.Atoi(header)
		if err != nil {
			return 0, err
		}
		return limit, nil
	}
	return 0, nil
}

func getRegistryLimits(username string, password string) (int, int, error) {
	token, err := getToken(username, password)
	client := &http.Client{}
	req, err := http.NewRequest("HEAD", registryURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	limitMap := map[string]int{
		"RateLimit-Limit":     0,
		"RateLimit-Remaining": 0,
		"RateLimit-Reset":     0,
	}
	for key := range limitMap {
		headerValue := resp.Header.Get(key)
		limit, err := extractLimit(headerValue)
		if err != nil {
			return 0, 0, err
		}
		limitMap[key] = limit
	}

	return limitMap["RateLimit-Limit"], limitMap["RateLimit-Remaining"], nil

}
func main() {
	username := os.Getenv("DOCKERHUB_USERNAME")
	password := os.Getenv("DOCKERHUB_PASSWORD")
	port := os.Getenv("DOCKERHUB_EXPORTER_PORT")
	if port == "" {
		port = "8881"
	}
	collector := newDockerHubLimitCollector(username, password)
	prometheus.MustRegister(collector)
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("serving requests at :" + port)
	http.ListenAndServe(":"+port, nil)
}
