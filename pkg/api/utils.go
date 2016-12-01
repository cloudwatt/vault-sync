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
	"fmt"
	"strings"
)

// String returns a string representation of the backend
func (r Backend) String() string {
	return fmt.Sprintf("path: %s, type: %s", r.Path, r.Type)
}

// GetDefaultTTL returns the default ttl
func (r Backend) GetDefaultTTL() string {
	if r.DefaultLeaseTTL <= 0 {
		return "system"
	}

	return r.DefaultLeaseTTL.String()
}

// GetPath returns the backend path
func (r Backend) GetPath() string {
	return fmt.Sprintf("%s", strings.TrimPrefix(strings.TrimSuffix(r.Path, "/"), "/"))
}

// GetMaxTTL returns the max ttl
func (r Backend) GetMaxTTL() string {
	if r.MaxLeaseTTL <= 0 {
		return "system"
	}

	return r.MaxLeaseTTL.String()
}

// URI returns the uri for the config item
func (r *Attributes) URI() string {
	x, _ := (*r)["uri"]
	return x.(string)
}

// IsOneshot checks if the attribute is a oneshot attribute
func (r *Attributes) IsOneshot() bool {
	_, found := (*r)["oneshot"]
	return found
}

// Values retrieves the values from the attributes
func (r *Attributes) Values() map[string]interface{} {
	return (*r)
}

// GetPath returns the uri of the config
func (r *Attributes) GetPath(ns string) string {
	return fmt.Sprintf("%s/%s", ns, r.URI())
}

func (r *Attributes) String() string {
	var items []string
	for k, v := range *r {
		items = append(items, fmt.Sprintf("[%s|%s]", k, v))
	}

	return strings.Join(items, ",")
}

func (r User) Username() string {
	if r.UserPass != nil {
		return r.UserPass.Username
	}
	if r.UserToken != nil {
		return r.UserToken.DisplayName
	}

	return ""
}

// GetPolicies returns the policies associated to a user
func (r User) GetPolicies() string {
	if len(r.Policies) <= 0 {
		return "none"
	}

	return strings.Join(r.Policies, ",")
}
