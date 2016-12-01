### **Vaultctl**

---
Vaultctl is a command line utilty for provisioning a Hashicorp's [Vault](https://www.vaultproject.io) from configuration files. Essentially it was written so we could source control our users, policies, backends and secrets, synchronize the vault against them and rebuild on-demand if required.
 
##### **Build**
---
 
 There is a Makefile in the root directory, so a simply ***make*** will build the project. Alternatively you can run the build inside a docker via ***make docker-build***
 
##### **Usage**
---
 
```shell
[jest@starfury vaultctl]$ bin/vaultctl --help
NAME:
   vaultctl - is a utility for provisioning a hashicorp's vault service

USAGE:
   vaultctl [global options] command [command options] [arguments...]
   
VERSION:
   v0.0.1
   
AUTHOR(S):
   Rohith <gambol99@gmail.com> 
   
COMMANDS:
   synchronize, sync	synchonrize the users, policies, secrets and backends
   transit, tr, trans	Encrypts / decrypts files using the Vault transit backend
   help, h		Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   -A, --vault-addr "http://127.0.0.1:8200"	the url address of the vault service [$VAULT_ADDR]
   -u, --vault-username 			the vault username to use to authenticate to vault service [$VAULT_USERNAME]
   -p, --vault-password 			the vault password to use to authenticate to vault service [$VAULT_PASSWORD]
   -c, --credentials 				the path to a file (json|yaml) containing the username and password for userpass authenticaion [$VAULT_CRENDENTIALS]
   --verbose					switch on verbose logging for debug purposed
   --kube-populate				whether or not to populate the vault crendentials into the namespaces
   --help, -h					show help
   --version, -v				print the version
``` 

##### **Configuration**

The configuration files for vaultctl can be written in json or yml format *(note, it check the file extension to determine the format)*. You can specify multiple configuration files and or multiple directories containing config files. 

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
[jest@starfury vaultctl]$ bin/vaultctl -u admin -p password  sync -p tests/policies -c platform.yml
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

- Need to finish off the Kubernetes intregetion to place the vault credentials in k8s secrets.
