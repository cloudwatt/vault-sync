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
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
)

// Encrypt encrypts the content with a transit endpoint
func (r *Client) Encrypt(path, key string, reader io.Reader) (string, error) {
	// step: build the path
	endpoint := fmt.Sprintf("%s/encrypt/%s", path, key)
	// step: read the contents of the file
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	// step: base64 the content of the file
	encoded := base64.StdEncoding.EncodeToString(content)
	// step: encrypt the content
	resp, err := r.client.Logical().Write(endpoint, map[string]interface{}{
		"plaintext": encoded,
	})
	if err != nil {
		return "", err
	}
	// step: get the ciphertext from response
	cipher, found := resp.Data["ciphertext"]
	if !found {
		return "", fmt.Errorf("ciphertext not found")
	}

	return fmt.Sprintf("%s", cipher), nil
}

// Decrypt decrypt the content with a transit endpoint
func (r *Client) Decrypt(path, key string, reader io.Reader) (string, error) {
	// step: build the path
	endpoint := fmt.Sprintf("%s/decrypt/%s", path, key)
	// step: read the contents
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	// step: encrypt the content
	resp, err := r.client.Logical().Write(endpoint, map[string]interface{}{
		"ciphertext": fmt.Sprintf("%s", content),
	})
	if err != nil {
		return "", err
	}
	// step: extract the plaintext
	cipher, found := resp.Data["plaintext"]
	if !found {
		return "", err
	}
	// step: decode the base64
	decoded, err := base64.StdEncoding.DecodeString(cipher.(string))
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}
