package main

import "golang.org/x/oauth2"
import "github.com/digitalocean/godo"
import "log"
import "os"

type TokenSource struct {
	AccessToken string
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

func FirstSnapshot(client *godo.Client) *godo.Image {
	images, _, err := client.Images.ListUser(nil)
	if err != nil {
		log.Println(err)
	}
	return &images[0]
}

func AllSSHKey(client *godo.Client) []godo.Key {
	keys, _, err := client.Keys.List(nil)
	if err != nil {
		log.Println(err)
	}
	return keys
}

func main() {
	pat := os.Getenv("TOKEN")
	tokenSource := &TokenSource{
		AccessToken: pat,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)

	image := FirstSnapshot(client)
	droplet, err := CreateDroplet(client, "sf.myalert.info", image)
	for {
		droplet, _, err = client.Droplets.Get(droplet.ID)
		if err != nil {
			log.Println(err)
			break
		}
		if len(droplet.Networks.V4) > 0 {
			ip := droplet.Networks.V4[0].IPAddress
			log.Println(ip)
			break
		}
	}

	// droplets, err := DropletList(client)
	// log.Println(droplets, err)
}
