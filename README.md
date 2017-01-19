### **Vault-sync**

[![Build Status](https://travis-ci.org/cloudwatt/vault-sync.svg?branch=master)](https://travis-ci.org/cloudwatt/vault-sync)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudwatt/vault-sync)](https://goreportcard.com/report/github.com/cloudwatt/vault-sync)

---
Disclaimer : this started as a fork of the [vaultctl](https://github.com/UKHomeOffice/vaultctl)

Vault-sync is a command line utilty for provisioning a Hashicorp's [Vault](https://www.vaultproject.io) from configuration files. Essentially it was written so we could source control our users, policies, backends and secrets, synchronize the vault against them and rebuild on-demand if required.

##### **Build**
---

 There is a Makefile in the root directory, so a simply `make` will build the project. Alternatively you can run the build inside a docker via `make docker-build`

##### **Usage**
---

An example for most supported configurations can be found in the `tests` directory.

To sync the directory you can run for instance :
`vault-sync -A https://vault.mydomain:8200 -t $VAULT_TOKEN sync --sync-full -c tests/config.yaml -p tests/policies`

This will sync all config from `tests/config.yaml` and all policies in `tests/policies` that are in format `*.hcl`

```shell
NAME:
   vault-sync - is a utility for provisioning a hashicorp's vault service

USAGE:
   vault-sync [global options] command [command options] [arguments...]

VERSION:
   v0.1.0-alpha1

AUTHOR:
   Félix Cantournet <felix.cantournet@gmail.com>

COMMANDS:
     synchronize, sync   synchonrize the users, policies, secrets and backends
     transit, tr, trans  Encrypts / decrypts files using the Vault transit backend
     help, h             Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -A value, --vault-addr value      the url address of the vault service (default: "http://127.0.0.1:8200") [$VAULT_ADDR]
   -u value, --vault-username value  the vault username to use to authenticate to vault service [$VAULT_USERNAME]
   -p value, --vault-password value  the vault password to use to authenticate to vault service [$VAULT_PASSWORD]
   -t value, --vault-token value     a vault token used to authenticate to vault service [$VAULT_TOKEN]
   -c value, --credentials value     the path to a file (json|yaml) containing the username and password for userpass authenticaion [$VAULT_CRENDENTIALS]
   --verbose                         switch on verbose logging for debug purposed
   --help, -h                        show help
   --version, -v                     print the version
```

##### **Configuration**

The configuration files for vault-sync can be written in json or yml format *(note, it check the file extension to determine the format)*. You can specify multiple configuration files and or multiple directories containing config files.

###### - **Authentication**

Authentication backends can be created using the following

```YAML
auths:
- path: userpass
  type: userpass
- path: some/path/users
  type: userpass
- path: github
  type: github
  attributes:
  - uri: config
    organization: SomeOrganization
```

###### - **Users**

Users are place in a users: [] collection, the vault authentication type *(at present only userpass is supported, though it would be trivial to add more)* followed by the policies associated to the user

```YAML
users:
- userpass:
    username: rohithj
    password: password1
  policies:
    - common
    - platform_tls
```

###### - **Backends**

The backends are defined under the 'backends[]' collection, each backend must have a path *(i.e. a mount point)*, a type which is the Vault backend type, a description *(which is enforced)* and an optional collection of config items. Keeping it simple the config[] is essentially a series of PUT requests. You can grab the configuration options and the uri from the Vault documentation. Note. an extra option *'oneshot'* been added, it simply means the config option will ONLY is run the first time the backend is created, which is useful for some backends like PKI, transit etc.

```YAML
backends:
- type: transit
  path: platform/encode
  description: A transit backend used to encrypt configuration files
  attributes:
  - uri: keys/default
    oneshot: true
- type: generic
  path: platform/secrets
  description: platform secrets
- path: platform/platform_tls
  description: platform tls
  type: generic
- path: platform/pki
  type: pki
  description: Platform PKI backend
  attributes:
  - uri: root/generate/internal
    common_name: example.com
    ttl: 3h
    oneshot: true
  - uri: roles/example-dot-com
    allowed_domains: example.com
    allow_subdomains: true
    max_ttl: 1h
# one of the annoying things about the mysql backend is it attempts to connect to the db when
# adding the config/connection config??
- path: platform/db
  type: mysql
  description: Platform Database
  attributes:
  - uri: config/connection
    value: root:root@tcp(127.0.0.1:3306)/
    oneshot: true
  - uri: roles/readonly
    sql: CREATE USER '{{name}}'@'%' IDENTIFIED BY '{{password}}';GRANT SELECT ON *.* TO '{{name}}'@'%'
```

###### - **Secrets**

```YAML
secrets:
  - path: platform/secrets/platform_tls
    values:
      hello: world
      rohith: yes
  - path: platform/secrets/se1
    values:
      hello: world
      rohith: yes
```

###### - **Example Output**

```shell
[jest@starfury vault-sync]$ bin/vault-sync -u admin -p password  sync -p tests/policies -c platform.yml
INFO[0000] -> synchronizing the vault policies, 3 files
INFO[0001] [policy: common.hcl] successfully applied the policy, filename: tests/policies/common.hcl
INFO[0001] [policy: platform.hcl] successfully applied the policy, filename: tests/policies/platform.hcl
INFO[0001] [policy: platform_tls.hcl] successfully applied the policy, filename: tests/policies/platform_tls.hcl
INFO[0001] -> synchronizing the vault users, users: 1
INFO[0001] [user: rohithj] ensuring user, policies: root
INFO[0001] -> synchronizing the backends, backend: 2
INFO[0001] [backend: platform/encode]: already exist, moving to configuration
INFO[0001] [backend:platform/encode/keys/default] skipping the config, as it's a oneshot setting
INFO[0001] [backend: platform/secrets]: already exist, moving to configuration
INFO[0001] -> synchronizing the secrets with vault, secrets: 0
INFO[0001] synchronization complete, time took: 1.733908869s
```


#### **Transit Encryption**
---
The sub-command 'transit' permits you to encrypt and decrypt the file contents using a [Vault transit](https://www.vaultproject.io/docs/secrets/transit/index.html) backend. The current use case being we hand off management to others to manage their our namespaces, secret, backends etc and behold a generic endpoint for encryption.

##### **TODO**
---

- More tests
- Keep up with Vault.
- Documentation
