package api

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/viper"
)

type Client struct {
	HttpClient http.Client
}

func NewClient(timeout time.Duration) Client {
	currProxy := viper.GetString("proxy")
	if currProxy != "" {
		urlProxy, err := url.Parse(currProxy)
		if err != nil {
			fmt.Println("Invalid proxy, running normal client")
			return Client{
				HttpClient: http.Client{
					Timeout: timeout,
				},
			}
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(urlProxy),
		}
		return Client{
			HttpClient: http.Client{
				Transport: transport,
				Timeout:   timeout,
			},
		}
	}
	return Client{
		HttpClient: http.Client{
			Timeout: timeout,
		},
	}
}
