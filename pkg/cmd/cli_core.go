package cmd

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	restClient "github.com/vmware-tanzu/astrolabe/gen/client"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	astrolabeClient "github.com/vmware-tanzu/astrolabe/pkg/client"
	"github.com/vmware-tanzu/astrolabe/pkg/server"
	"io"
	"os"
)

type cliContext struct {
	context.Context
	addonInits map[string]server.InitFunc
}

func CliCore(addonInits map[string]server.InitFunc) {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Usage: "Astrolabe server",
			},
			&cli.StringFlag{
				Name:     "destHost",
				Usage:    "Optional different destination Astrolabe server for cp",
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "insecure",
				Usage:    "Only use HTTP",
				Required: false,
				Hidden:   false,
				Value:    false,
			},
			&cli.StringFlag{
				Name:  "confDir",
				Usage: "Configuration directory",
			},
			&cli.StringFlag{
				Name:     "destConfDir",
				Usage:    "Optional different destination Astrolabe configuration directory for cp",
				Required: false,
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "types",
				Usage:  "shows Protected Entity Types",
				Action: types,
			},
			{
				Name:   "ls",
				Usage:  "lists entities for a type",
				Action: ls,
			},
			{
				Name:      "show",
				Usage:     "shows info for a protected entity",
				ArgsUsage: "<protected entity id>",
				Action:    show,
			},
			{
				Name:      "lssn",
				Usage:     "lists snapshots for a Protected Entity",
				Action:    lssn,
				ArgsUsage: "<protected entity id>",
			},
			{
				Name:      "snap",
				Usage:     "snapshots a Protected Entity",
				Action:    snap,
				ArgsUsage: "<protected entity id>",
			},
			{
				Name:      "rmsn",
				Usage:     "removes a Protected Entity snapshot",
				Action:    rmsn,
				ArgsUsage: "<protected entity snapshot id>",
			},
			{
				Name:      "cp",
				Usage:     "copies a Protected Entity snapshot",
				Action:    cp,
				ArgsUsage: "<src> <dest>",
			},
			{
				Name:      "overwrite",
				Usage:     "overwrites a Protected Entity with source file",
				Action:    overwrite,
				ArgsUsage: "<src> <dest>",
			},
		},
	}

	ctx := cliContext{
		Context:    context.Background(),
		addonInits: addonInits,
	}
	err := app.RunContext(ctx, os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

func setupProtectedEntityManager(c *cli.Context, addonInits map[string]server.InitFunc) (pem astrolabe.ProtectedEntityManager, err error) {
	pem, _, err = setupProtectedEntityManagers(c, addonInits, false)
	return
}

func setupProtectedEntityManagers(c *cli.Context, addonInits map[string]server.InitFunc, allowDual bool) (srcPem astrolabe.ProtectedEntityManager,
	destPem astrolabe.ProtectedEntityManager, err error) {
	confDirStr := c.String("confDir")
	if confDirStr != "" {
		srcPem = server.NewProtectedEntityManager(confDirStr, addonInits, logrus.StandardLogger())
		return
	}
	host := c.String("host")
	if host != "" {
		insecure := c.Bool("insecure")
		transport := restClient.DefaultTransportConfig()
		transport.Host = host
		if insecure {
			transport.Schemes = []string{"http"}
		}
	}

	if allowDual {
		destConfDirStr := c.String("destConfDir")
		if destConfDirStr != "" {
			destPem = server.NewProtectedEntityManager(confDirStr, nil, logrus.New())
		}
		if destPem == nil {
			destHost := c.String("destHost")
			if destHost != "" {
				destPem, err = setupHostPEM(destHost, c)
				if err != nil {
					return
				}
			}
		}
	}
	if destPem == nil {
		destPem = srcPem
	}
	return
}

func setupHostPEM(host string, c *cli.Context) (hostPem astrolabe.ProtectedEntityManager, err error) {
	insecure := c.Bool("insecure")
	transport := restClient.DefaultTransportConfig()
	transport.Host = host
	if insecure {
		transport.Schemes = []string{"http"}
	}

	restClient := restClient.NewHTTPClientWithConfig(nil, transport)
	hostPem, err = astrolabeClient.NewClientProtectedEntityManager(restClient)
	if err != nil {
		err = errors.Wrap(err, "Failed to create new ClientProtectedEntityManager")
		return
	}
	return
}

func types(c *cli.Context) error {
	var addonInits map[string]server.InitFunc
	cliCTX, ok := c.Context.(cliContext)
	if ok {
		addonInits = cliCTX.addonInits
	}

	pem, err := setupProtectedEntityManager(c, addonInits)
	if err != nil {
		logrus.Fatalf("Could not setup protected entity manager, err =%v", err)
	}
	for _, curPETM := range pem.ListEntityTypeManagers() {
		fmt.Println(curPETM.GetTypeName())
	}
	return nil
}

func ls(c *cli.Context) error {
	var addonInits map[string]server.InitFunc
	cliCTX, ok := c.Context.(cliContext)
	if ok {
		addonInits = cliCTX.addonInits
	}

	pem, err := setupProtectedEntityManager(c, addonInits)
	if err != nil {
		logrus.Fatalf("Could not setup protected entity manager, err =%v", err)
	}
	peType := c.Args().First()
	petm := pem.GetProtectedEntityTypeManager(peType)
	if petm == nil {
		logrus.Fatalf("Could not find type named %s", peType)
	}
	peIDs, err := petm.GetProtectedEntities(context.Background())
	if err != nil {
		logrus.Fatalf("Could not retrieve protected entities for type %s err:%b", peType, err)
	}

	for _, curPEID := range peIDs {
		fmt.Println(curPEID.String())
	}
	return nil
}

func lssn(c *cli.Context) error {
	var addonInits map[string]server.InitFunc
	cliCTX, ok := c.Context.(cliContext)
	if ok {
		addonInits = cliCTX.addonInits
	}

	peIDStr := c.Args().First()
	peID, err := astrolabe.NewProtectedEntityIDFromString(peIDStr)
	if err != nil {
		logrus.Fatalf("Could not parse protected entity ID %s, err: %v", peIDStr, err)
	}

	pem, err := setupProtectedEntityManager(c, addonInits)
	if err != nil {
		logrus.Fatalf("Could not setup protected entity manager, err =%v", err)
	}

	pe, err := pem.GetProtectedEntity(context.Background(), peID)
	if err != nil {
		logrus.Fatalf("Could not retrieve protected entity ID %s, err: %v", peIDStr, err)
	}

	snaps, err := pe.ListSnapshots(context.Background())
	if err != nil {
		logrus.Fatalf("Could not get snapshots for protected entity ID %s, err: %v", peIDStr, err)
	}

	for _, curSnapshotID := range snaps {
		curPESnapshotID := peID.IDWithSnapshot(curSnapshotID)
		fmt.Println(curPESnapshotID.String())
	}
	return nil
}

func show(c *cli.Context) error {
	var addonInits map[string]server.InitFunc
	cliCTX, ok := c.Context.(cliContext)
	if ok {
		addonInits = cliCTX.addonInits
	}
	peIDStr := c.Args().First()
	peID, err := astrolabe.NewProtectedEntityIDFromString(peIDStr)
	if err != nil {
		logrus.Fatalf("Could not parse protected entity ID %s, err: %v", peIDStr, err)
	}

	pem, err := setupProtectedEntityManager(c, addonInits)
	if err != nil {
		logrus.Fatalf("Could not setup protected entity manager, err =%v", err)
	}

	pe, err := pem.GetProtectedEntity(context.Background(), peID)
	if err != nil {
		logrus.Fatalf("Could not retrieve protected entity ID %s, err: %v", peIDStr, err)
	}

	info, err := pe.GetInfo(context.Background())
	if err != nil {
		logrus.Fatalf("Could not retrieve info for %s, err: %v", peIDStr, err)
	}
	fmt.Printf("%v\n", info)
	fmt.Println("Components:")
	components, err := pe.GetComponents(context.Background())
	for component := range components {
		fmt.Printf("%v\n", component)
	}
	return nil
}

func snap(c *cli.Context) error {
	var addonInits map[string]server.InitFunc
	cliCTX, ok := c.Context.(cliContext)
	if ok {
		addonInits = cliCTX.addonInits
	}
	peIDStr := c.Args().First()
	peID, err := astrolabe.NewProtectedEntityIDFromString(peIDStr)
	if err != nil {
		logrus.Fatalf("Could not parse protected entity ID %s, err: %v", peIDStr, err)
	}

	pem, err := setupProtectedEntityManager(c, addonInits)
	if err != nil {
		logrus.Fatalf("Could not setup protected entity manager, err =%v", err)
	}

	pe, err := pem.GetProtectedEntity(context.Background(), peID)
	if err != nil {
		logrus.Fatalf("Could not retrieve protected entity ID %s, err: %v", peIDStr, err)
	}
	snap, err := pe.Snapshot(context.Background(), make(map[string]map[string]interface{}))
	if err != nil {
		logrus.Fatalf("Could not snapshot protected entity ID %s, err: %v", peIDStr, err)
	}
	fmt.Println(snap.String())
	return nil
}

func rmsn(c *cli.Context) error {
	var addonInits map[string]server.InitFunc
	cliCTX, ok := c.Context.(cliContext)
	if ok {
		addonInits = cliCTX.addonInits
	}
	peIDStr := c.Args().First()
	peID, err := astrolabe.NewProtectedEntityIDFromString(peIDStr)
	if err != nil {
		logrus.Fatalf("Could not parse protected entity ID %s, err: %v", peIDStr, err)
	}
	if !peID.HasSnapshot() {
		logrus.Fatalf("Protected entity ID %s does not have a snapshot ID", peIDStr)
	}

	pem, err := setupProtectedEntityManager(c, addonInits)
	if err != nil {
		logrus.Fatalf("Could not setup protected entity manager, err =%v", err)
	}

	pe, err := pem.GetProtectedEntity(context.Background(), peID)
	if err != nil {
		logrus.Fatalf("Could not retrieve protected entity ID %s, err: %v", peIDStr, err)
	}
	success, err := pe.DeleteSnapshot(context.Background(), peID.GetSnapshotID(), make(map[string]map[string]interface{}))
	if err != nil {
		logrus.Fatalf("Could not remove snapshot ID %s, err: %v", peIDStr, err)
	}
	if success {
		logrus.Printf("Removed snapshot %s\n", peIDStr)
	}
	return nil
}

func cp(c *cli.Context) error {
	if c.NArg() != 2 {
		logrus.Fatalf("Expected two arguments for cp, got %d", c.NArg())
	}
	var addonInits map[string]server.InitFunc
	cliCTX, ok := c.Context.(cliContext)
	if ok {
		addonInits = cliCTX.addonInits
	}
	srcStr := c.Args().First()
	destStr := c.Args().Get(1)
	ctx := context.Background()
	var err error
	var srcPE astrolabe.ProtectedEntity
	var srcPEID, destPEID astrolabe.ProtectedEntityID
	var srcFile, destFile string
	srcPEID, err = astrolabe.NewProtectedEntityIDFromString(srcStr)
	if err != nil {
		srcFile = srcStr
	}
	destPEID, err = astrolabe.NewProtectedEntityIDFromString(destStr)
	if err != nil {
		destFile = destStr
	}
	pem, err := setupProtectedEntityManager(c, addonInits)
	if err != nil {
		logrus.Fatalf("Could not setup protected entity manager, err = %v", err)
	}

	var reader io.ReadCloser
	var writer io.WriteCloser
	fmt.Printf("cp from ")
	if srcFile != "" {
		fmt.Printf("file %s", srcFile)
		reader, err = os.Open(srcFile)
		if err != nil {
			logrus.Fatalf("Could not open srcFile %s, err = %v", srcFile, err)
		}
	} else {
		fmt.Printf("pe %s", srcPEID.String())
		srcPE, err = pem.GetProtectedEntity(ctx, srcPEID)
		if err != nil {
			logrus.Fatalf("Could not retrieve protected entity ID %s, err: %v", srcPEID.String(), err)
		}
	}
	fmt.Printf(" to ")
	if destFile != "" {
		fmt.Printf("file %s", destFile)
	} else {
		fmt.Printf("pe %s", destPEID.String())
	}
	fmt.Printf("\n")

	var bytesCopied int64
	if srcPE != nil && destFile != "" {
		// copy a src pe snapshot to a dest file
		var zipFileWriter io.WriteCloser
		zipFileWriter, err = os.Create(destFile)
		if err != nil {
			logrus.Fatalf("Could not create file %s, err: %v", destFile, err)
		}
		defer func() {
			if err := zipFileWriter.Close(); err != nil {
				logrus.Fatalf("Could not close file %s, err: %v", destFile, err)
			}
		}()

		bytesCopied, err = astrolabe.ZipProtectedEntityToFile(ctx, srcPE, zipFileWriter)
	} else {
		reader, err = os.Open(srcFile)
		if err != nil {
			logrus.Fatalf("Could not open srcFile %s, err = %v", srcFile, err)
		}
		defer reader.Close()

		bytesCopied, err = io.Copy(writer, reader)
	}

	if err != nil {
		logrus.Fatalf("Error copying %v", err)
	}

	fmt.Printf("Copied %d bytes\n", bytesCopied)
	return nil
}

func overwrite(c *cli.Context) error {
	if c.NArg() != 2 {
		logrus.Fatalf("Expected two arguments for overwrite, got %d", c.NArg())
	}
	var addonInits map[string]server.InitFunc
	cliCTX, ok := c.Context.(cliContext)
	if ok {
		addonInits = cliCTX.addonInits
	}
	srcStr := c.Args().First()
	destStr := c.Args().Get(1)
	ctx := context.Background()
	var err error
	var destPE astrolabe.ProtectedEntity
	var destPEID astrolabe.ProtectedEntityID
	var srcFile, destFile string
	_, err = astrolabe.NewProtectedEntityIDFromString(srcStr)
	if err != nil {
		srcFile = srcStr
	}
	destPEID, err = astrolabe.NewProtectedEntityIDFromString(destStr)
	if err != nil {
		destFile = destStr
	}
	pem, err := setupProtectedEntityManager(c, addonInits)
	if err != nil {
		logrus.Fatalf("Could not setup protected entity manager, err = %v", err)
	}

	fmt.Printf("overwrite from ")
	if srcFile != "" {
		fmt.Printf("file %s", srcFile)
	} else {
		logrus.Fatalf("Use a Protected Entity ID(%v) as the src is not supported", srcStr)
	}
	fmt.Printf(" to ")
	if destFile != "" {
		logrus.Fatalf("Use a file path(%v) as the destination is not supported", destStr)
	} else {
		fmt.Printf("pe %s", destPEID.String())
		destPE, err = pem.GetProtectedEntity(ctx, destPEID)
		if err != nil {
			logrus.Fatalf("Could not retrieve protected entity ID %s, err: %v", destPEID.String(), err)
		}
		if _, err := destPE.GetInfo(context.Background()); err != nil {
			logrus.Fatalf("Could not retrieve info for %s, err: %v", destPE.GetID().String(), err)
		}

	}
	fmt.Printf("\n")

	file, err := os.Open(srcFile)
	if err != nil {
		logrus.Fatalf("Could not open srcFile %s, err = %v", srcFile, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		logrus.Fatalf("Got err %v getting stat for source file %v", err, srcFile)
	}

	// unzip the source zip file and use its data to overwrite the dest protected entity
	if err = astrolabe.UnzipFileToProtectedEntity(ctx, file, fileInfo.Size(), destPE); err != nil {
		logrus.Fatalf("Error overwriting %v", err)
	}

	return nil
}
