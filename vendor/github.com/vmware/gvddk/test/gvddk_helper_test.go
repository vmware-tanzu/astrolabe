package main

import (
	"github.com/vmware/gvddk/gDiskLib"
	"os"
	"strings"
	"testing"
)

func TestGetThumbPrintForServer(t *testing.T) {
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	vmwareThumbprint := os.Getenv("THUMBPRINT")
	if vmwareThumbprint == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	thumbprint, err := gDiskLib.GetThumbPrintForServer(host, port)
	if err != nil {
		t.Errorf("Thumbprint for %s:%s failed, err = %s\n", host, port, err)
	}
	t.Logf("Thumbprint for %s:%s is %s\n", host, port, thumbprint)
	if strings.Compare(vmwareThumbprint, thumbprint) != 0 {
		t.Errorf("Thumbprint %s does not match expected thumbprint %s for %s - check to see if cert has been updated at %s\n",
			thumbprint, vmwareThumbprint, host, host)
	}
}
