package osbuild2

import (
	"math/rand"
	"testing"

	"github.com/osbuild/osbuild-composer/internal/blueprint"
	"github.com/osbuild/osbuild-composer/internal/disk"
	"github.com/stretchr/testify/assert"
)

var lusk_lvm = disk.PartitionTable{
	UUID: "D209C89E-EA5E-4FBD-B161-B461CCE297E0",
	Type: "gpt",
	Partitions: []disk.Partition{
		{
			Size:     1048576, // 1MB
			Bootable: true,
			Type:     disk.BIOSBootPartitionGUID,
			UUID:     disk.BIOSBootPartitionUUID,
		},
		{
			Size: 209715200, // 200 MB
			Type: disk.EFISystemPartitionGUID,
			UUID: disk.EFISystemPartitionUUID,
			Payload: &disk.Filesystem{
				Type:         "vfat",
				UUID:         disk.EFIFilesystemUUID,
				Mountpoint:   "/boot/efi",
				Label:        "EFI-SYSTEM",
				FSTabOptions: "defaults,uid=0,gid=0,umask=077,shortname=winnt",
				FSTabFreq:    0,
				FSTabPassNo:  2,
			},
		},
		{
			Size: 1024000, // 500 MB
			Type: disk.FilesystemDataGUID,
			UUID: disk.FilesystemDataUUID,
			Payload: &disk.Filesystem{
				Type:         "xfs",
				Mountpoint:   "/boot",
				Label:        "boot",
				FSTabOptions: "defaults",
				FSTabFreq:    0,
				FSTabPassNo:  0,
			},
		},
		{
			Type: disk.FilesystemDataGUID,
			UUID: disk.RootPartitionUUID,
			Size: 5 * 1024 * 1024 * 1024,
			Payload: &disk.LUKSContainer{
				Label:      "crypt-root",
				Passphrase: "osbuild",
				PBKDF: disk.Argon2id{
					Memory:      32,
					Iterations:  4,
					Parallelism: 1,
				},
				Payload: &disk.LVMVolumeGroup{
					Name:        "root",
					Description: "root volume group",
					LogicalVolumes: []disk.LVMLogicalVolume{
						{
							Size: 2 * 1024 * 1024 * 1024,
							Name: "rootlv",
							Payload: &disk.Filesystem{
								Type:         "xfs",
								Label:        "root",
								Mountpoint:   "/",
								FSTabOptions: "defaults",
								FSTabFreq:    0,
								FSTabPassNo:  0,
							},
						},
						{
							Size: 2 * 1024 * 1024 * 1024,
							Name: "homelv",
							Payload: &disk.Filesystem{
								Type:         "xfs",
								Label:        "root",
								Mountpoint:   "/home",
								FSTabOptions: "defaults",
								FSTabFreq:    0,
								FSTabPassNo:  0,
							},
						},
					},
				},
			},
		},
	},
}

func TestGenDeviceCreationStages(t *testing.T) {
	assert := assert.New(t)

	// math/rand is good enough in this case
	/* #nosec G404 */
	rng := rand.New(rand.NewSource(13))

	pt, err := disk.NewPartitionTable(&lusk_lvm, []blueprint.FilesystemCustomization{}, 0, rng)
	assert.NoError(err)

	stages := GenDeviceCreationStages(pt, "image.raw")

	// we should have two stages
	assert.Equal(len(stages), 2)

	// first one should be a "org.osbuild.luks2.format"
	luks := stages[0]
	assert.Equal(luks.Type, "org.osbuild.luks2.format")

	// it needs to have one device
	assert.Equal(len(luks.Devices), 1)

	// the device should be called `device`
	device, ok := luks.Devices["device"]
	assert.True(ok, "Need device called `device`")

	// device should be a loopback device
	assert.Equal(device.Type, "org.osbuild.loopback")

	lvm := stages[1]
	assert.Equal(lvm.Type, "org.osbuild.lvm2.create")
	lvmOptions, ok := lvm.Options.(*LVM2CreateStageOptions)
	assert.True(ok, "Need LVM2CreateStageOptions for org.osbuild.lvm2.create")

	// LVM should have two volumes
	assert.Equal(len(lvmOptions.Volumes), 2)
	rootlv := lvmOptions.Volumes[0]
	assert.Equal(rootlv.Name, "rootlv")

	homelv := lvmOptions.Volumes[1]
	assert.Equal(homelv.Name, "homelv")

	// it needs to have two(!) devices, the loopback and the luks
	assert.Equal(len(lvm.Devices), 2)

	// this is the target one, which should be the luks one
	device, ok = lvm.Devices["device"]
	assert.True(ok, "Need device called `device`")
	assert.Equal(device.Type, "org.osbuild.luks2")
	assert.NotEmpty(device.Parent, "Need a parent device for LUKS on loopback")

	luksOptions, ok := device.Options.(*LUKS2DeviceOptions)
	assert.True(ok, "Need LUKS2DeviceOptions for luks device")
	assert.Equal(luksOptions.Passphrase, "osbuild")

	parent, ok := lvm.Devices[device.Parent]
	assert.True(ok, "Need device called `device`")
	assert.Equal(parent.Type, "org.osbuild.loopback")

}
