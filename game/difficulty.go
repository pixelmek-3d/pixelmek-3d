package game

var DifficultyLevels []*DifficultyLevel

type DifficultyLevel struct {
	Name                      string
	EnemyDamageTakenModifier  float64
	PlayerDamageTakenModifier float64
	FriendlyFireEnabled       bool
}

func (d *DifficultyLevel) String() string {
	return d.Name
}

func init() {
	DifficultyLevels = []*DifficultyLevel{
		DifficultyRecruit(),
		DifficultyRegular(),
		DifficultyVeteran(),
		DifficultyAce(),
	}
}

func DifficultyRecruit() *DifficultyLevel {
	return &DifficultyLevel{
		Name:                      `Recruit`,
		EnemyDamageTakenModifier:  4.0,
		PlayerDamageTakenModifier: 0.5,
		FriendlyFireEnabled:       false,
	}
}

func DifficultyRegular() *DifficultyLevel {
	return &DifficultyLevel{
		Name:                      `Regular`,
		EnemyDamageTakenModifier:  2.5,
		PlayerDamageTakenModifier: 0.75,
		FriendlyFireEnabled:       false,
	}
}

func DifficultyVeteran() *DifficultyLevel {
	return &DifficultyLevel{
		Name:                      `Veteran`,
		EnemyDamageTakenModifier:  2.0,
		PlayerDamageTakenModifier: 1.0,
		FriendlyFireEnabled:       true,
	}
}

func DifficultyAce() *DifficultyLevel {
	return &DifficultyLevel{
		Name:                      `Ace`,
		EnemyDamageTakenModifier:  1.5,
		PlayerDamageTakenModifier: 1.5,
		FriendlyFireEnabled:       true,
	}
}
