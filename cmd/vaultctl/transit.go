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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/UKHomeOffice/vaultctl/pkg/utils"
	"github.com/UKHomeOffice/vaultctl/pkg/vault"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

type transitCommand struct {
	// the vault client
	client *vault.Client
	// a list of files to action upon
	files []string
	// the transit path
	transit string
	// the transit key to use
	key string
	// the operation
	encrypting bool
	// is decryption
	decryption bool
	// print to stdout
	stdout bool
	// is the extension to save files
	savedExt string
	// is the extension to glob for in directories
	globExt string
	// whether to delete files post
	deleteFiles bool
}

func newTransitCommand() cli.Command {
	return new(transitCommand).getCommand()
}

func (r *transitCommand) action(cx *cli.Context) error {
	r.files = cx.StringSlice("file")

	if r.transit == "" {
		return fmt.Errorf("you have not specified a transit path")
	}
	if r.key == "" {
		return fmt.Errorf("you have not specified a transit key")
	}
	if !r.encrypting && !r.decryption {
		return fmt.Errorf("you have to choose encryption or decryption")
	}

	// step: get a vault client
	client, err := getVaultClient(cx)
	if err != nil {
		return err
	}
	r.client = client

	// step: get a list of files to encrypt
	list, err := utils.FindFilesInDirectory(cx.StringSlice("directory"), cx.String("extension"))
	if err != nil {
		return err
	}
	r.files = append(r.files, list...)

	switch r.encrypting {
	case true:
		err = r.encryptContent()
	default:
		err = r.decryptContent()
	}

	return err
}

// encryptContent encrypts the file contents
func (r *transitCommand) encryptContent() error {
	// step: iterate the files and encrypt
	for _, f := range r.files {
		// step: save the encrypted content
		filename := fmt.Sprintf("%s%s", f, r.savedExt)
		// step: read the contents of the file
		content, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		// step: encrypt the content
		encrypted, err := r.client.Encrypt(r.transit, r.key, bytes.NewReader(content))
		if err != nil {
			return err
		}
		if r.stdout {
			fmt.Fprintf(os.Stdout, "%s", encrypted)
			continue
		}
		// step: save the encrypted content to the file
		log.Infof("saving the encrypted content to from: %s to: %s", f, filename)
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0640)
		if err != nil {
			return err
		}
		file.WriteString(fmt.Sprintf("%s", encrypted))
		// step: are we deleting the original file?
		if r.deleteFiles {
			if err := os.Remove(f); err != nil {
				return err
			}
		}
	}

	return nil
}

// decryptContent decrypts the content
func (r *transitCommand) decryptContent() error {
	// step: iterate the files and encrypt
	for _, f := range r.files {
		// step: save the decrypted content to
		filename := fmt.Sprintf("%s", strings.TrimSuffix(f, r.savedExt))
		// step: read the contents of the file
		content, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		// step: decrypt the content
		decrypted, err := r.client.Decrypt(r.transit, r.key, bytes.NewReader(content))
		if err != nil {
			return err
		}
		if r.stdout {
			fmt.Fprintf(os.Stdout, "%s", decrypted)
			continue
		}
		log.Infof("saving the decrypted content from: %s, to: %s", f, filename)
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0640)
		if err != nil {
			return err
		}
		file.WriteString(fmt.Sprintf("%s", decrypted))
		// step: are we deleting the original file?
		if r.deleteFiles {
			if err := os.Remove(f); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *transitCommand) getCommand() cli.Command {
	return cli.Command{
		Name:    "transit",
		Aliases: []string{"tr", "trans"},
		Usage:   "Encrypts / decrypts files using the Vault transit backend",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:        "e, encrypt",
				Usage:       "encrypt the contents of the files using the transit service",
				Destination: &r.encrypting,
			},
			cli.BoolFlag{
				Name:        "d, decrypt",
				Usage:       "decrypt the contents of the files using the transit service",
				Destination: &r.decryption,
			},
			cli.StringFlag{
				Name:        "t, transit",
				Usage:       "the vault transit endpoint used for encryption operations",
				Destination: &r.transit,
			},
			cli.StringFlag{
				Name:        "k, key",
				Usage:       "the name of the key in the transit backend to use",
				Destination: &r.key,
			},
			cli.StringSliceFlag{
				Name:  "f, file",
				Usage: "the path to a file you wish to encrypt",
			},
			cli.StringSliceFlag{
				Name:  "D, directory",
				Usage: "the path to a directory containing files you wosh to encrypt",
			},
			cli.BoolFlag{
				Name:        "O, stdout",
				Usage:       "print the output to stdout rather than saving in the file",
				Destination: &r.stdout,
			},
			cli.BoolFlag{
				Name:        "delete",
				Usage:       "delete the plaintext files after content encrypted",
				Destination: &r.deleteFiles,
			},
			cli.StringFlag{
				Name:        "glob-filter",
				Usage:       "the file extension when using directories to glob for",
				Value:       "*.yml",
				Destination: &r.globExt,
			},
			cli.StringFlag{
				Name:        "file-extension",
				Usage:       "is the extension applied to files when the content is encrypted",
				Value:       ".enc",
				Destination: &r.savedExt,
			},
		},
		Action: func(cx *cli.Context) {
			executeCommand(cx, r.action)
		},
	}
}
