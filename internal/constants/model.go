package constants

type TelnyxModel string

const (
	ModelMoonShotKimi TelnyxModel = "moonshotai/Kimi-K2.6"
	ModelZai52        TelnyxModel = "zai-org/GLM-5.2"
	ModelZai51        TelnyxModel = "zai-org/GLM-5.1-FP8"
	ModelMiniMax      TelnyxModel = "MiniMaxAI/MiniMax-M3-MXFP8"
)

const DefaultModel = ModelMoonShotKimi

func (m *TelnyxModel) String() string {
	return string(*m)
}

func (m *TelnyxModel) IsValid() bool {
	switch *m {
	case ModelMoonShotKimi, ModelZai52, ModelZai51, ModelMiniMax:
		return true
	default:
		return false
	}
}
