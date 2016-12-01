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

import "time"

// Attributes is a map of configuration
type Attributes map[string]interface{}

// Config is the definition for a config file
type Config struct {
	// Users is a series of users
	Users []*User `yaml:"users" json:"users" hcl:"users"`
	// Backends is a series of backend's
	Backends []*Backend `yaml:"backends" json:"backends" hcl:"backends"`
	// Secrets is a series of secrets
	Secrets []*Secret `yaml:"secrets" json:"secrets" hcl:"secrets"`
	// Auths is a series of authentication backend's
	Auths []*Auth `yaml:"auths" json:"auths" hcl:"auths"`
}

// Auth defined a authentication backend
type Auth struct {
	// Path is the path of the authentication backend
	Path string `yaml:"path" json:"path" hcl:"path"`
	// Type is the authentication type
	Type string `yaml:"type" json:"type" hcl:"type"`
	// Description is the a description for the backend
	Description string `yaml:"description" json:"description" hcl:"description"`
	// Attributes is a map of configurations for the backend
	Attrs []*Attributes `yaml:"attributes" json:"attributes" hcl:"attributes"`
}

// Backend defined the type and configuration for a backend in vault
type Backend struct {
	// Path is the mountpoint for the mount
	Path string `yaml:"path" json:"path" hcl:"path"`
	// Description is the a description for the backend
	Description string `yaml:"description" json:"description" hcl:"description"`
	// Type is the type of backend
	Type string `yaml:"type" json:"type" hcl:"type"`
	// DefaultLeaseTTL is the default lease of the backend
	DefaultLeaseTTL time.Duration `yaml:"default-lease-ttl" json:"default-lease-ttl" hcl:"default-lease-ttl"`
	// MaxLeaseTTL is the max ttl
	MaxLeaseTTL time.Duration `yaml:"max-lease-ttl" json:"max-lease-ttl" hcl:"max-lease-ttl"`
	// Attrs is the configuration of the mount point
	Attrs []*Attributes `yaml:"attributes" json:"attributes" hcl:"attributes"`
}

// Secret defines a secret
type Secret struct {
	// Path is key for this secret
	Path string `yaml:"path" json:"path" hcl:"path"`
	// Values is a series of values associated to the secret
	Values map[string]interface{} `yaml:"values" json:"values" hcl:"values"`
}

// User is the definition for a user
type User struct {
	// Path is the authentication path for the user
	Path string `yaml:"path" json:"path" hcl:"path"`
	// UserPass is the credentials for a userpass auth backend
	UserPass *UserPass `yaml:"userpass" json:"userpass" hcl:"userpass"`
	// UserToken is a token struct for this user
	UserToken *UserToken `yaml:"usertoken" json:"usertoken" hcl:"usertoken"`
	// Policies is a list of policies the user has access to
	Policies []string `yaml:"policies" json:"policies" hcl:"policies"`
	// Namespace is optional and used when adding to kubernetes
	Namespace string `yaml:"namespace" json:"namespace" hcl:"namespace"`
}

// UserCredentials are the userpass credentials
type UserPass struct {
	// Username is the id of the user
	Username string `yaml:"username" json:"username" hcl:"username"`
	// Password is the password of the user
	Password string `yaml:"password" json:"password" hcl:"password"`
}

// UserToken is the token
type UserToken struct {
	// ID is the actual token itselg
	ID string `yaml:"id" json:"id" hcl:"id"`
	// TTL is the time duration of the token
	TTL time.Duration `yaml:"ttl" json:"ttp" hcl:"ttl"`
	// DisplayName is a generic name for the token
	DisplayName string `yaml:"display-name" json:"display-name" hcl:"display-name"`
	// MaxUses is the max number of times the token can be used
	MaxUses int `yaml:"max-uses" json:"max-uses" hcl:"max-uses"`
}
