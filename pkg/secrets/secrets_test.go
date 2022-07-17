package secrets

import (
	"flag"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/util/homedir"
)

type envs struct {
	number   int
	reverted int
}

func TestCreateSecretK8s(t *testing.T) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		// filepath.Join(home, ".kube", "config")
		kubeconfig = flag.String("kubeconfig", "/home/sylar/.kube/k3s.yaml", "(optional) absolute path to the kubeconfig file")
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
