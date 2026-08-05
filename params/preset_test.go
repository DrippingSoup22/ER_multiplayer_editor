package params

import (
	"er_pvp_mod/core"
	"testing"
)

func TestAllPresetsPassCoreValidation(t *testing.T) {
	views := []ParamView{ViewInvader, ViewSign, ViewSignPlace, ViewHunter}
	for _, view := range views {
		for _, preset := range PresetsForView(view) {
			got := preset.Apply(core.NetworkParamDefaults())
			if err := core.ValidateNetworkParams(got); err != nil {
				t.Errorf("%s/%s preset is invalid: %v", view, preset.Key, err)
			}
		}
	}
}

func TestParameterMetadataIsCompleteAndUnique(t *testing.T) {
	views := []ParamView{ViewInvader, ViewSign, ViewSignPlace, ViewHunter}
	seen := make(map[string]bool)
	count := 0
	for _, view := range views {
		if len(PresetsForView(view)) != 3 {
			t.Errorf("%s has %d presets, want 3", view, len(PresetsForView(view)))
		}
		for _, param := range ParamsForView(view) {
			count++
			if param.Key == "" || param.Label == "" || param.ShortDef == "" || param.LongDef == "" {
				t.Errorf("%s contains incomplete parameter metadata: %+v", view, param)
			}
			if param.Min > param.Max {
				t.Errorf("%s has reversed range %.0f..%.0f", param.Key, param.Min, param.Max)
			}
			if seen[param.Key] {
				t.Errorf("duplicate parameter key %q", param.Key)
			}
			seen[param.Key] = true
		}
	}
	if count != 24 {
		t.Errorf("parameter count = %d, want 24", count)
	}
}
