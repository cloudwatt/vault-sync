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
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/UKHomeOffice/vaultctl/pkg/api"
	"github.com/UKHomeOffice/vaultctl/pkg/utils"
	"github.com/UKHomeOffice/vaultctl/pkg/vault"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	v "github.com/hashicorp/vault/api"
)

type syncCommand struct {
	// a collection of auths
	auths []*api.Auth
	// a collection of users
	users []*api.User
	// a collection of backends
	backends []*api.Backend
	// a collection of secrets
	secrets []*api.Secret
	// a collection of polices files
	policyFiles []string
	// a list of configuration files
	configFiles []string
	// whether to perform a full sync
	fullsync bool
	// the vault client
	client *vault.Client
	// whether to skip applying the auths to vault
	skipAuths bool
	// whether to skip applying the users to vault
	skipUsers bool
	// whether to skip applying the backend's to vault
	skipBackends bool
	// whether to skip applying the policies to vault
	skipPolicies bool
	// whether to skip applying the secrets to vault
	skipSecrets bool
	// delete will delete any resources no longer referenced
	delete bool
	// policyExtension
	policyExtension string
	// configExtension
	configExtension string
}

// newSyncCommand create a new sync command
func newSyncCommand() cli.Command {
	return new(syncCommand).getCommand()
}

func (r *syncCommand) action(cx *cli.Context) error {
	startTime := time.Now()
	// step: valid the command line options
	if err := r.validateAction(cx); err != nil {
		return err
	}
	// step: get a vault client
	client, err := getVaultClient(cx)
	if err != nil {
		return err
	}
	r.client = client

	// step: parse the configuration files
	if err := r.parseConfigFiles(); err != nil {
		return err
	}
	// step: synchronize the elements
	if err := r.synchronize(); err != nil {
		return err
	}

	log.Infof("synchronization complete, time took: %s", time.Now().Sub(startTime).String())

	return nil
}

// synchronize process the items and sync them
func (r *syncCommand) synchronize() error {
	if !r.skipAuths {
		if err := r.applyAuths(r.auths); err != nil {
			return err
		}
	}
	if !r.skipPolicies {
		if err := r.applyPolicies(r.policyFiles); err != nil {
			return err
		}
	}
	if !r.skipUsers {
		if err := r.applyUsers(r.users); err != nil {
			return err
		}
	}
	if !r.skipBackends {
		if err := r.applyBackends(r.backends); err != nil {
			return err
		}
	}
	if !r.skipSecrets {
		if err := r.applySecrets(r.secrets); err != nil {
			return err
		}
	}

	return nil
}

// applyAuths applies the auth backends
func (r *syncCommand) applyAuths(auths []*api.Auth) error {
	log.Infof("%s", color.GreenString("-> synchronizing the auth backends, %d backends", len(auths)))

	var list []string

	for _, x := range auths {
		// step: check the backend is valid
		if err := x.IsValid(); err != nil {
			return err
		}
		list = append(list, x.Path)

		// step: check the backend is mounted
		mounted, err := r.client.Client().Sys().ListAuth()
		if err != nil {
			return err
		}

		// step: if not mounted? attempt to mount
		if _, found := mounted[x.Path+"/"]; !found {
			log.Infof("[auth: %s] type: %s is not mounted, attempting to mount now", x.Path, x.Type)
			if err := r.client.Client().Sys().EnableAuth(x.Path, x.Type, x.Description); err != nil {
				return err
			}
		} else {
			log.Infof("[auth: %s] already create, skipping to configuration", x.Path)
		}

		// step: config the backend
		for _, c := range x.Attrs {
			// step: get the full path
			uri := fmt.Sprintf("/auth/%s", c.GetPath(x.Path))
			// step: check its valid
			if err := c.IsValid(); err != nil {
				return fmt.Errorf("the attribute for auth backend: %s invalid, error: %s", x.Path, err)
			}
			log.Infof("[auth->config: %s] applying configuration to auth", uri)

			resp, err := r.client.Request("PUT", uri, &c)
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusNoContent {
				return err
			}
		}
	}

	if r.fullsync {
		// step: get a list of backends
		mounted, err := r.client.Client().Sys().ListAuth()
		if err != nil {
			return err
		}
		for name, _ := range mounted {
			if utils.ContainedIn(name, []string{"token/"}) {
				continue
			}

			if !utils.ContainedIn(strings.TrimSuffix(name, "/"), list) {
				log.Warnf("[auth: %s] is no longer referenced, delete: %t", name, r.delete)
				if !r.delete {
					continue
				}
				if err := r.client.Client().Sys().DisableAuth(name); err != nil {
					return fmt.Errorf("failed to disable the auth: %s, error: %s", name, err)
				}
			}
		}
	}

	return nil
}

