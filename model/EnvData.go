package model

import "github.com/jinzhu/gorm"

type EnvData struct{
	gorm.Model
	LightBright int
	LightCT int
	LightState int
	BtSensor1 int
	BtSensor2 int
	DisSensor1 float32
	DisSensor2 float32
	IrSensor1 int
	IrSensor2 int
	UserState int
}
