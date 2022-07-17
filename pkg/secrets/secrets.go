package secrets

import (
	"context"
	"fmt"
	"log"
	"strings"
	"swisscom/config"
	"swisscom/pkg/logging"

	vault "github.com/hashicorp/vault/api"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apply "k8s.io/client-go/applyconfigurations/core/v1"
	applyConfV1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func CreateVaultSecret(secret map[string]interface{}, cfg *config.Config) {
	config := vault.DefaultConfig()

	config.Address = cfg.VaultHost

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}
	// Authenticate
	client.SetToken(cfg.VaultToken)
	// Write a secret

	_, err = client.KVv2(cfg.VaultMountPath).Put(context.Background(), cfg.VaultDirName, secret)
	if err != nil {
		log.Fatalf("unable to write secret: %v", err)
	}
	log.Println("Secret successfully written in Vault.")
}

func ApplySecretVault(secret map[string]interface{}, cfg *config.Config) {
	// vaultMountPath string, vaultDirName string, vaultHost string, vaultToken string
	config := vault.DefaultConfig()

	config.Address = cfg.VaultHost

	client, err := vault.NewClient(config)
	if err != nil {
		logging.Errorln("unable to initialize Vault client: %v", err)
	}
	// Authenticate
	client.SetToken(cfg.VaultToken)
	// Write a secret

	_, err = client.KVv2(cfg.VaultMountPath).Patch(context.Background(), cfg.VaultDirName, secret)
	if err != nil {
		logging.Errorln("unable to patch secret: %v", err)
	}
	log.Println("Secret successfully patched in Vault.")
}

func CreateSecretK8s(secretData map[string]string, namespaceForSecret string, secretName string, kubeconfig *string) map[string][]byte {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	secretClient := clientset.CoreV1().Secrets(namespaceForSecret)
	data := make(map[string][]byte)

	for key, value := range secretData {
		data[key] = []byte(value)
	}

	object := metav1.ObjectMeta{Name: secretName}
	secret := &apiv1.Secret{Data: data, ObjectMeta: object}
	log.Println("Creating secret in Kubernetes...")
	result, err := secretClient.Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Println("Secret alredy exists, applying ...")
			ApplySecretK8s(secretData, namespaceForSecret, secretName, kubeconfig)
		} else {
			panic(err)
		}
	}
	if err == nil {
		log.Printf("Secret created:  %q.\n", result.GetObjectMeta().GetName())
	}
	return result.Data
}

func ApplySecretK8s(secretData map[string]string, namespaceForSecret string, secretName string, kubeconfig *string) {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		logging.Errorln(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logging.Errorln(err)
	}

	secretClient := clientset.CoreV1().Secrets(namespaceForSecret)
	data := make(map[string][]byte)
	for key, value := range secretData {
		data[key] = []byte(value)
	}
	kind := "Secret"
	apiVerStr := "v1"
	object := &applyConfV1.ObjectMetaApplyConfiguration{Name: &secretName, Namespace: &namespaceForSecret}

	apiVersion := &applyConfV1.TypeMetaApplyConfiguration{Kind: &kind, APIVersion: &apiVerStr}

	secret := &apply.SecretApplyConfiguration{StringData: secretData, ObjectMetaApplyConfiguration: object, TypeMetaApplyConfiguration: *apiVersion}

	result, err := secretClient.Apply(context.TODO(), secret, metav1.ApplyOptions{FieldManager: "application/apply-patch"})

	if err != nil {
		panic(err)
	}
	fmt.Printf("Secret applied: %q.\n", result.GetObjectMeta().GetName())
}
