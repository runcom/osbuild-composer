// Disk package contains abstract data-types to define disk-related entities.
//
// PartitionTable, Partition and Filesystem types are currently defined.
// All of them can be 1:1 converted to osbuild.QEMUAssemblerOptions.
package disk

import (
	"encoding/hex"
	"io"

	"github.com/google/uuid"
	osbuild "github.com/osbuild/osbuild-composer/internal/osbuild1"
)

const (
	// Default sector size in bytes
	DefaultSectorSize = 512

	DefaultGrainBytes = uint64(1024 * 1024) // 1 MiB
)

type Entity interface {
	IsContainer() bool
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

type Partition struct {
	Start    uint64 // Start of the partition in sectors
	Size     uint64 // Size of the partition in sectors
	Type     string // `gpt` or `dos`
	Bootable bool   // `Legacy BIOS bootable` (GPT) or `active` (DOS) flag
	// ID of the partition, dos doesn't use traditional UUIDs, therefore this
	// is just a string.
	UUID string
	// If nil, the partition is raw; It doesn't contain a filesystem.
	Filesystem *Filesystem
}

// Grow the partition to the given size in sectors. Will do nothing
// if the partition is larger than the requested size.
func (p *Partition) GrowTo(sectors uint64) {
	if p.Size < sectors {
		p.Size = sectors
	}
}

// Converts Partition to osbuild.QEMUPartition that encodes the same partition.
func (p *Partition) QEMUPartition() osbuild.QEMUPartition {
	var fs *osbuild.QEMUFilesystem
	if p.Filesystem != nil {
		f := p.Filesystem.QEMUFilesystem()
		fs = &f
	}
	return osbuild.QEMUPartition{
		Start:      p.Start,
		Size:       p.Size,
		Type:       p.Type,
		Bootable:   p.Bootable,
		UUID:       p.UUID,
		Filesystem: fs,
	}
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
