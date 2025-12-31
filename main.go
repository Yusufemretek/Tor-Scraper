package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/proxy"
)

type TorResponse struct {
	IsTor bool   `json:"IsTor"`
	IP    string `json:"IP"`
}

func readTarget(filePath string) ([]string, error) {
	var targets []string
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			targets = append(targets, line)
		}
	}
	return targets, scanner.Err()
}

func GetTorClient() (*http.Client, error) {
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9150", nil, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{Dial: dialer.Dial},
		Timeout:   time.Minute,
	}, nil
}

func VerifyTor(client *http.Client) (bool, string) {
	resp, err := client.Get("https://check.torproject.org/api/ip")
	if err != nil {
		return false, "Connection Error"
	}
	defer resp.Body.Close()

	var result TorResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result.IsTor, result.IP
}

func takeScreenshot(url, safeName string) error {
	os.MkdirAll("results/photos", 0755)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer("socks5://127.0.0.1:9150"),
		chromedp.NoSandbox,
		chromedp.Headless,
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(5*time.Second),
		chromedp.FullScreenshot(&buf, 90),
	); err != nil {
		return err
	}

	return os.WriteFile(fmt.Sprintf("results/photos/%s.png", safeName), buf, 0644)
}

func writeLog(message string) {
	f, _ := os.OpenFile("scan_report.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), message))
}

func fetchAndSave(client *http.Client, url string) error {
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	resp, err := client.Get(url)
	if err != nil {
		writeLog("FAILED: " + url)
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("[Status: %d] ", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		writeLog(fmt.Sprintf("INACTIVE: %s (Status: %d)", url, resp.StatusCode))
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	safeName := strings.NewReplacer("/", "_", ":", "_", ".", "_", "https__", "", "http__", "").Replace(url)
	os.MkdirAll("results", 0755)
	os.WriteFile(fmt.Sprintf("results/%s.html", safeName), body, 0644)
	writeLog("ACTIVE: " + url)
	fmt.Print("(Taking Screenshot...) ")
	return takeScreenshot(url, safeName)
}

func main() {
	targets, err := readTarget("targets.yaml")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	client, _ := GetTorClient()
	fmt.Print("Security check... ")
	isSecure, currentIP := VerifyTor(client)
	if !isSecure {
		fmt.Printf("\n[CRITICAL] Leak detected! IP: %s\n", currentIP)
		return
	}
	fmt.Printf("SECURE (IP: %s)\n\n", currentIP)

	for _, url := range targets {
		fmt.Printf("Processing: %s ", url)
		if err := fetchAndSave(client, url); err != nil {
			fmt.Printf("-> ERROR: %v\n", err)
		} else {
			fmt.Println("-> SUCCESS.")
		}
	}
}
