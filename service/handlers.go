package service

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/onionwyl/smart-light/db"
	"github.com/onionwyl/smart-light/light"
	"github.com/onionwyl/smart-light/model"
	"html/template"
	"net/http"
	"strconv"
)

type status struct {
	Ip string		`json:"ip"`
	State string 		`json:"state"`
	Bright int		`json:"bright"`
	Ct int			`json:"ct"`
	Btsensor1 int 	`json:"btsensor1"`
	Btsensor2 int 	`json:"btsensor2"`
	Dissensor1 float32 	`json:"dissensor1"`
	Irsensor1 int 	`json:"irsensor1"`
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, nil)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(1)
	l := light.GetLight()
	l.GetInfo()
	fmt.Println(l.Addr, l.Power, l.Bright, l.CT)
}

func LightsInfoHandler(w http.ResponseWriter, r *http.Request) {
	dbConn := db.GetDBConn()
	l := light.GetLight()
	var lastData model.EnvData
	dbConn.Last(&lastData)
	if lastData.BtSensor1 == 0 {
		var data model.EnvData
		dbConn.Where("id = ?", lastData.ID-1).First(&data)
		lastData = data
	}
	var stat status
	stat.Ip = l.Addr
	stat.Bright = l.Bright
	stat.Ct = l.CT
	stat.State = l.Power
	stat.Btsensor1 = lastData.BtSensor1
	stat.Btsensor2 = lastData.BtSensor2
	stat.Dissensor1 = lastData.DisSensor1
	stat.Irsensor1 = lastData.IrSensor1
	ret, _ := json.Marshal(stat)
	fmt.Fprint(w, string(ret))
}

func LightsControlHandler(w http.ResponseWriter, r *http.Request) {
	dbConn := db.GetDBConn()
	params := mux.Vars(r)
	id := params["id"]
	fmt.Println(id)
	r.ParseForm()
	form := r.Form
	fmt.Println(form)
	l := light.GetLight()
	action := model.Actions{}
	if form["method"][0] != "toggle" && l.Power == "off" {
		return
	}
	if form["method"][0] == "set_bright" {
		bright, _ := strconv.Atoi(form["value"][0])
		action.Action = "bright"
		action.LightBrightFrom = l.Bright
		action.LightBright = bright
		l.SetBright(bright)
	} else if form["method"][0] == "set_ct" {
		ct, _ := strconv.Atoi(form["value"][0])
		fmt.Println(ct)
		action.Action = "ct"
		action.LightCt = ct
		action.LightCtFrom = l.CT
		resp := l.SetCT(ct)
		fmt.Println(resp)
	} else if form["method"][0] == "toggle" {
		state := form["value"][0]
		fmt.Println(state)
		action.Action = "toggle"
		if state == "on" {
			resp := l.Open()
			action.LightState = 1
			fmt.Print(resp)
		} else if state == "off" {
			resp := l.Close()
			fmt.Print(resp)
			action.LightState = 0
		}
	}
	if _, ok := form["changer"]; ok {
		changer := form["changer"][0]
		action.Changer, _ = strconv.Atoi(changer)
		dbConn.Create(&action)
	}
	fmt.Fprint(w, "SUCCESS")
}