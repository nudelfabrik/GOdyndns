package Porkbun

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
    "net/http/httputil"
	"time"

	"github.com/nudelfabrik/GOdyndns/settings"
)

type PorkbunClient struct {
	domain    string
	subdomain string
	recordID  int
	lastIP    string
	token     string
	apikey    string
}

type record struct {
	SecretAPI string `json:"secretapikey"`
	APIKey    string `json:"apikey"`
	TTL       string `json:"ttl"`
	Content   string `json:"content"`
}

func NewPorkbunClient(setting *settings.Settings) (*PorkbunClient, error) {
	client := &PorkbunClient{}

	client.apikey = setting.ApiKey
	client.token = setting.Token
	client.domain = setting.Domain
	client.subdomain = setting.Subdomain
	setting.Token = ""

	return client, nil
}

func (c *PorkbunClient) Update(ip string) error {
	if c.lastIP == ip {
		// Record is up to date
		fmt.Println(time.Now().Format(time.RFC1123), " Record is up to date")
		return nil
	}
	c.lastIP = ip
	request := record{}
	request.APIKey = c.apikey
	request.SecretAPI = c.token
	request.TTL = "1800"
	request.Content = ip

	data, err := json.Marshal(request)

    req, err := http.NewRequest("POST", "https://porkbun.com/api/json/v3/dns/editByNameType/"+c.domain+"/A/"+c.subdomain, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
        respDump, _ := httputil.DumpResponse(resp, true)
        fmt.Println(string(respDump))
		return errors.New(resp.Status)
	}

	fmt.Println(time.Now().Format(time.RFC1123), " Changed IP: ", ip)
	return nil
}
