package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"path/filepath"
	"swisscom/cmd/checker"
	. "swisscom/config"
	"swisscom/pkg/secrets"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/util/homedir"
)

var (
	errNotModified = errors.New("not modified")
	// cfg            *Config
)

var config_path = flag.String("c", "config/config.yaml", "Path to a config.yaml file")

// var cfg = Config{}
var cfg *Config
var kubeconfig *string

func main() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	content, err := ioutil.ReadFile(*config_path)
	if err != nil {
		log.Fatalf("Problem reading configuration file: %v", err)
	}
	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		log.Fatalf("Error parsing configuration file: %v", err)
	}
	if cfg.K8sCreateSecret == false && cfg.VaultCrearteSecret == false {
		panic("k8s_create_secret and vault_create_secret are both equal false, you need to enable at least one")
	}
	secretFromAWS, currentVerId, err := checker.GetSecretAWS(*cfg)
	if err != nil {
		log.Fatalf("Error getting secret from AWS: %v", err)
	}
	secForK8S := make(map[string]string)
	err = json.Unmarshal([]byte(secretFromAWS), &secForK8S)
	if err != nil {
		log.Fatalf("Error unmarshal JSON: %v", err)
	}

	secForVault := make(map[string]interface{})
	err = json.Unmarshal([]byte(secretFromAWS), &secForVault)
	if err != nil {
		log.Fatalf("Error unmarshal JSON: %v", err)
	}
	if cfg.VaultCrearteSecret {
		secrets.CreateVaultSecret(secForVault, cfg)
	}
	if cfg.K8sCreateSecret {
		secrets.CreateSecretK8s(secForK8S, cfg.SecretNamespace, cfg.KubeSecretName, kubeconfig)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go checker.ScanLoop(ctx, cfg, currentVerId, secForK8S, kubeconfig)
	checker.HandleSignals(cancel)
}
