package GOdyndns

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/nudelfabrik/GOdyndns/Digitalocean"
	"github.com/nudelfabrik/GOdyndns/Gandi"
	"github.com/nudelfabrik/GOdyndns/Porkbun"
	"github.com/nudelfabrik/GOdyndns/settings"
)

type Client interface {
	Update(string) error
}

func CreateClient(setting *settings.Settings) (Client, error) {

	switch setting.API {
	case "Gandi", "gandi":
		return Gandi.NewGandiClient(setting)
	case "DO", "do", "DigitalOcean", "digitalocean":
		return Digitalocean.NewDoClient(setting)
	case "Porkbun", "porkbun", "PorkBun":
		return Porkbun.NewPorkbunClient(setting)
	default:
		return nil, fmt.Errorf("Not supported API endpoint: %s", setting.API)
	}

}

func Update(c Client) error {
	ip, err := getIP()
	if err != nil {
		go func(c Client) {
			time.Sleep(time.Minute * 5)
			Update(c)
		}(c)
		return err
	}

	err = c.Update(ip)
	return err
}

func getIP() (string, error) {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	response, err := netClient.Get("http://ipv4.icanhazip.com")
	if err != nil {
		return "", err
	}
	responseText, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		return "", err
	}

	str := string(responseText)
	str = strings.TrimSpace(str)
	ip := net.ParseIP(str)
	if ip == nil {
		return "", errors.New("Cannot Parse IP: " + str)
	}

	return ip.String(), err
}
