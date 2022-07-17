//5BD16627-D074-45CE-876B-A09AB85CE13B
package checker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"swisscom/config"
	. "swisscom/config"
	"swisscom/pkg/logging"
	"swisscom/pkg/secrets"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

var (
	errNotModified = errors.New("not modified")
	processingchan = make(chan processingItem, 1000)
	copyError      = make(chan error, 0)
)

type processingItem struct {
	AWSsecretVersion string
	params           *config.Config
}

const configFileName = "config/config.yaml"

func HandleSignals(cancel context.CancelFunc) {

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	for {
		sig := <-sigCh
		switch sig {
		case os.Interrupt:
			cancel()
			log.Println("programm stopped")
			return
		}
	}
}

// Scans a secret's version in AWS and if it changed, function for renewing secret wiil be called
func ScanLoop(ctx context.Context, cfg *config.Config, currentVerId *string, secretData map[string]string, kubeconfig *string) {
	for {
		select {
		case <-ctx.Done():
			log.Println("programm stopped")
		case <-time.After(time.Duration(cfg.RescanInterval) * time.Second):
			CheckSecretAWSversion(cfg, currentVerId, secretData, kubeconfig)
		}
		cfgTmp, err := config.ReloadConfig(configFileName)
		if err != nil {
			if err != errNotModified {
				logging.Debug.Println("readconfig:", err)
			}
		} else {
			log.Println("rescanning config file")
			cfg = cfgTmp
			logging.InitLogger(cfg)
		}
	}
}

// Checks if a secret in AWS has changed, if so, then it invokes updating secret in Kubernetes
func CheckSecretAWSversion(cfg *config.Config, currentVerId *string, secretData map[string]string, kubeconfig *string) {

	sess, err := session.NewSession()
	// &aws.Config{
	// Region:      aws.String(region),
	// Credentials: credentials.NewStaticCredentials("AKID", "SECRET_KEY", "TOKEN"),}
	if err != nil {
		// Handle session creation error
		logging.Errorln(err)
	}
	svc := secretsmanager.New(sess,
		aws.NewConfig().WithRegion(cfg.Region))

	inputForVersion := &secretsmanager.ListSecretVersionIdsInput{
		SecretId: aws.String(cfg.AWSSecretName),
	}

	inputForValue := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(cfg.AWSSecretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	resultForValue, err := svc.GetSecretValue(inputForValue)

	resultForVersion, err := svc.ListSecretVersionIds(inputForVersion)

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
			logging.Errorln(err)
		}
	}
	// get actual secret data from AWS
	//================================================================

	var newSecretData string
	if resultForValue.SecretString != nil {
		newSecretData = *resultForValue.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(resultForValue.SecretBinary)))
		_, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, resultForValue.SecretBinary)
		if err != nil {
			fmt.Println("Base64 Decode Error:", err)
		}
	}
	newSecForK8S := make(map[string]string)
	err = json.Unmarshal([]byte(newSecretData), &newSecForK8S)
	if err != nil {
		log.Fatalf("Error unmarshal JSON: %v", err)
	}

	// check current version in AWS
	//================================================================
	log.Println("Checking secret's version in AWS...")
	var actualCurrentVerId string
	var verId string
	// currentVersionId := *currentVerId
	for _, val := range resultForVersion.Versions {
		verId = *val.VersionId
		if *val.VersionStages[0] == "AWSCURRENT" {
			if verId != *currentVerId {
				if cfg.K8sCreateSecret {
					log.Println("Secret's version has changed, applying new version to Kubernetes...")
					secrets.ApplySecretK8s(newSecForK8S, cfg.SecretNamespace, cfg.KubeSecretName, kubeconfig)
				}
				log.Println(cfg.VaultCrearteSecret)
				if cfg.VaultCrearteSecret {
					log.Println("Secret's version has changed, applying new version to Vault...")
					newSecForVault := make(map[string]interface{})
					err = json.Unmarshal([]byte(newSecretData), &newSecForVault)
					if err != nil {
						log.Fatalf("Error unmarshal JSON: %v", err)
					}
					secrets.ApplySecretVault(newSecForVault, cfg)
				}
			}
			actualCurrentVerId = verId

		}
		log.Printf(`Version ID of "AWSCURRENT": %s`, actualCurrentVerId)
	}

	// return secretVersion, err
}

func GetSecretAWS(cfg Config) (string, *string, error) {
	// region string, secretName string, secretVersion string
	//Create a Secrets Manager client
	sess, err := session.NewSession()
	// &aws.Config{
	// Region:      aws.String(region),
	// Credentials: credentials.NewStaticCredentials("AKID", "SECRET_KEY", "TOKEN"),}

	if err != nil {
		// Handle session creation error
		logging.Errorln(err)
	}
	svc := secretsmanager.New(sess,
		aws.NewConfig().WithRegion(cfg.Region))
	inputForValue := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(cfg.AWSSecretName),
		VersionStage: aws.String(cfg.AWSSecretVersion), // VersionStage defaults to AWSCURRENT if unspecified
	}
	log.Println("Retrieving secret from AWS...")
	resultForValue, err := svc.GetSecretValue(inputForValue)

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
			logging.Errorln(err)
		}
	}

	currentVerId := resultForValue.VersionId

	var secretString string
	// decodedBinarySecret string
	if resultForValue.SecretString != nil {
		secretString = *resultForValue.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(resultForValue.SecretBinary)))
		_, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, resultForValue.SecretBinary)
		if err != nil {
			fmt.Println("Base64 Decode Error:", err)
		}
	}
	return secretString, currentVerId, err
}
