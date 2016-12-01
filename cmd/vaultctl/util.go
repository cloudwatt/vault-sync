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

package main

import (
	"fmt"

	"github.com/UKHomeOffice/vaultctl/pkg/api"
	"github.com/UKHomeOffice/vaultctl/pkg/utils"
	"github.com/UKHomeOffice/vaultctl/pkg/vault"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

// parseConfigFiles parses the configuration files and extracts the resources
func parseConfigFiles(files []string) (*resources, error) {
	r := new(resources)

	// step: iterate the configuration files and decode
	for _, c := range files {
		cfg := new(api.Config)
		if err := utils.DecodeFile(c, cfg); err != nil {
			return nil, fmt.Errorf("unable to decode the file: %s, error: %s", c, err)
		}
		// step: appends the elements
		r.users = append(r.users, cfg.Users...)
		r.backends = append(r.backends, cfg.Backends...)
		r.secrets = append(r.secrets, cfg.Secrets...)
		r.auths = append(r.auths, cfg.Auths...)
	}

	return r, nil
}

// getKubeClient retrieves a kube client
func getKubeClient(cx *cli.Context) (*unversioned.Client, error) {
	filename := cx.String("kubeconfig")
	context := cx.String("kube-context")
	hostname := cx.String("kube-server")
	token := cx.String("kube-token")

	config := new(unversioned.Config)

	// step: are we using a kubeconfig?
	if filename != "" {
		log.Infof("loading the kubeconfig from: %s, context: %s, server: %s", filename, context, hostname)

		kube, err := clientcmd.LoadFromFile(filename)
		if err != nil {
			return nil, err
		}

		if config, err = clientcmd.NewDefaultClientConfig(*kube,
			&clientcmd.ConfigOverrides{
				CurrentContext: context,
			},
		).ClientConfig(); err != nil {
			return nil, err
		}

		if hostname != "" {
			config.Host = hostname
		}
	} else {
		config.BearerToken = token
		config.Host = hostname
		config.Insecure = true
	}

	log.Infof("using the kube api: %s", config.Host)

	// step: create the client
	client, err := unversioned.New(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// getVaultClient retrieves a vault client for use
func getVaultClient(cx *cli.Context) (*vault.Client, error) {
	host := cx.GlobalString("vault-addr")
	username := cx.GlobalString("vault-username")
	password := cx.GlobalString("vault-password")
	token := cx.GlobalString("vault-token")
	creds := cx.GlobalString("credentials")

	// step: validate we have the requirements
	if creds != "" {
		if !utils.IsFile(creds) {
			printUsage("the vault credentials file does not exist")
		}
	} else if token == "" {
		if username == "" || password == "" {
			return nil, fmt.Errorf("you need to specify a username and password if no token")
		}
	}

	// step: create a vault client
	client, err := vault.New(host, username, password, creds, token)
	if err != nil {
		return nil, err
	}

	return client, nil
}
