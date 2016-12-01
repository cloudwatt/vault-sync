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
	"os"

	"github.com/UKHomeOffice/vaultctl/pkg/utils"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
)

type kubeCmd struct {
	// the kubernetes client
	client *unversioned.Client
	// the kubeHost
	kubeHost string
	// the kubeconfig
	kubeConfig string
	// the content
	kubeContext string
	// the token
	kubeToken string
	// perform a dry run
	dryrun bool
	// the name of the secret to inject
	secretName string
	// the name of the file
	secretFilename string
	// the config extension
	configExtension string
	// the config files
	configFiles []string
}

func newKubeCommand() cli.Command {
	return new(kubeCmd).getCommand()
}

type credential struct {
	Method   string `yaml:"method"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (r *kubeCmd) action(cx *cli.Context) error {
	// step: validate the kubeconfig
	if err := r.validate(cx); err != nil {
		return err
	}
	// step: connect to kubernetes
	client, err := getKubeClient(cx)
	if err != nil {
		return err
	}
	r.client = client

	return r.synchronize()
}

func (r *kubeCmd) synchronize() error {
	// step: get all the users from the config files and inject the
	resources, err := parseConfigFiles(r.configFiles)
	if err != nil {
		return err
	}

	secretName := "vault"
	dataName := "vault.yml"

	// step: iterate the users and inject into k8s
	for _, u := range resources.users {
		// step: skip any users which dont have namespaces or userpass
		if u.Namespace == "" || u.UserPass == nil {
			continue
		}

		log.Infof("[kube %s/%s] adding the vault token to namespace", u.Namespace, u.UserPass.Username)
		cred := &credential{
			Method:   "userpass",
			Username: u.UserPass.Username,
			Password: u.UserPass.Password,
		}

		// step: encode the yaml
		content, err := utils.EncodeConfig(cred, "yml")
		if err != nil {
			return err
		}

		// step: create the secret
		secret := &api.Secret{
			ObjectMeta: api.ObjectMeta{
				Namespace: u.Namespace,
				Name:      secretName,
			},
			Data: map[string][]byte{
				dataName: []byte(content),
			},
		}

		// step: check if the secret is there
		exists, err := r.hasSecret(secretName, u.Namespace)
		if err != nil {
			return err
		}

		if r.dryrun {
			fmt.Fprint(os.Stdout, "%s", secret)
			continue
		}

		if exists {
			_, err := r.client.Secrets(u.Namespace).Update(secret)
			if err != nil {
				return fmt.Errorf("unable to update the secret in namespace: %s, error: %s", u.Namespace, err)
			}
		} else {
			_, err := r.client.Secrets(u.Namespace).Create(secret)
			if err != nil {
				return fmt.Errorf("unable to create the secret in namespace: %s, error: %s", u.Namespace, err)
			}
		}
	}

	return nil
}

// hasSecret check if the secret exists
func (r *kubeCmd) hasSecret(name, namespace string) (bool, error) {
	list, err := r.client.Secrets(namespace).List(labels.Everything(), fields.Everything())
	if err != nil {
		return false, err
	}

	for _, x := range list.Items {
		if x.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (r *kubeCmd) validate(cx *cli.Context) error {
	r.configFiles = cx.StringSlice("config")

	// step: get the files from any config directories
	files, err := utils.FindFilesInDirectory(cx.StringSlice("config-dir"), r.configExtension)
	if err != nil {
		return err
	}
	r.configFiles = append(r.configFiles, files...)

	if len(r.configFiles) <= 0 {
		return fmt.Errorf("you have to specified any configuration files")
	}

	if r.kubeConfig != "" {
		if !utils.IsFile(r.kubeConfig) {
			return fmt.Errorf("the kubeconfig: %s does not exist", r.kubeConfig)
		}
		if r.kubeContext == "" {
			return fmt.Errorf("you have not specified a kubeconfig context")
		}
	}

	if r.kubeConfig == "" {
		if r.kubeToken == "" {
			return fmt.Errorf("you need to specify a token if not using a kubeconfig")
		}
		if r.kubeHost == "" {
			return fmt.Errorf("you need to specify a kube api host if not using a kubeconfig")
		}
	}


	return nil
}

func (r *kubeCmd) getCommand() cli.Command {
	return cli.Command{
		Name:  "kube",
		Usage: "synchronizes the users and injects the vault credentials in namespaces",
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "c, config",
				Usage: "the path to a configuration file containing users, backends and or secrets",
			},
			cli.StringSliceFlag{
				Name:  "C, config-dir",
				Usage: "the path to a directory containing one of more config files",
			},
			cli.StringFlag{
				Name:        "k, kubeconfig",
				Usage:       "the path to the kubeconfig which has the credentails to injects credentials",
				EnvVar:      "KUBECONFIG",
				Destination: &r.kubeConfig,
			},
			cli.StringFlag{
				Name:        "x, kube-context",
				Usage:       "the kubeconfig context to use when connecting",
				Destination: &r.kubeContext,
			},
			cli.StringFlag{
				Name:        "s, kube-server",
				Usage:       "the url for the kubernetes api, will override the one in kubeconfig",
				Destination: &r.kubeHost,
				EnvVar:      "KUBERNETES_HOST",
			},
			cli.StringFlag{
				Name:	     "T, kube-token",
				Usage:       "a bearer token to use for kubernetes authentication",
				Destination: &r.kubeToken,
			},
			cli.BoolFlag{
				Name:        "d, dryrun",
				Usage:       "perform a dryrun, i.e. simply output the secrets to the screen",
				Destination: &r.dryrun,
			},
			cli.StringFlag{
				Name:        "t, token-file",
				Usage:       "the name of the secret to inject into the namespace",
				Value:       "vault-token.yml",
				Destination: &r.secretName,
			},
			cli.StringFlag{
				Name:        "secret-name",
				Usage:       "the name of the secret to add into the namespace",
				Value:       "vault",
				Destination: &r.secretName,
			},
			cli.StringFlag{
				Name:        "secret-filename",
				Usage:       "the name of the secret filename",
				Value:       "vault.yml",
				Destination: &r.secretFilename,
			},
			cli.StringFlag{
				Name:        "config-extension",
				Usage:       "when using a config-dir, the file extension to glob",
				Value:       "*.yml",
				Destination: &r.configExtension,
			},
		},
		Action: func(cx *cli.Context) {
			executeCommand(cx, r.action)
		},
	}
}
