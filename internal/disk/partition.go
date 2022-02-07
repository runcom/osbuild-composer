package disk

import (
	"fmt"

	osbuild "github.com/osbuild/osbuild-composer/internal/osbuild1"
)

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

func (p *Partition) IsContainer() bool {
	return true
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

func (pt *Partition) GetItemCount() uint {
	if pt.Filesystem == nil {
		return 0
	}
	return 1
}

func (p *Partition) GetChild(n uint) Entity {
	if n != 0 {
		panic(fmt.Sprintf("invalid child index for Partition: %d != 0", n))
	}
	return p.Filesystem
}

func (p *Partition) GetSize() uint64 {
	return p.Size
}

func (p *Partition) EnsureSize(s uint64) bool {
	if s > p.Size {
		p.Size = s
		return true
	}
	return false
}
