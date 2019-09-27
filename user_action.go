package main

import (
	"github.com/onionwyl/smart-light/db"
	"github.com/onionwyl/smart-light/light"
	"github.com/onionwyl/smart-light/model"
	"time"
)

func detectUserAction() {
	dbConn := db.GetDBConn()
	l := light.GetLight()
	tmp := light.Light{}
	for {
		l.GetInfo()
		action := model.Actions{}
		action.Changer = 0
		if tmp.Addr == "" {
			goto NEXT
		} else if l.Bright != tmp.Bright{
			lastAction := model.Actions{}
			res := dbConn.Where("action = ? AND light_bright_from = ? AND light_bright = ?", "bright", tmp.Bright, l.Bright).Last(&lastAction)
			if res.Error == nil && time.Now().Sub(lastAction.CreatedAt) < time.Second * 5 {
				goto NEXT
			}
			action.Action = "bright"
			action.LightBrightFrom = tmp.Bright
			action.LightBright = l.Bright
			dbConn.Create(&action)

		} else if l.CT != tmp.CT {
			lastAction := model.Actions{}
			res := dbConn.Where("action = ? AND light_ct_from = ? AND light_ct = ?", "ct", tmp.CT, l.CT).Last(&lastAction)
			if res.Error == nil && time.Now().Sub(lastAction.CreatedAt) < time.Second * 5 {
				goto NEXT
			}
			action.Action = "ct"
			action.LightCtFrom = tmp.CT
			action.LightCt = l.CT
			dbConn.Create(&action)
		} else if l.Power != tmp.Power {
			lastAction := model.Actions{}
			var state int
			if tmp.Power == "on" {
				state = 1
			} else {
				state = 0
			}
			res := dbConn.Where("action = ? AND light_state = ?", "toggle",state).Last(&lastAction)
			if res.Error == nil && time.Now().Sub(lastAction.CreatedAt) < time.Second * 5 {
				goto NEXT
			}
			action.Action = "toggle"
			if l.Power == "on" {
				action.LightState = 1
			} else {
				action.LightState = 0
			}
			dbConn.Create(&action)
		}
	NEXT:
		tmp = l
		time.Sleep(time.Second * 3)
	}
}