package osbuild2

// Provide access to LUKS2 container

type LUKS2DeviceOptions struct {
	UUID       string `json:"uuid"`
	Passphrase string `json:"passphrase"`
}

func (LUKS2DeviceOptions) isDeviceOptions() {}

func NewLUKS2Device(options *LUKS2DeviceOptions) *Device {
	return &Device{
		Type:    "org.osbuild.luks2",
		Options: options,
	}
}
