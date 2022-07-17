package config

import (
	"errors"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	configModtime  int64
	errNotModified = errors.New("not modified")
)

type Config struct {
	Region             string `yaml:"region"`
	AWSSecretName      string `yaml:"aws_secret_name"`
	K8sCreateSecret    bool   `yaml:"k8s_create_secret"`
	KubeSecretName     string `yaml:"kube_secret_name"`
	AWSSecretVersion   string `yaml:"aws_secret_version"`
	SecretNamespace    string `yaml:"kube_secret_namespace"`
	VaultCrearteSecret bool   `yaml:"vault_create_secret"`
	VaultHost          string `yaml:"vault_host"`
	VaultMountPath     string `yaml:"vault_mount_path"`
	VaultDirName       string `yaml:"vault_dir_name"`
	VaultToken         string `yaml:"vault_token"`
	RescanInterval     int    `yaml:"inerval_rescan_for_renew"`
}

func readConfig(ConfigName string) (x *Config, err error) {
	var file []byte
	if file, err = ioutil.ReadFile(ConfigName); err != nil {
		return nil, err
	}
	x = new(Config)
	if err = yaml.Unmarshal(file, x); err != nil {
		return nil, err
	}
	// if x.LogLevel == "" {
	// 	x.LogLevel = "Debug"
	// }
	return x, nil
}

func ReloadConfig(configName string) (cfg *Config, err error) {
	info, err := os.Stat(configName)
	if err != nil {
		return nil, err
	}
	if configModtime != info.ModTime().UnixNano() {
		configModtime = info.ModTime().UnixNano()
		cfg, err = readConfig(configName)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}
	return nil, errNotModified
}
