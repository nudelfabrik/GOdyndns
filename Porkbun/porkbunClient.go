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
    domainid  string
}

type record struct {
	SecretAPI string `json:"secretapikey"`
	APIKey    string `json:"apikey"`
	TTL       string `json:"ttl"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Content   string `json:"content"`
}

type answer struct {
    Records []rec `json:"records"`

}
type rec struct {
    ID string `json:"id"`

}

func NewPorkbunClient(setting *settings.Settings) (*PorkbunClient, error) {
	client := &PorkbunClient{}

	client.apikey = setting.ApiKey
	client.token = setting.Token
	client.domain = setting.Domain
	client.subdomain = setting.Subdomain
	setting.Token = ""

    // get id
    client.getDomainID()

	return client, nil
}

func (c *PorkbunClient) getDomainID() error {
	request := record{}
	request.APIKey = c.apikey
	request.SecretAPI = c.token

	data, err := json.Marshal(request)

    req, err := http.NewRequest("POST", "https://api.porkbun.com/api/json/v3/dns/retrieveByNameType/"+c.domain+"/A/"+c.subdomain, bytes.NewBuffer(data))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
    target := answer{}
    err = json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return err
	}
    if len(target.Records) == 0 {
        return nil
    }
    c.domainid = target.Records[0].ID
    return nil
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
	request.Type = "A"
	request.Name = c.subdomain
	request.Content = ip

	data, err := json.Marshal(request)

	client := &http.Client{}

    req, err := http.NewRequest("POST", "https://api.porkbun.com/api/json/v3/dns/delete/"+c.domain+ "/" + c.domainid, bytes.NewBuffer(data))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

    req, err = http.NewRequest("POST", "https://porkbun.com/api/json/v3/dns/create/" + c.domain, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
        respDump, _ := httputil.DumpResponse(resp, true)
        fmt.Println(string(respDump))
		return errors.New(resp.Status)
	}
    c.getDomainID()

	fmt.Println(time.Now().Format(time.RFC1123), " Changed IP: ", ip)
	return nil
}
