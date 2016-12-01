/*
Copyright 2015 All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vault

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/UKHomeOffice/vaultctl/pkg/api"
	"github.com/UKHomeOffice/vaultctl/pkg/utils"

	log "github.com/Sirupsen/logrus"
	v "github.com/hashicorp/vault/api"
)

const (
	apiVersion = "v1"
)

func New(hostname, username, password, filename, token string) (*Client, error) {
	log.Debugf("create vault client to host: %s", hostname)

	// step: get the client configuration
	config := v.DefaultConfig()
	config.Address = hostname
	config.HttpClient = &http.Client{
		Timeout: time.Duration(15) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	// step: get the client
	client, err := v.NewClient(config)
	if err != nil {
		return nil, err
	}

	service := &Client{
		client: client,
	}

	// step: attempt to login
	if filename != "" {
		creds := new(api.UserPass)
		if err := utils.DecodeFile(filename, creds); err != nil {
			return nil, err
		}
		if err := creds.IsValid(); err != nil {
			return nil, err
		}

		token, err := service.userLogin(creds)
		if err != nil {
			return nil, err
		}
		client.SetToken(token)
	} else if username != "" && password != "" {
		token, err := service.userLogin(&api.UserPass{
			Username: username,
			Password: password,
		})
		if err != nil {
			return nil, err
		}
		client.SetToken(token)
	} else {
		client.SetToken(token)
	}

	return &Client{
		client: client,
	}, nil
}

// Clients returns the underlining client
func (r *Client) Client() *v.Client {
	return r.client
}

// AddSecret adds a secret to the vault
func (r *Client) AddSecret(secret *api.Secret) error {
	log.Debugf("adding the secret: %s, %v", secret.Path, secret.Values)
	_, err := r.client.Logical().Write(secret.Path, secret.Values)
	if err != nil {
		return err
	}

	return nil
}

// Mounts is a list of mounts
func (r *Client) Mounts() (map[string]*v.MountOutput, error) {
	return r.client.Sys().ListMounts()
}

// Policies is a list of policies currently in vault
func (r *Client) Policies() (map[string]bool, error) {
	p := make(map[string]bool, 0)

	list, err := r.client.Sys().ListPolicies()
	if err != nil {
		return p, err
	}

	for _, k := range list {
		p[k] = true
	}

	return p, nil
}

// SetPolicy sets a policy in vault
func (r *Client) SetPolicy(name, policy string) error {
	return r.client.Sys().PutPolicy(name, policy)
}

// Request performs a request to vault
func (r *Client) Request(method, uri string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("/%s/%s", apiVersion, strings.TrimPrefix(uri, "/"))

	log.Debugf("make request to %s %s, body: %#v", method, url, body)
	// step: create a request
	request := r.client.NewRequest(method, url)
	if err := request.SetJSONBody(body); err != nil {
		return nil, err
	}

	// step: make the request
	resp, err := r.client.RawRequest(request)
	if err != nil {
		return nil, err
	}

	return resp.Response, nil
}

func (r *Client) userLogin(credentials *api.UserPass) (string, error) {
	log.Debugf("logging into vault service, username: %s", credentials.Username)
	var param struct {
		// Password is the password for the account
		Password string `json:"password"`
	}
	param.Password = credentials.Password

	// step: make the request to vault
	resp, err := r.Request("POST", fmt.Sprintf("auth/userpass/login/%s", credentials.Username), &param)
	if err != nil {
		return "", err
	}

	// step: parse and return auth
	secret, err := v.ParseSecret(resp.Body)
	if err != nil {
		return "", err
	}

	return secret.Auth.ClientToken, nil
}