// applyPolicies applies the policies to a vault instance
func (r *syncCommand) applyPolicies(policies []string) error {
	log.Infof("%s", color.GreenString("-> synchronizing the vault policies, %d files", len(policies)))

	var list []string

	for _, p := range policies {
		name := strings.TrimRight(filepath.Base(p), filepath.Ext(filepath.Base(p)))
		list = append(list, name)

		// step: read in the content
		content, err := ioutil.ReadFile(p)
		if err != nil {
			return err
		}

		if err := r.client.SetPolicy(name, string(content)); err != nil {
			return err
		}

		log.Infof("[policy: %s] successfully applied the policy, filename: %s", name, p)
	}

	if r.fullsync {
		// step: delete any policies no longer referenced
		p, err := r.client.Client().Sys().ListPolicies()
		if err != nil {
			return err
		}

		for _, x := range p {
			if utils.ContainedIn(x, []string{"default", "root"}) {
				continue
			}
			// step: check the policy is referenced still
			if !utils.ContainedIn(x, list) {
				log.Warningf("[policy: %s] no longer referenced in config, delete: %t", x, r.delete)
				if !r.delete {
					continue
				}
				if err := r.client.Client().Sys().DeletePolicy(x); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// applyUsers synchronizes the users with vault
func (r *syncCommand) applyUsers(users []*api.User) error {
	log.Infof("%s", color.GreenString("-> synchronizing the vault users, users: %d", len(users)))

	for _, x := range users {
		// step: validate the user
		if err := x.IsValid(); err != nil {
			return err
		}
		path := x.Path
		if path == "" {
			path = "default"
		}

		log.Infof("[user: %s/%s] ensuring user, policies: %s", path, x.Username(), x.GetPolicies())

		// step: attempt to add the user
		if err := r.client.AddUser(x); err != nil {
			return err
		}
	}

	return nil
}

// syncBackends synchronizes the backend's with vault
func (r *syncCommand) applyBackends(backends []*api.Backend) error {
	log.Infof("%s", color.GreenString("-> synchronizing the backends, backend: %d", len(backends)))

	var list []string

	for _, backend := range backends {
		if err := backend.IsValid(); err != nil {
			return err
		}
		// step: add to the list
		list = append(list, backend.GetPath())

		// step: get the backend path
		path := backend.GetPath()

		// step: get a list of mounts
		mounted, err := r.client.Mounts()
		if err != nil {
			return err
		}

		// step: check if the backend if already mounted
		_, found := mounted[backend.GetPath()+"/"]
		if !found {
			log.Infof("[backend: %s] creating backend", path)
			if err := r.client.Client().Sys().Mount(path, &v.MountInput{
				Type:        backend.Type,
				Description: backend.Description,
				Config: v.MountConfigInput{
					DefaultLeaseTTL: backend.DefaultLeaseTTL.String(),
					MaxLeaseTTL:     backend.MaxLeaseTTL.String(),
				},
			}); err != nil {
				return err
			}
		} else {
			log.Infof("[backend: %s]: already exist, moving to configuration", path)
		}

		// step: apply the configuration
		for _, c := range backend.Attrs {
			// step: get the path
			uri := c.GetPath(path)

			// step: check if a once type setting?
			if found && c.IsOneshot() {
				log.Infof("[backend:%s] skipping the config, as it's a oneshot setting", uri)
				continue
			}

			log.Infof("[backend->config: %s] applying configuration for backend", uri)

			resp, err := r.client.Request("PUT", uri, &c)
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusNoContent {
				return err
			}
		}
	}

	if r.fullsync {
		mounted, err := r.client.Client().Sys().ListMounts()
		if err != nil {
			return err
		}

		// step: remove any backends?
		for name, _ := range mounted {
			// step: skip some inbuilt ones
			if utils.ContainedIn(name, []string{"secret/", "cubbyhole/", "sys/"}) {
				continue
			}
			if !utils.ContainedIn(strings.TrimSuffix(name, "/"), list) {
				log.Warnf("[backend: %s] no longer referenced, delete: %t", name, r.delete)
				if !r.delete {
					continue
				}
				if err := r.client.Client().Sys().Unmount(name); err != nil {
					return fmt.Errorf("failed to unmount the backend: %s, error: %s", name, err)
				}
			}
		}
	}

	return nil
}

// applySecrets synchronizes the secrets in vault
func (r *syncCommand) applySecrets(secrets []*api.Secret) error {
	log.Infof("%s", color.GreenString("-> synchronizing the secrets with vault, secrets: %d", len(secrets)))

	for _, s := range secrets {
		// step: validate the secret
		if err := s.IsValid(); err != nil {
			return err
		}

		log.Infof("[secret: %s] adding the secret", s.Path)

		// step: apply the secret
		if err := r.client.AddSecret(s); err != nil {
			return err
		}
	}

	return nil
}

// parseConfigFiles reads a series of configuration files or directories and extracts the items from them
func (r *syncCommand) parseConfigFiles() error {
	// step: iterate the configuration files and decode
	for _, c := range r.configFiles {
		cfg := new(api.Config)

		if err := utils.DecodeFile(c, cfg); err != nil {
			return fmt.Errorf("unable to decode the file: %s, error: %s", c, err)
		}

		// step: appends the elements
		r.users = append(r.users, cfg.Users...)
		r.backends = append(r.backends, cfg.Backends...)
		r.secrets = append(r.secrets, cfg.Secrets...)
		r.auths = append(r.auths, cfg.Auths...)
	}

	return nil
}

// validateAction validates the inputs from the command line
func (r *syncCommand) validateAction(cx *cli.Context) error {
	r.configFiles = cx.StringSlice("config")

	// step: check the skips
	if r.skipBackends && r.skipPolicies && r.skipUsers {
		return fmt.Errorf("you are skipping all the resources, what exactly are we syncing")
	}
	if r.skipBackends {
		log.Infof("skipping the synchronization of backends")
	}
	if r.skipPolicies {
		log.Infof("skipping the synchronization of policies")
	}
	if r.skipUsers {
		log.Infof("skipping the synchronization of users")
	}

	// step: get the files from any config directories
	files, err := utils.FindFilesInDirectory(cx.StringSlice("config-dir"), r.configExtension)
	if err != nil {
		return err
	}
	r.configFiles = append(r.configFiles, files...)

	// step: get a list of policy files from the directories
	r.policyFiles, err = utils.FindFilesInDirectory(cx.StringSlice("policies"), r.policyExtension)
	if err != nil {
		return err
	}

	return nil
}

// getCommand returns the command set
func (r *syncCommand) getCommand() cli.Command {
	return cli.Command{
		Name:    "synchronize",
		Aliases: []string{"sync"},
		Usage:   "synchonrize the users, policies, secrets and backends",
		Action: func(cx *cli.Context) {
			executeCommand(cx, r.action)
		},
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "c, config",
				Usage: "the path to a configuration file containing users, backends and or secrets",
			},
			cli.StringSliceFlag{
				Name:  "C, config-dir",
				Usage: "the path to a directory containing one of more config files",
			},
			cli.StringSliceFlag{
				Name:  "p, policies",
				Usage: "the path to a directory containing one of more policy files",
			},
			cli.BoolFlag{
				Name:	     "sync-full",
				Usage:       "a full sync will also check the resources are still reference and attempt to delete",
				Destination: &r.fullsync,
			},
			cli.BoolFlag{
				Name:        "delete",
				Usage:       "wheather to delete resources which are no longer referenced",
				Destination: &r.delete,
			},
			cli.BoolFlag{
				Name:        "skip-policies",
				Usage:       "wheather or not to skip synchronizing the policies",
				Destination: &r.skipPolicies,
			},
			cli.BoolFlag{
				Name:        "skip-users",
				Usage:       "wheather or not to skip synchronizing the vault users",
				Destination: &r.skipUsers,
			},
			cli.BoolFlag{
				Name:        "skip-backends",
				Usage:       "wheather or not to skip synchronizing the backends",
				Destination: &r.skipBackends,
			},
			cli.BoolFlag{
				Name:        "skip-secrets",
				Usage:       "wheather or not to skip synchronizing the secrets",
				Destination: &r.skipSecrets,
			},
			cli.StringFlag{
				Name:        "policy-extension",
				Usage:       "the file extenions of the policy files",
				Value:       "*.hcl",
				Destination: &r.policyExtension,
			},
			cli.StringFlag{
				Name:        "config-extension",
				Usage:       "when using a config-dir, the file extension to glob",
				Value:       "*.yml",
				Destination: &r.configExtension,
			},
		},
	}
}
