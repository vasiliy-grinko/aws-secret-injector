package secrets

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	. "swisscom/config"
	"testing"

	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/util/homedir"
)

type envs struct {
	number   int
	reverted int
}

func TestCreateSecretK8s(t *testing.T) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	testSecret := map[string]string{
		"password": "a password",
		"dbname":   "a database name",
		"engine":   "mariadb",
		"port":     "3306",
		"host":     "a host name",
		"username": "a username",
	}
	data := CreateSecretK8s(testSecret, "default", "test-secret", kubeconfig)
	createdSecred := make(map[string]string)
	for k, v := range data {
		createdSecred[k] = string(v)
		assert.Equal(t, createdSecred[k], testSecret[k], "Must be equal")
	}
	fmt.Println(createdSecred)
}

func TestCreateSecretVault(t *testing.T) {
	var cfg *Config
	configPath := "/home/sylar/swisscom/config/config.yaml"
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Problem reading configuration file: %v", err)
	}
	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		log.Fatalf("Error parsing configuration file: %v", err)
	}
	testSecret := map[string]interface{}{
		"password": "a password",
		"dbname":   "a database name",
		"engine":   "mariadb",
		"port":     "3306",
		"host":     "a host name",
		"username": "a username",
	}
	CreateVaultSecret(testSecret, cfg)

	config := vault.DefaultConfig()

	config.Address = cfg.VaultHost

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}
	// Authenticate
	client.SetToken(cfg.VaultToken)

	vaultData, err := client.KVv2(cfg.VaultMountPath).Get(context.Background(), cfg.VaultDirName)
	createdSecred := make(map[string]string)
	for k, v := range vaultData.Data {
		str, _ := v.(string)
		fmt.Println(k, str)
		createdSecred[k] = str
		assert.Equal(t, createdSecred[k], testSecret[k], "Must be equal")
	}
	fmt.Println(createdSecred)
}
