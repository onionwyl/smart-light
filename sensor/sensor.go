package sensor

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/onionwyl/smart-light/db"
	"github.com/onionwyl/smart-light/light"
	"github.com/onionwyl/smart-light/model"
	"net"
	"strings"
)

type sensorJson struct {
	Msg float32 `json:"msg"`
	Sensor string `json:"sensor"`
}

func Init() bool {
	dbConn := db.GetDBConn()
	listener, err := net.Listen("tcp", ":8010")
	if err != nil {
		fmt.Println("error listen:", err)
		return false
	}
	go monitorSensor(listener, dbConn)
	return true
}

func monitorSensor(listener net.Listener, dbConn *gorm.DB) {
	defer listener.Close()
	fmt.Println("listen ok")
	for{
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			break
		}
		go handleSensor(conn, dbConn)
	}
}

func handleSensor(conn net.Conn, dbConn *gorm.DB) {
	defer conn.Close()
	addr := conn.RemoteAddr()

	fmt.Println("Receive message from nodemcu:", addr)
	buffer := make([]byte, 2048)
	for {
		n, err := conn.Write([]byte("ok"))
		if err != nil {
			fmt.Println("conn disconnect ", err)
			break
		}
		n, err = conn.Read(buffer)
		if err != nil {
			fmt.Println("conn read error: ", err)
			break
		}
		fmt.Printf("receive %d bytes data: %s\n", n, string(buffer[:n]))
		envData := model.EnvData{}
		data := string(buffer[:n])
		for{
			pos := strings.Index(data, "}")
			sensorData := data[:pos+1]
			var sensor sensorJson
			json.Unmarshal([]byte(sensorData), &sensor)
			switch sensor.Sensor {
			case "SR3":
				envData.IrSensor1 = int(sensor.Msg)
			case "HC2":
				envData.DisSensor1 = sensor.Msg
			case "GY2":
				envData.BtSensor2 = int(sensor.Msg)
			case "GY1":
				envData.BtSensor1 = int(sensor.Msg)
			default:
				fmt.Println("default")
				return
			}
			if pos == len(data) - 1 {
				break
			}
			data = data[pos+1:]
		}
		if envData.BtSensor1 != 0 {
			var lastData model.EnvData
			dbConn.Last(&lastData)
			dbConn.Model(&lastData).Update("bt_sensor1", envData.BtSensor1)
		} else {
			var count int
			dbConn.Table("env_data").Count(&count)
			if count == 0 {
				dbConn.Create(&envData)
				continue
			}
			var lastData model.EnvData
			dbConn.Last(&lastData)
			if lastData.BtSensor1 == 0 {
				continue
			}
			l := light.GetLight()
			fmt.Println(l)
			if l.Power == "on" {
				envData.LightState = 1
			} else {
				envData.LightState = 0
			}
			envData.LightBright = l.Bright
			envData.LightCT = l.CT
			dbConn.Create(&envData)
		}
	}
}