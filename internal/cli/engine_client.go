package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func callEngineStart(port int, count int) error {
	url := "http://127.0.0.1:" + strconv.Itoa(port) + "/start"
	body, _ := json.Marshal(map[string]int{"count": count})
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		detail := strings.TrimSpace(string(respBody))
		if detail != "" {
			return fmt.Errorf("engine returned status %d: %s", resp.StatusCode, detail)
		}
		return fmt.Errorf("engine returned status %d", resp.StatusCode)
	}
	return nil
}

func callEngineStop(port int) error {
	url := "http://127.0.0.1:" + strconv.Itoa(port) + "/stop"
	req, err := http.NewRequest("POST", url, bytes.NewReader(nil))
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("engine returned status %d", resp.StatusCode)
	}
	return nil
}

func callEngineScale(port int, count int) error {
	url := "http://127.0.0.1:" + strconv.Itoa(port) + "/scale"
	body, _ := json.Marshal(map[string]int{"count": count})
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("engine returned status %d", resp.StatusCode)
	}
	return nil
}
