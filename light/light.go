package light

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"
)

type Light struct {
	Addr   string `json:"addr"`
	Power  string `json:"power"`
	Bright int    `json:"bright"`
	CT     int    `json:"ct"`
}

type req struct {
	Id     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type Result struct {
	Id     int      `json:"id"`
	Result []string `json:"result"`
}

var id int
var m sync.Mutex
var rwLock *sync.RWMutex
var light Light
var tcpConn *net.TCPConn

func Init() bool{
	rwLock = new(sync.RWMutex)
	var err error
	err = light.search()
	if err != nil {
		fmt.Println("search light error: ", err)
		return false
	}
	if !connectLight() {
		return false
	}
	go monitorLight()
	fmt.Println("init light finished")
	return true
}

func monitorLight() {
	for {
		light.GetInfo()
		time.Sleep(time.Millisecond*500)
	}
}

func GetLight() Light {
	rwLock.RLock()
	l := light
	rwLock.RUnlock()
	return l
}

func connectLight() bool{
	var err error
	address, _ := net.ResolveTCPAddr("tcp", light.Addr)
	tcpConn, err = net.DialTCP("tcp", nil, address)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (l *Light) search() error {
	address, _ := net.ResolveUDPAddr("udp", "239.255.255.250:1982")
	conn, err := net.ListenMulticastUDP("udp", nil, address)
	if err != nil {
		fmt.Println("listen error: ", err)
		return err
	}
	defer conn.Close()
	msg := "M-SEARCH * HTTP/1.1\r\n"
	msg += "HOST: 239.255.255.250:1982\r\n"
	msg += "MAN: \"ssdp:discover\"\r\n"
	msg += "ST: wifi_bulb"
	_, err = conn.WriteToUDP([]byte(msg), address)
	if err != nil {
		fmt.Println("write to udp error: ", err)
		return err
	}

	data := make([]byte, 0)
	readResp(conn, &data)
	res := string(data)
	l.Addr = getLightAddr(res)
	return nil
}

func (l *Light) Open() Result {
	var r req
	r.Id = getNextID()
	r.Method = "set_power"
	r.Params = []interface{}{"on"}
	resp := setLight(l.Addr, r)
	return resp
}

func (l *Light) Close() Result {
	var r req
	r.Id = getNextID()
	r.Method = "set_power"
	r.Params = []interface{}{"off"}
	resp := setLight(l.Addr, r)
	return resp
}

func (l *Light) GetInfo() {
	var r req
	r.Id = getNextID()
	r.Method = "get_prop"
	r.Params = []interface{}{"power", "bright", "ct"}
	resp := setLight(l.Addr, r)
	if len(resp.Result) != 3 {
		return
	}
	rwLock.Lock()
	l.Power = resp.Result[0]
	l.Bright, _ = strconv.Atoi(resp.Result[1])
	l.CT, _ = strconv.Atoi(resp.Result[2])
	rwLock.Unlock()
}

func (l *Light) SetBright(bright int) Result {
	var r req
	if bright > 100 {
		bright = 100
	}
	if bright < 1 {
		bright = 1
	}
	r.Id = getNextID()
	r.Method = "set_bright"
	r.Params = []interface{}{bright, "smooth", 500}
	resp := setLight(l.Addr, r)
	return resp
}

func (l *Light) SetCT(ct int) Result {
	if ct > 4800 {
		ct = 4800
	}
	if ct < 2500 {
		ct = 2500
	}
	var r req
	r.Id = getNextID()
	r.Method = "set_ct_abx"
	r.Params = []interface{}{ct, "smooth", 500}
	resp := setLight(l.Addr, r)
	return resp
}

func getLightAddr(data string) string {
	reg := regexp.MustCompile(`Location: yeelight://([0-9]{1,3}(\.[0-9]{1,3}){3}):([0-9]*)`)
	match := reg.FindStringSubmatch(string(data))
	return match[1] + ":" + match[3]
}

func setLight(addr string, r req) (resp Result) {
	msg, _ := json.Marshal(r)
	_, err := tcpConn.Write(msg)
	tcpConn.Write([]byte("\r\n"))
	if err != nil {
		fmt.Println(err)
		return
	}
	data := make([]byte, 0)
	readResp(tcpConn, &data)
	_ = json.Unmarshal(data, &resp)
	if len(resp.Result) == 0 {
		time.Sleep(time.Second)
		tcpConn.Close()
		connectLight()
		r.Id = getNextID()
		return setLight(addr, r)
	}
	return resp
}

func getNextID() int {
	m.Lock()
	id += 1
	id %= 1e7
	m.Unlock()
	return id
}

func readResp(conn net.Conn, data *[]byte) {
	const BufLength = 5120
	buf := make([]byte, BufLength)
	for {
		conn.SetReadDeadline(time.Now().Add(time.Millisecond*100))
		n, err := conn.Read(buf)
		//fmt.Println(n)
		if err != nil && err != io.EOF {
			fmt.Println(err)
			return
		}
		*data = append(*data, buf[:n]...)
		if n != BufLength {
			break
		}
	}
}
