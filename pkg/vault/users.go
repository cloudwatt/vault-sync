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
	"fmt"
	"net/http"
	"strings"

	"github.com/UKHomeOffice/vaultctl/pkg/api"

	log "github.com/Sirupsen/logrus"
	"io/ioutil"
)

type userConfig struct {
	Password string `json:"password"`
	Policies string `json:"policies"`
}

type tokenConfig struct {
	ID string `json:"id,omitempty"`
	DisplayName string `json:"display_name"`
	Policies []string `json:"policies"`
	TTL string `json:"ttl,omitempty"`
	MaxUses int `json:"num_uses"`
}

// AddUser adds a user to vault
func (r *Client) AddUser(user *api.User) error {
	var params interface{}
	// step: set the path
	uri := user.Path
	// step: provision the type
	if user.UserPass != nil {
		if err := user.UserPass.IsValid(); err != nil {
			return err
		}
		// step: use the path or default to the type
		path := "userpass"
		if user.Path != "" {
			path = user.Path
		}
		uri = fmt.Sprintf("auth/%s/users/%s", path, user.UserPass.Username)

		params = &userConfig{
			Password: user.UserPass.Password,
			Policies: strings.Join(user.Policies, ","),
		}
	}
	if user.UserToken != nil {
		if err := user.UserToken.IsValid(); err != nil {
			return err
		}
		// step: use the path or default to the type
		path := "token"
		if user.Path != "" {
			path = user.Path
		}
		uri = fmt.Sprintf("auth/%s/create", path)

		params = &tokenConfig{
			ID: user.UserToken.ID,
			DisplayName: user.UserToken.DisplayName,
			TTL: user.UserToken.TTL.String(),
			MaxUses: user.UserToken.MaxUses,
			Policies: user.Policies,
		}
	}

	log.Debugf("adding the user: %s", params)

	resp, err := r.Request("POST", uri, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("unable to add user: code: %d, body: %s", resp.StatusCode, content)
	}

	return nil
}
