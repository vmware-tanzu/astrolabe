/*
 * Copyright 2019 the Astrolabe contributors
 * SPDX-License-Identifier: Apache-2.0
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package s3repository

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/fs"
	"github.com/vmware-tanzu/astrolabe/pkg/ivd"
    "log"
	"testing"
)

func TestProtectedEntityTypeManager(t *testing.T) {
	s3petm, err := setupPETM(t, "test")
	if err != nil {
		t.Fatal(err)
	}
	ids, err := s3petm.GetProtectedEntities(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("# of PEs returned = %d\n", len(ids))

	for _, id := range ids {
		t.Logf("%s\n", id)

	}
}

func setupPETM(t *testing.T, typeName string) (*ProtectedEntityTypeManager, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-1")},
	)
	if err != nil {
		return nil, err
	}
	s3petm, err := NewS3RepositoryProtectedEntityTypeManager(typeName, *sess, "velero-plugin-s3-repo",
		"backups/vsphere-volumes-repo/", nil)
	if err != nil {
		return nil, err
	}
	return s3petm, err
}

/*
func TestCreateDeleteProtectedEntity(t *testing.T) {
	s3petm, err := setupPETM(t, "test")
	if err != nil {
		t.Fatal(err)
	}
	peID := astrolabe.NewProtectedEntityIDWithSnapshotID("test", "unique1", astrolabe.NewProtectedEntitySnapshotID("snapshot1"))
	peInfo := astrolabe.NewProtectedEntityInfo(peID, "testPE", nil, nil, nil, nil)
	repoPE, err := s3petm.CopyFromInfo(context.Background(), peInfo, astrolabe.AllocateNewObject)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Create repo PE %s\n", repoPE.GetID().String())

}
*/

func TestCopyFSProtectedEntity(t *testing.T) {
	s3petm, err := setupPETM(t, "fs")
	if err != nil {
		t.Fatal(err)
	}

	fsParams := make(map[string]interface{})
	fsParams["root"] = "/Users/dsmithuchida/astrolabe_fs_root"

	fsPETM, err := fs.NewFSProtectedEntityTypeManagerFromConfig(fsParams, "notUsed", logrus.New())
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	fsPEs, err := fsPETM.GetProtectedEntities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, fsPEID := range fsPEs {
		// FS doesn't have snapshots, but repository likes them, so fake one
		snapPEID := astrolabe.NewProtectedEntityIDWithSnapshotID(fsPEID.GetPeType(), fsPEID.GetID(),
			astrolabe.NewProtectedEntitySnapshotID("dummy-snap-id"))
		fsPE, err := fsPETM.GetProtectedEntity(ctx, snapPEID)
		if err != nil {
			t.Fatal(err)
		}
		s3PE, err := s3petm.Copy(ctx, fsPE, astrolabe.AllocateNewObject)
		if err != nil {
			t.Fatal(err)
		}

		newFSPE, err := fsPETM.Copy(ctx, s3PE, astrolabe.AllocateNewObject)
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("Restored new FSPE %s\n", newFSPE.GetID().String())
	}
}

