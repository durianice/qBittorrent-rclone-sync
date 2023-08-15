package util

import (
	"fmt"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"
	netUrl "net/url"
	"net/http/cookiejar"
	"golang.org/x/net/publicsuffix"
)

var client *http.Client
var jar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})


func init() {
	def := http.DefaultTransport
	defPot, ok := def.(*http.Transport)
	if !ok {
		panic("Init Request Error")
	}
	defPot.MaxIdleConns = 100
	defPot.MaxIdleConnsPerHost = 100
	defPot.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client = &http.Client{
		Timeout:   time.Second * time.Duration(20),
		Transport: defPot,
		Jar: jar,
	}
}

func Get(url string, header map[string]string, params map[string]interface{}) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range header {
		req.Header.Add(key, value)
	}

	query := req.URL.Query()
	if params != nil {
		for key, val := range params {
			v, _ := toString(val)
			query.Add(key, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println(err)
		return "", err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(bodyBytes), nil
}

func Post(url string, header map[string]string, params map[string]interface{}) (string, error) {
	dd, _ := json.Marshal(params)
	re := bytes.NewReader(dd)
	req, err := http.NewRequest("POST", url, re)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range header {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(bodyBytes), nil
}

func PostForm(url string, header map[string]string, params map[string]string) (string, error) {
	formValue := netUrl.Values{}
	for key, value := range params {
		strValue, _ := toString(value)
		formValue.Set(key, strValue)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(formValue.Encode()))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range header {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(bodyBytes), nil
}

func PostMultipart(url string, header map[string]string, payload *bytes.Buffer) (string, error) {
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	for key, value := range header {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(bodyBytes), nil
}
