 aws-secret-injector 
## What is aws-secret-injector?
A tool for syncing secrets from AWS SecretManager to Vault/Kubernetes. 

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
*  Create the config.yaml with parameters (shown below) or you can use the config.yaml from this repo -  you can specify the path to the config.yaml file by flag `-c`.
*  Choose where you want to save secret from AWS: in Kuberenetes secret or in Hashicorp Vault.
*  Make changes in parameters (required shown below)
*  Make sure you have configured access to AWS account in ~/.aws
*  (If you chose to create secret in k8s, then you need to make sure that EKS (Kubernetes) cluster is reachable - you can specify the path to your kubeconfig file by flag `-kubeconfig`.
*  You can make changes in the config.yaml while the program is running - it will reload config every ```inerval_rescan_for_renew``` 

## Parameters

| Parameter                                                      | required                         | Comment                                                                |
|----------------------------------------------------------------|----------------------------------|------------------------------------------------------------------------|
| region                                                         | `+`                              | AWS region name                                                        |
| aws_secret_name                                                | `+`                              | A name of a secret in AWS SecretManager                                |
| aws_secret_version                                             | `+`                              | default `AWSCURRENT`                                                   |
| inerval_rescan_for_renew                                       | `+`                              | Time interval in seconds for check changes in a secret in AWS          |
| vault_create_secret                                            | `+`                              | If `true`, a secret will be created in Vault                           |
| vault_host                                                     | `-` if vault_create_secret true  | Vault host URL+PORT                                                    |
| vault_token                                                    | `-` if vault_create_secret true  | Vault authorization token                                              |
| vault_mount_path                                               | `-` if vault_create_secret true  | Vault mount path of Kv engine                                          |
| vault_dir_name                                                 | `-` if vault_create_secret true  | A name of a directory, where the secret variables will be saved        |
| kube_create_secret                                             | `+`                              | If `true`, a secret will be created in Kuvernetes                      |
| kube_secret_namespace                                          | `-`                              | A name of a Kubernetes namespace, where a secret will be created       |
| kube_secret_name                                               | `-`                              | A name of a Kubernetes secret which will be created                    |

