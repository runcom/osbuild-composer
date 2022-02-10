package disk

import (
	"fmt"
)

type Partition struct {
	Start    uint64 // Start of the partition in sectors
	Size     uint64 // Size of the partition in sectors
	Type     string // `gpt` or `dos`
	Bootable bool   // `Legacy BIOS bootable` (GPT) or `active` (DOS) flag

	// ID of the partition, dos doesn't use traditional UUIDs, therefore this
	// is just a string.
	UUID string

	// If nil, the partition is raw; It doesn't contain a payload.
	Payload Entity
}

func (p *Partition) IsContainer() bool {
	return true
}

func (p *Partition) Clone() Entity {
	if p == nil {
		return nil
	}

	partition := &Partition{
		Start:    p.Start,
		Size:     p.Size,
		Type:     p.Type,
		Bootable: p.Bootable,
		UUID:     p.UUID,
	}

	if p.Payload != nil {
		partition.Payload = p.Payload.Clone()
	}

	return partition
}

// Grow the partition to the given size in sectors. Will do nothing
// if the partition is larger than the requested size.
func (p *Partition) GrowTo(sectors uint64) {
	if p.Size < sectors {
		p.Size = sectors
	}
}

func (pt *Partition) GetItemCount() uint {
	if pt == nil || pt.Payload == nil {
		return 0
	}
	return 1
}

func (p *Partition) GetChild(n uint) Entity {
	if n != 0 {
		panic(fmt.Sprintf("invalid child index for Partition: %d != 0", n))
	}
	return p.Payload
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
