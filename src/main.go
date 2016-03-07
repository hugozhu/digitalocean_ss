package main

import (
	"encoding/json"
	"flag"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type TokenSource struct {
	AccessToken string
}

type Dnspod struct {
	Email     string `json:"login_email,omitempty"`
	Password  string `json:"login_password,omitempty"`
	DomainId  int    `json:"domain_id,omitempty"`
	RecordId  int    `json:"record_id,omitempty"`
	SubDomain string `json:"sub_domain,omitempty"`
}

type Config struct {
	Token  string
	Domain string
	Dnspod Dnspod
}

func UpdateDnspod(ip string) error {
	client := &http.Client{}
	body := url.Values{
		"login_email":    {c.Dnspod.Email},
		"login_password": {c.Dnspod.Password},
		"format":         {"json"},
		"domain_id":      {strconv.Itoa(c.Dnspod.DomainId)},
		"record_id":      {strconv.Itoa(c.Dnspod.RecordId)},
		"sub_domain":     {c.Dnspod.SubDomain},
		"record_line":    {"默认"},
		"value":          {ip},
	}
	req, err := http.NewRequest("POST", "https://dnsapi.cn/Record.Ddns", strings.NewReader(body.Encode()))
	req.Header.Set("Accept", "text/json")
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)
	log.Println(string(bytes))
	return err
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func DropletList(client *godo.Client) ([]godo.Droplet, error) {
	// create a list to hold our droplets
	list := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := client.Droplets.List(opt)
		if err != nil {
			return nil, err
		}

		// append the current page's droplets to our list
		for _, d := range droplets {
			list = append(list, d)
		}

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}

func CreateDroplet(client *godo.Client, dropletName string, image *godo.Image) (*godo.Droplet, error) {
	keys := AllSSHKey(client)
	createKeys := make([]godo.DropletCreateSSHKey, len(keys))
	for i, key := range keys {
		createKeys[i].ID = key.ID
	}

	createRequest := &godo.DropletCreateRequest{
		Name:   dropletName,
		Region: "sfo1",
		Size:   "512mb",
		Image: godo.DropletCreateImage{
			ID: image.ID,
		},
		SSHKeys: createKeys,
	}

	newDroplet, _, err := client.Droplets.Create(createRequest)
	return newDroplet, err
}

func SnapshotByName(client *godo.Client, name string) *godo.Image {
	images, _, err := client.Images.ListUser(nil)
	if err != nil {
		log.Println(err)
	}
	for _, image := range images {
		if image.Name == name {
			return &image
		}
	}
	return nil
}

func AllSSHKey(client *godo.Client) []godo.Key {
	keys, _, err := client.Keys.List(nil)
	if err != nil {
		log.Println(err)
	}
	return keys
}

func DeleteDroplet(client *godo.Client, name string) error {
	droplets, err := DropletList(client)
	if err != nil {
		return err
	}
	for _, droplet := range droplets {
		if droplet.Name != name {
			continue
		}
		_, err := client.Droplets.Delete(droplet.ID)
		if err != nil {
			log.Println("failed to delete droplet:", droplet.Name, err)
			return err
		} else {
			log.Println("delete droplet:", droplet.Name)
		}
	}
	return nil
}

var config string
var ip string
var destroy bool
var create bool
var c Config

func init() {
	flag.StringVar(&config, "config", "config.json", "config file path")
	flag.BoolVar(&destroy, "destroy", false, "destroy all droplets")
	flag.BoolVar(&create, "create", false, "create a droplet")
	flag.Parse()
	bytes, _ := ioutil.ReadFile(config)
	json.Unmarshal(bytes, &c)
}

func main() {
	pat := c.Token
	tokenSource := &TokenSource{
		AccessToken: pat,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)

	if destroy {
		DeleteDroplet(client, c.Domain)
	} else if create {
		image := SnapshotByName(client, c.Domain)
		if image == nil {
			log.Fatal("Can't find snapshot:", c.Domain)
			os.Exit(-1)
		}
		droplet, err := CreateDroplet(client, c.Domain, image)
		for {
			droplet, _, err = client.Droplets.Get(droplet.ID)
			if err != nil {
				log.Println(err)
				break
			}
			if !droplet.Locked && len(droplet.Networks.V4) > 0 {
				ip = droplet.Networks.V4[0].IPAddress
				break
			}
			log.Print(".")
			time.Sleep(1)
		}
		log.Println(ip)
		UpdateDnspod(ip)
	}
}
