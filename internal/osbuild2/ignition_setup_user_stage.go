package osbuild2

type IgnitionSetupUserStageOptions struct {
}

func (IgnitionSetupUserStageOptions) isStageOptions() {}

func NewIgnitionSetupUserStage(options *IgnitionSetupUserStageOptions) *Stage {
	return &Stage{
		Type:    "org.osbuild.ignition.setup-user",
		Options: options,
	}
}
