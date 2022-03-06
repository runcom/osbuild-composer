package osbuild2

type IgnitionFirstbootCompleteStageOptions struct {
}

func (IgnitionFirstbootCompleteStageOptions) isStageOptions() {}

func NewIgnitionFirstbootCompleteStage(options *IgnitionFirstbootCompleteStageOptions) *Stage {
	return &Stage{
		Type:    "org.osbuild.ignition.firstboot-complete",
		Options: options,
	}
}
