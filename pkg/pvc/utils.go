package pvc

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	// ClaimBindingTimeout is how long claims have to become bound.
	ClaimBindingTimeout = 3 * time.Minute

	// Poll is how often to Poll an API object.
	Poll = 2 * time.Second
)

// WaitForPersistentVolumeClaimPhase waits for a PersistentVolumeClaim to be in a specific phase or until timeout occurs, whichever comes first.
func WaitForPersistentVolumeClaimPhase(ctx context.Context, phase v1.PersistentVolumeClaimPhase, c clientset.Interface, ns string, pvcName string, Poll, timeout time.Duration, logger logrus.FieldLogger) error {
	return WaitForPersistentVolumeClaimsPhase(ctx, phase, c, ns, []string{pvcName}, Poll, timeout, true, logger)
}

// WaitForPersistentVolumeClaimsPhase waits for any (if matchAny is true) or all (if matchAny is false) PersistentVolumeClaims
// to be in a specific phase or until timeout occurs, whichever comes first.
func WaitForPersistentVolumeClaimsPhase(ctx context.Context, phase v1.PersistentVolumeClaimPhase, c clientset.Interface, ns string, pvcNames []string, Poll, timeout time.Duration, matchAny bool, logger logrus.FieldLogger) error {
	if len(pvcNames) == 0 {
		return fmt.Errorf("incorrect parameter: Need at least one PVC to track. Found 0")
	}
	logger.Infof("Waiting up to %v for PersistentVolumeClaims %v to have phase %s", timeout, pvcNames, phase)
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(Poll) {
		phaseFoundInAllClaims := true
		for _, pvcName := range pvcNames {
			pvc, err := c.CoreV1().PersistentVolumeClaims(ns).Get(ctx, pvcName, metav1.GetOptions{})
			if err != nil {
				logger.Errorf("Failed to get claim %q, retrying in %v. Error: %v", pvcName, Poll, err)
				continue
			}
			if pvc.Status.Phase == phase {
				logger.Infof("PersistentVolumeClaim %s found and phase=%s (%v)", pvcName, phase, time.Since(start))
				if matchAny {
					return nil
				}
			} else {
				logger.Infof("PersistentVolumeClaim %s found but phase is %s instead of %s.", pvcName, pvc.Status.Phase, phase)
				phaseFoundInAllClaims = false
			}
		}
		if phaseFoundInAllClaims {
			return nil
		}
	}
	return fmt.Errorf("PersistentVolumeClaims %v not all in phase %s within %v", pvcNames, phase, timeout)
}