func TestRetrieveEntity(t *testing.T) {
	s3petm, err := setupPETM(t, "ivd")
	if err != nil {
		t.Fatal(err)
	}

	ivdParams := make(map[string]interface{})
	ivdParams[ivd.HostVcParamKey] = "10.208.22.211"
	ivdParams[ivd.PortVcParamKey] = "443"
	ivdParams[ivd.InsecureFlagVcParamKey] = "Y"
	ivdParams[ivd.UserVcParamKey] = "administrator@vsphere.local"
	ivdParams[ivd.PasswordVcParamKey] = "Admin!23"

	ivdPETM, err := ivd.NewIVDProtectedEntityTypeManagerFromConfig(ivdParams, "notUsed", logrus.New())
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	peid := astrolabe.NewProtectedEntityIDWithSnapshotID("ivd", "ff9ac770-7ecd-405d-841b-6232857520d4",
		astrolabe.NewProtectedEntitySnapshotID("bde7e96d-8065-4bd5-a82a-edd7b2f540de"))
	s3PE, err := s3petm.GetProtectedEntity(ctx, peid)
	if err != nil {
		t.Fatal(err)
	}

	//mdr, err := s3PE.GetMetadataReader(ctx)
	/*
		mdr, err := s3PE.GetDataReader(ctx)

		if err != nil {
			t.Fatal(err)
		}
		op, err := os.Create("/home/dsmithuchida/tmp/xyzzy")
		if err != nil {
			t.Fatal(err)
		}

		bytesCopied, err := io.Copy(op, mdr)
		if err != nil {
			t.Fatal(err)
		}


		fmt.Printf("%d bytes copied\n", bytesCopied)
	*/
	newIVDPE, err := ivdPETM.Copy(ctx, s3PE, astrolabe.AllocateNewObject)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Restored new IVDPE %s\n", newIVDPE.GetID().String())

	/*
		dataReader, err := pe.GetDataReader(ctx)
		if err != nil {
			t.Fatal(err)
		}
		buf := make([]byte, 1024*1024)
		keepReading := true
		for keepReading {
			read, err := dataReader.Read(buf)
			if err != nil {
				keepReading = false
				if err != io.EOF {
					t.Fatal(err)
				}
			}
			fmt.Printf("Read %d bytes\n", read)
		}

	*/
}
func TestCopyIVDProtectedEntity(t *testing.T) {
	s3petm, err := setupPETM(t, "ivd")
	if err != nil {
		t.Fatal(err)
	}

	ivdParams := make(map[string]interface{})
	ivdParams[ivd.HostVcParamKey] = "10.208.22.211"
	ivdParams[ivd.PortVcParamKey] = "443"
	ivdParams[ivd.InsecureFlagVcParamKey] = "Y"
	ivdParams[ivd.UserVcParamKey] = "administrator@vsphere.local"
	ivdParams[ivd.PasswordVcParamKey] = "Admin!23"

	ivdPETM, err := ivd.NewIVDProtectedEntityTypeManagerFromConfig(ivdParams, "notUsed", logrus.New())
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	/*ivdPEs, err := ivdPETM.GetProtectedEntities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	*/
	ctx = context.Background()

	//PESSID := astrolabe.NewProtectedEntitySnapshotID("ecb7fa78-cef9-4459-b898-17a39f582d9b")
	//ivdPEID := astrolabe.NewProtectedEntityIDWithSnapshotID("ivd", "cf29221a-381b-4036-825a-56bf8294ed38", ivdPESSID)
	//ivdPEID := astrolabe.NewProtectedEntityID("ivd", "22dd64e4-987f-4bec-b997-04c12a0e0d86")
	ivdPEID := astrolabe.NewProtectedEntityID("ivd", "c5ff5363-f6d2-4240-ae2e-8d9f8cee484c")
	var snapID astrolabe.ProtectedEntitySnapshotID
	if false {
		ivdPE, err := ivdPETM.GetProtectedEntity(ctx, ivdPEID)

		snapID, err = ivdPE.Snapshot(ctx)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		snapID = astrolabe.NewProtectedEntitySnapshotID("1c929d85-1148-43d8-aa1d-12098b119325")
	}
	var s3PE astrolabe.ProtectedEntity
	snapPEID := astrolabe.NewProtectedEntityIDWithSnapshotID("ivd", ivdPEID.GetID(), snapID)

	if false {

		snapPE, err := ivdPETM.GetProtectedEntity(ctx, snapPEID)
		if err != nil {
			t.Fatal(err)
		}
		s3PE, err = s3petm.Copy(ctx, snapPE, astrolabe.AllocateNewObject)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		s3PE, err = s3petm.GetProtectedEntity(ctx, snapPEID)
	}
	newIVDPE, err := ivdPETM.Copy(ctx, s3PE, astrolabe.AllocateNewObject)
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Restored new IVDPE %s\n", newIVDPE.GetID().String())

	/*
		for _, ivdPEID := range ivdPEs {
			ivdPE, err := ivdPETM.GetProtectedEntity(ctx, ivdPEID)
			if err != nil {
				t.Fatal(err)
			}
			snapID, err := ivdPE.Snapshot(ctx)
			if err == nil {

				snapPEID := astrolabe.NewProtectedEntityIDWithSnapshotID("ivd", ivdPEID.GetID(), *snapID)
				snapPE, err := ivdPETM.GetProtectedEntity(ctx, snapPEID)
				if err != nil {
					t.Fatal(err)
				}
				s3PE, err := s3petm.Copy(ctx, snapPE, astrolabe.AllocateNewObject)
				if err != nil {
					t.Fatal(err)
				}

				newIVDPE, err := ivdPETM.Copy(ctx, s3PE, astrolabe.AllocateNewObject)
				if err != nil {
					t.Fatal(err)
				}
				log.Printf("Restored new IVDPE %s\n", newIVDPE.GetID().String())
				status, err := ivdPE.DeleteSnapshot(ctx, *snapID)
				if err != nil {
					t.Fatal(err)
				}
				if !status {
					t.Fatal("Snapshot delete returned false")
				}
			} else {
				log.Printf("Snapshot failed for %s, skipping\n", ivdPEID.String())
			}
		}
	*/
}
