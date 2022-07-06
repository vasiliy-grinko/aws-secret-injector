package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	vault "github.com/hashicorp/vault/api"
	"gopkg.in/yaml.v2"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	Region             string `yaml:"region"`
	AWSSecretName      string `yaml:"aws_secret_name"`
	KubeSecretName     string `yaml:"kube_secret_name"`
	AWSSecretVersion   string `yaml:"aws_secret_version"`
	SecretNamespace    string `yaml:"kube_secret_namespace"`
	VaultCrearteSecret bool   `yaml:"vault_create_secret"`
	VaultHost          string `yaml:"vault_host"`
	VaultMountPath     string `yaml:"vault_mount_path"`
	VaultDirName       string `yaml:"vault_dir_name"`
	VaultToken         string `yaml:"vault_token"`
}

var config_path = flag.String("c", "config.yaml", "Path to a config.yaml file")
var cfg = Config{}
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
	secretFromAWS, err := getSecretAWS(cfg.Region, cfg.AWSSecretName, cfg.AWSSecretVersion)
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
	if cfg.VaultCrearteSecret == true {
		createVaultSecret(secForVault, cfg.VaultMountPath, cfg.VaultDirName, cfg.VaultHost, cfg.VaultToken)
	} else {
		createSecretK8s(secForK8S, cfg.SecretNamespace, cfg.KubeSecretName)
	}
}

func getSecretAWS(region string, secretName string, secretVersion string) (string, error) {

	//Create a Secrets Manager client
	sess, err := session.NewSession()
	// &aws.Config{
	// Region:      aws.String(region),
	// Credentials: credentials.NewStaticCredentials("AKID", "SECRET_KEY", "TOKEN"),}

	if err != nil {
		// Handle session creation error
		panic(err.Error())
	}
	svc := secretsmanager.New(sess,
		aws.NewConfig().WithRegion(region))
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String(secretVersion), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				fmt.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			}
		} else {
			panic(err.Error())
		}
	}
	var secretString string
	// decodedBinarySecret string
	if result.SecretString != nil {
		secretString = *result.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		_, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			fmt.Println("Base64 Decode Error:", err)
		}
	}
	return secretString, err
}

func createVaultSecret(secretFromAWS map[string]interface{}, vaultMountPath string, vaultDirName string, vaultHost string, vaultToken string) {
	config := vault.DefaultConfig()

	config.Address = vaultHost

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("unable to initialize Vault client: %v", err)
	}
	// Authenticate
	client.SetToken(vaultToken)
	// Write a secret

	_, err = client.KVv2(vaultMountPath).Put(context.Background(), vaultDirName, secretFromAWS)
	if err != nil {
		log.Fatalf("unable to write secret: %v", err)
	}
	fmt.Println("Secret successfully written in Vault.")
}

func createSecretK8s(secretData map[string]string, namespaceForSecret string, secretName string) {

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
	fmt.Println("Creating secret in Kubernetes...")
	result, err := secretClient.Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Secret created:  %q.\n", result.GetObjectMeta().GetName())
}
