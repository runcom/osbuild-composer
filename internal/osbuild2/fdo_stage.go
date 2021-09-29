package osbuild2

type FDOStageOptions struct {
	// PEM encoded string of root certificates for FDO DIUN
	RootCerts string `json:"diun_pub_key_root_certs"`
}

func (FDOStageOptions) isStageOptions() {}

// Creates FDO specific options
func NewFDOStage(options *FDOStageOptions) *Stage {
	return &Stage{
		Type:    "org.osbuild.fdo",
		Options: options,
	}
}
