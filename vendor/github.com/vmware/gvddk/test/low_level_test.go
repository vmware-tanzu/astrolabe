package main

import (
	"github.com/vmware/gvddk/gDiskLib"
	"os"
	"testing"
)

func TestCreate(t *testing.T) {
	// Set up
	path := os.Getenv("LIBPATH")
	if path == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	res := gDiskLib.Init(7, 0, path)
	if res != nil {
		t.Errorf("Init failed, got error code: %d, error message: %s.", res.VixErrorCode(), res.Error())
	}
	serverName := os.Getenv("IP")
	thumPrint := os.Getenv("THUMBPRINT")
	userName := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	fcdId := os.Getenv("FCDID")
	ds := os.Getenv("DATASTORE")
	identity := os.Getenv("IDENTITY")
	params := gDiskLib.NewConnectParams("", serverName,thumPrint, userName,
		password, fcdId, ds, "", "", identity, "", gDiskLib.VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ,
		false, gDiskLib.NBD)
	err1 := gDiskLib.PrepareForAccess(params)
	if err1 != nil {
		t.Errorf("Prepare for access failed. Error code: %d. Error message: %s.", err1.VixErrorCode(), err1.Error())
	}
	conn, err2 := gDiskLib.ConnectEx(params)
	if err2 != nil {
		gDiskLib.EndAccess(params)
		t.Errorf("Connect to vixdisk lib failed. Error code: %d. Error message: %s.", err2.VixErrorCode(), err2.Error())
	}
	gDiskLib.Disconnect(conn)
	gDiskLib.EndAccess(params)
}