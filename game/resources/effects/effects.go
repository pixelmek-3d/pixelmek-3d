package effects

import (
	"github.com/harbdog/pixelmek-3d/game/model"
)

var (
	Explosions map[string]*model.ModelEffectResource
	Smokes     map[string]*model.ModelEffectResource
	_exploKeys []string
	_smokeKeys []string
)

func init() {
	// init known explosion animations
	Explosions = make(map[string]*model.ModelEffectResource)

	Explosions["01"] = &model.ModelEffectResource{
		Audio: "explosion-1.ogg",
		Scale: 0.25,
		Image: "explosion_01.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       6,
			Rows:          4,
			AnimationRate: 4,
		},
	}

	Explosions["02"] = &model.ModelEffectResource{
		Audio: "explosion-0.ogg",
		Scale: 0.25,
		Image: "explosion_02.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       6,
			Rows:          4,
			AnimationRate: 4,
		},
	}

	Explosions["03"] = &model.ModelEffectResource{
		Audio: "explosion-0.ogg",
		Scale: 0.25,
		Image: "explosion_03.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       6,
			Rows:          4,
			AnimationRate: 4,
		},
	}

	Explosions["04"] = &model.ModelEffectResource{
		Audio: "explosion-0.ogg",
		Scale: 0.25,
		Image: "explosion_04.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       6,
			Rows:          4,
			AnimationRate: 4,
		},
	}

	Explosions["07"] = &model.ModelEffectResource{
		Audio: "explosion-4.ogg",
		Scale: 0.35,
		Image: "explosion_07.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       8,
			Rows:          4,
			AnimationRate: 3,
		},
	}

	Explosions["09"] = &model.ModelEffectResource{
		Audio: "explosion-3.ogg",
		Scale: 0.35,
		Image: "explosion_07.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       8,
			Rows:          4,
			AnimationRate: 3,
		},
	}

	Explosions["10"] = &model.ModelEffectResource{
		Audio: "explosion-3.ogg",
		Scale: 0.35,
		Image: "explosion_10.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       8,
			Rows:          4,
			AnimationRate: 3,
		},
	}

	Explosions["11"] = &model.ModelEffectResource{
		Audio: "explosion-2.ogg",
		Scale: 0.25,
		Image: "explosion_11.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       6,
			Rows:          4,
			AnimationRate: 4,
		},
	}

	// init known smoke animations
	Smokes = make(map[string]*model.ModelEffectResource)

	Smokes["01"] = &model.ModelEffectResource{
		Scale: 0.25,
		Image: "smoke_01.png",
		ImageSheet: &model.ModelResourceImageSheet{
			Columns:       8,
			Rows:          4,
			AnimationRate: 5,
		},
	}

	Smokes["01.5"] = &model.ModelEffectResource{
		Scale: 0.25,
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

	// generate list of keys so rand function can be used
	_exploKeys = make([]string, 0, len(Explosions))
	_smokeKeys = make([]string, 0, len(Smokes))
	for key := range Explosions {
		_exploKeys = append(_exploKeys, key)
	}
	for key := range Smokes {
		_smokeKeys = append(_smokeKeys, key)
	}
}

func RandExplosionKey() string {
	return _exploKeys[model.Randish.Intn(len(_exploKeys))]
}

func RandSmokeKey() string {
	return _smokeKeys[model.Randish.Intn(len(_smokeKeys))]
}
