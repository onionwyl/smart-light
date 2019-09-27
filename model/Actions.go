package model

import "github.com/jinzhu/gorm"

type Actions struct {
	gorm.Model
	Action string	// bright ct toggle
	LightBright int
	LightBrightFrom int
	LightCt int
	LightCtFrom int
	LightState int
	Changer int // 0 for user from api or light and 1 for algorithm
}
