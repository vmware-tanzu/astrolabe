package util

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"strings"
)

// TODO: Merge constants from plugin and here
const VddkConfigPath = "/tmp/config"

const (
	VCSecretNs             = "kube-system"
	VCSecretNsSupervisor   = "vmware-system-csi"
	VCSecret               = "vsphere-config-secret"
	VCSecretTKG            = "csi-vsphere-config"
	VCSecretData           = "csi-vsphere.conf"
	VCSecretDataSupervisor = "vsphere-cloud-provider.conf"
)

const (
	TkgSupervisorService = "supervisor"
)

// Indicates the type of cluster where Plugin is installed
type ClusterFlavor string

const (
	Unknown    ClusterFlavor = "Unknown"
	Supervisor               = "Supervisor Cluster"
	TkgGuest                 = "TKG Guest Cluster"
	VSphere                  = "vSphere Kubernetes Cluster"
)

func GetStringFromParamsMap(params map[string]interface{}, key string, logger logrus.FieldLogger) (value string, ok bool) {
	valueIF, ok := params[key]
	if ok {
		value, ok := valueIF.(string)
		if !ok {
			logger.Errorf("Value for params key %s is not a string", key)
		}
		return value, ok
	} else {
		logger.Errorf("No such key %s in params map", key)
		return "", ok
	}
}

func IsConnectionResetError(err error) bool {
	if strings.Contains(err.Error(), "connection reset by peer") {
		return true
	}
	return false
}

func RetrievePlatformInfoFromConfig(config *rest.Config, params map[string]interface{}) error {
	var err error
	if config == nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return errors.Errorf("Failed to get k8s inClusterConfig")
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Errorf("Failed to get k8s clientset from the given config: %v", config)
	}
	// Get the cluster flavor
	var ns string
	clusterFlavor, err := GetClusterFlavor(config)
	if clusterFlavor == TkgGuest || clusterFlavor == Unknown {
		return errors.New("RetrieveVcConfigSecret: Cannot retrieve VC secret")
	} else if clusterFlavor == Supervisor {
		ns = VCSecretNsSupervisor
	} else {
		ns = VCSecretNs
	}

	secretApis := clientset.CoreV1().Secrets(ns)
	vsphere_secrets := []string{VCSecret, VCSecretTKG}
	var secret *v1.Secret
	for _, vsphere_secret := range vsphere_secrets {
		secret, err = secretApis.Get(context.TODO(), vsphere_secret, v12.GetOptions{})
		if err == nil {
			break
		}
	}

	// No valid secret found.
	if err != nil {
		return errors.Errorf("Failed to get k8s secret, %s", vsphere_secrets)
	}

	// No kv pairs in the secret.
	if len(secret.Data) == 0 {
		return errors.Errorf("Failed to get k8s secret, %s", vsphere_secrets)
	}

	var sEnc string
	var lines []string

	for _, value := range secret.Data {
		sEnc = string(value)
		lines = strings.Split(sEnc, "\n")
		// will always have only one kv pair.
		break
	}

	for _, line := range lines {
		if strings.Contains(line, "VirtualCenter") {
			parts := strings.Split(line, "\"")
			params["VirtualCenter"] = parts[1]
		} else if strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Skip the quotes in the value if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				params[key] = value[1 : len(value)-1]
			} else {
				params[key] = value
			}
		}
	}

	// If port is missing, add an entry in the params to use the standard https port
	if _, ok := params["port"]; !ok {
		params["port"] = "443"
	}

	return nil
}

// Check the cluster flavor that the plugin is deployed in
func GetClusterFlavor(config *rest.Config) (ClusterFlavor, error) {
	var err error
	if config == nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return "Unknown", err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return Unknown, err
	}

	// Direct vSphere deployment.
	// Check if vSphere secret is available in appropriate namespace.
	ns := VCSecretNs
	secretApis := clientset.CoreV1().Secrets(ns)
	vsphere_secrets := []string{VCSecret, VCSecretTKG}
	for _, vsphere_secret := range vsphere_secrets {
		_, err := secretApis.Get(context.TODO(), vsphere_secret, v12.GetOptions{})
		if err == nil {
			return VSphere, nil
		}
	}

	// Check if in supervisor.
	// Check if vSphere secret is available in appropriate namespace.
	ns = VCSecretNsSupervisor
	secretApis = clientset.CoreV1().Secrets(ns)
	_, err = secretApis.Get(context.TODO(), VCSecret, v12.GetOptions{})
	if err == nil {
		return Supervisor, nil
	}

	// Check if in guest cluster.
	// Check for the supervisor service in the guest cluster.
	serviceApi := clientset.CoreV1().Services("default")
	_, err = serviceApi.Get(context.TODO(), TkgSupervisorService, v12.GetOptions{})
	if err == nil {
		return TkgGuest, nil
	}

	// Did not match any search criteria. Unknown cluster flavor.
	return Unknown, errors.New("GetClusterFlavor: Failed to identify cluster flavor")
}

func CreateConfigFile(vddkConfig map[string]string, logger logrus.FieldLogger) (string, error) {
	path := VddkConfigPath
	logger.Infof("Customized vddk config: %v", vddkConfig)
	err := createFile(path, vddkConfig, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to create config file")
		return "", err
	}
	return path, nil
}

func DeleteConfigFile(path string, logger logrus.FieldLogger) error {
	return deleteFile(path, logger)
}

func createFile(path string, config map[string]string, logger logrus.FieldLogger) error {
	// check if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for k, v := range config {
		logger.Infof("Writing config %s=%s", k, v)
		// Write some text line-by-line to file.
		_, err = file.WriteString(k + "=" + v + "\n")
		if err != nil {
			return err
		}
	}

	// Save file changes.
	err = file.Sync()
	if err != nil {
		return err
	}

	logger.Debugf("Config File %v Created Successfully", path)
	return nil
}

func deleteFile(path string, logger logrus.FieldLogger) error {
	// delete file
	var err = os.Remove(path)
	if err != nil {
		logger.Errorf("Failed to delete config file %s", path)
		return err
	}

	logger.Debugf("File %s Deleted", path)
	return nil
}