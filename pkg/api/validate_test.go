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

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserIsValid(t *testing.T) {
	users := []struct {
		User *User
		Ok   bool
	}{
		{
			User: &User{},
		},
		{
			User: &User{UserPass: &UserPass{
				Username: "test",
			}},
		},
		{
			User: &User{UserPass: &UserPass{
				Username: "test",
				Password: "password",
			}},
			Ok: true,
		},
		{
			User: &User{
				UserPass: &UserPass{Username: "test", Password: "pass"},
				Policies: []string{"pol"},
			},
			Ok: true,
		},
	}

	for i, u := range users {
		err := u.User.IsValid()
		if !u.Ok {
			assert.Error(t, err, "case %d should have errored", i)
		} else {
			assert.NoError(t, err, "case %d should have not errored", i)
		}
	}
}

func TestSecretIsValid(t *testing.T) {
	tests := []struct {
		Secret *Secret
		Ok     bool
	}{
		{
			Secret: &Secret{},
		},
		{
			Secret: &Secret{Path: "/"},
		},
		{
			Secret: &Secret{
				Path:   "/",
				Values: map[string]interface{}{},
			},
		},
		{
			Secret: &Secret{
				Path: "/",
				Values: map[string]interface{}{
					"uri": "/config",
				},
			},
			Ok: true,
		},
	}

	for i, c := range tests {
		err := c.Secret.IsValid()
		if !c.Ok {
			assert.Error(t, err, "case %d should have errored", i)
		} else {
			assert.NoError(t, err, "case %d should have not errored", i)
		}
	}
}

func TestBackendIsValid(t *testing.T) {
	tests := []struct {
		Backend *Backend
		Ok      bool
	}{
		{
			Backend: &Backend{},
		},
		{
			Backend: &Backend{Path: "/"},
		},
		{
			Backend: &Backend{Path: "/", Description: "test"},
		},
		{
			Backend: &Backend{Path: "/", Description: "test", Type: "test"},
		},
		{
			Backend: &Backend{Path: "/", Description: "test", Type: "mysql"},
			Ok:      true,
		},
		{
			Backend: &Backend{
				Path: "/", Description: "test", Type: "mysql",
				Attrs: []*Attributes{
					&Attributes{},
				},
			},
		},
		{
			Backend: &Backend{
				Path: "/", Description: "test", Type: "mysql",
				Attrs: []*Attributes{
					&Attributes{"uri": "test"},
				},
			},
			Ok: true,
		},
	}

	for i, c := range tests {
		err := c.Backend.IsValid()
		if !c.Ok {
			assert.Error(t, err, "case %d should have errored", i)
		} else {
			assert.NoError(t, err, "case %d should have not errored", i)
		}
	}
}

func TestSupportedBackends(t *testing.T) {
	assert.NotEmpty(t, supportedBackends())
}
