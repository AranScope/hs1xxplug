package hs1xxplug

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type Hs1xxPlug struct {
	IPAddress string
}

func (p *Hs1xxPlug) TurnOn() (err error) {
	js := `{"system":{"set_relay_state":{"state":1}}}`
	data := encrypt(js)
	_, err = send(p.IPAddress, data)
	return
}

func (p *Hs1xxPlug) TurnOff() (err error) {
	js := `{"system":{"set_relay_state":{"state":0}}}`
	data := encrypt(js)
	_, err = send(p.IPAddress, data)
	return
}

func (p *Hs1xxPlug) SystemInfo() (results string, err error) {
	js := `{"system":{"get_sysinfo":{}}}`
	data := encrypt(js)
	reading, err := send(p.IPAddress, data)
	if err == nil {
		results = decrypt(reading[4:])
	}
	return
}

func (p *Hs1xxPlug) MeterInfo() (results string, err error) {
	js := `{"system":{"get_sysinfo":{}}, "emeter":{"get_realtime":{},"get_vgain_igain":{}}}`
	data := encrypt(js)
	reading, err := send(p.IPAddress, data)
	if err == nil {
		results = decrypt(reading[4:])
	}
	return
}

func (p *Hs1xxPlug) DailyStats(month int, year int) (results string, err error) {
	js := fmt.Sprintf(`{"emeter":{"get_daystat":{"month":%d,"year":%d}}}`, month, year)
	data := encrypt(js)
	reading, err := send(p.IPAddress, data)
	if err == nil {
		results = decrypt(reading[4:])
	}
	return
}

func encrypt(plaintext string) []byte {
	n := len(plaintext)
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, uint32(n))
	ciphertext := []byte(buf.Bytes())

	key := byte(0xAB)
	payload := make([]byte, n)
	for i := 0; i < n; i++ {
		payload[i] = plaintext[i] ^ key
		key = payload[i]
	}

	for i := 0; i < len(payload); i++ {
		ciphertext = append(ciphertext, payload[i])
	}

	return ciphertext
}

func decrypt(ciphertext []byte) string {
	n := len(ciphertext)
	key := byte(0xAB)
	var nextKey byte
	for i := 0; i < n; i++ {
		nextKey = ciphertext[i]
		ciphertext[i] = ciphertext[i] ^ key
		key = nextKey
	}
	return string(ciphertext)
}

func send(ip string, payload []byte) (data []byte, err error) {
	conn, err := net.DialTimeout("tcp", ip+":9999", time.Duration(10)*time.Second)
	if err != nil {
		fmt.Println("Cannot connect to plug:", err)
		data = nil
		return
	}
	_, err = conn.Write(payload)

	buff := make([]byte, 2048)
	n, err := conn.Read(buff)
	data = buff[:n]
	if err != nil {
		fmt.Println("Cannot read data from plug:", err)
	}
	_ = conn.Close()
	return
}
