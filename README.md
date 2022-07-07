 aws-secret-injector 
## What is aws-secret-injector?
A programm that retrievs secrets from AWS SecretManager and puts them in Kubernetes as a secret resource. Alternatively it can put a secret in Hashicorp Vault instead of secret in Kuvernetes.

---

## Installation

* Via a GO install
  ```shell
  go get -u github.com/Mr-Mark2112/aws-secret-injector
  ```
---

## Building From Source

 In order to build aws-secret-injector from source you must:

 1. Clone the repo
 2. Build and run the executable

      ```shell
      make build && ./execs/awssec-inj
      ```
---
## How to use it?
* 1. Create the config.yaml with parameters (shown below) or you can use the config.yaml from this repo -  you can specify the path to the config.yaml file by flag `-c`.
* 2. Choose where you want to save secret from AWS: in Kuberenetes secret or in Hashicorp Vault.
* 3. Make changes in parameters (required shown below)
* 4. Make sure you have configured access to AWS account in ~/.aws
* 5. (If you chose to create secret in k8s, then you need to make sure that EKS (Kubernetes) cluster is reachable - you can specify the path to your kubeconfig file by flag `-kubeconfig`.

## Parameters

| Parameter                                                      | required                         | Comment                                                                |
|----------------------------------------------------------------|----------------------------------|------------------------------------------------------------------------|
| region                                                         | `+`                              | AWS region name                                                        |
| aws_secret_name                                                | `+`                              | A name of a secret in AWS SecretManager                                |
| aws_secret_version                                             | `+`                              | default `AWSCURRENT`                                                   |
| vault_create_secret                                            | `+`                              | If `true` a secret will be created in Vault enstead of k8s             |
| vault_host                                                     | `-` if vault_create_secret true  | Vault host URL+PORT                                                    |
| vault_token                                                    | `-` if vault_create_secret true  | Vault authorization token                                              |
| vault_mount_path                                               | `-` if vault_create_secret true  | Vault mount path of Kv engine                                          |
| vault_dir_name                                                 | `-` if vault_create_secret true  | A name of a directory, where the secret variables will be saved        |
| kube_secret_namespace                                          | `-` if vault_create_secret false | A name of a Kubernetes namespace, where a secret will be created       |
| kube_secret_name                                               | `-` if vault_create_secret true  | A name of a Kubernetes secret which will be created                    |

