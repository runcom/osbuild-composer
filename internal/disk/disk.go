// Disk package contains abstract data-types to define disk-related entities.
//
// PartitionTable, Partition and Filesystem types are currently defined.
// All of them can be 1:1 converted to osbuild.QEMUAssemblerOptions.
package disk

import (
	"encoding/hex"
	"io"

	"github.com/google/uuid"
)

// TODO: guard against nil dereferencing in pointer methods

const (
	// Default sector size in bytes
	DefaultSectorSize = 512

	DefaultGrainBytes = uint64(1048576) // 1 MiB

	// UUIDs
	BIOSBootPartitionGUID = "21686148-6449-6E6F-744E-656564454649"
	BIOSBootPartitionUUID = "FAC7F1FB-3E8D-4137-A512-961DE09A5549"

	FilesystemDataGUID = "0FC63DAF-8483-4772-8E79-3D69D8477DE4"
	FilesystemDataUUID = "CB07C243-BC44-4717-853E-28852021225B"

	EFISystemPartitionGUID = "C12A7328-F81F-11D2-BA4B-00A0C93EC93B"
	EFISystemPartitionUUID = "68B2905B-DF3E-4FB3-80FA-49D1E773AA33"
	EFIFilesystemUUID      = "7B77-95E7"

	RootPartitionUUID = "6264D520-3FB9-423F-8AB8-7A0A8E3D3562"
)

type Entity interface {
	IsContainer() bool
	Clone() Entity
}

type Container interface {
	Entity
	GetItemCount() uint
	GetChild(n uint) Entity
}

type Sizeable interface {
	EnsureSize(size uint64) bool
	GetSize() uint64
}

type Mountable interface {
	GetMountpoint() string
	GetFSType() string
	GetFSSpec() FSSpec
	GetFSTabOptions() FSTabOptions
}

type VolumeContainer interface {
	CreateVolume(mountpoint string, size uint64) (Entity, error)
}

// FSSpec for a filesystem (UUID and Label); the first field of fstab(5)
type FSSpec struct {
	UUID  string
	Label string
}

type FSTabOptions struct {
	// The fourth field of fstab(5); fs_mntops
	MntOps string
	// The fifth field of fstab(5); fs_freq
	Freq uint64
	// The sixth field of fstab(5); fs_passno
	PassNo uint64
}

// uuid generator helpers

// GeneratesnewRandomUUIDFromReader generates a new random UUID (version
// 4 using) via the given random number generator.
func newRandomUUIDFromReader(r io.Reader) (uuid.UUID, error) {
	var id uuid.UUID
	_, err := io.ReadFull(r, id[:])
	if err != nil {
		return uuid.Nil, err
	}
	id[6] = (id[6] & 0x0f) | 0x40 // Version 4
	id[8] = (id[8] & 0x3f) | 0x80 // Variant is 10
	return id, nil
}

// NewRandomVolIDFromReader creates a random 32 bit hex string to use as a
// volume ID for FAT filesystems
func NewRandomVolIDFromReader(r io.Reader) (string, error) {
	volid := make([]byte, 4)
	_, err := r.Read(volid)
	return hex.EncodeToString(volid), err
}
