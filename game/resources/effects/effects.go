package effects

import (
	"github.com/harbdog/pixelmek-3d/game/model"
)

var (
	Explosions map[string]*model.ModelEffectResource
	Smokes     map[string]*model.ModelEffectResource
)

func init() {
	// init known explosion animations
	Explosions = make(map[string]*model.ModelEffectResource)

	Explosions["01"] = &model.ModelEffectResource{
		Audio: "explosion.ogg",
		Scale: 1.0,
		Image: "explosion_01.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       8,
			Rows:          3,
			AnimationRate: 4,
		},
	}

	Explosions["07"] = &model.ModelEffectResource{
		Audio: "explosion.ogg",
		Scale: 0.5,
		Image: "explosion_07.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       13,
			Rows:          2,
			AnimationRate: 4,
		},
	}

	// init known smoke animations
	Smokes = make(map[string]*model.ModelEffectResource)

	Smokes["01"] = &model.ModelEffectResource{
		Scale: 0.5,
		Image: "smoke_01.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       8,
			Rows:          4,
			AnimationRate: 5,
		},
	}

	Smokes["01.5"] = &model.ModelEffectResource{
		Scale: 0.5,
		Image: "smoke_01.5.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       8,
			Rows:          4,
			AnimationRate: 5,
		},
	}

	Smokes["01.75"] = &model.ModelEffectResource{
		Scale: 0.5,
		Image: "smoke_01.75.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       8,
			Rows:          4,
			AnimationRate: 5,
		},
	}
}
