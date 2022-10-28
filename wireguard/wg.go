package wireguard

import (
	"bufio"
	"bytes"
	"net"
	"strconv"
	"strings"
	"time"
)

type Interface struct {
	Name       string
	PublicKey  string
	ListenPort int
	Peers      []*Peer
}
type Peer struct {
	PublicKey       string
	Endpoint        string
	Allowed         []*net.IPNet
	LatestHandshake time.Duration
	transfer        string
	Keepalive       time.Duration
}

func ParseWg(data []byte) (interfaces []*Interface) {
	var lines = bufio.NewScanner(bytes.NewReader(data))
	var tmp *Interface
	var tmpPeer *Peer
	for lines.Scan() {
		line := strings.Split(strings.TrimSpace(lines.Text()), ": ")
		switch line[0] {
		case "interface":
			if tmp != nil {
				interfaces = append(interfaces, tmp)
			}
			tmp = &Interface{Name: line[1]}
		case "public key":
			if tmp == nil {
				return
			}
			tmp.PublicKey = line[1]
		case "listening port":
			if tmp == nil {
				return
			}
			p, _ := strconv.Atoi(line[1])
			tmp.ListenPort = p
		case "peer":
			if tmpPeer != nil {
				if tmp == nil {
					return nil
				}
				tmp.Peers = append(tmp.Peers, tmpPeer)
			}
			tmpPeer = &Peer{PublicKey: line[1]}
		case "endpoint":
			if tmpPeer == nil {
				return
			}
			tmpPeer.Endpoint = line[1]
		case "allowed ips":
			if tmpPeer == nil {
				return
			}
			for _, v := range strings.Split(line[1], ", ") {
				_, m, err := net.ParseCIDR(v)
				if err != nil {
					return
				}
				tmpPeer.Allowed = append(tmpPeer.Allowed, m)
			}
		case "latest handshake":
			if tmpPeer == nil {
				return
			}
			// 解析时间
			var tmpNum int
			for _, v := range strings.Split(line[1], " ") {
				n, err := strconv.Atoi(v)
				if err == nil {
					tmpNum = n
				} else {
					if strings.Contains(v, "minutes") && tmpNum != 0 {
						tmpPeer.LatestHandshake += time.Minute * time.Duration(tmpNum)
						tmpNum = 0
					} else if strings.Contains(v, "seconds") && tmpNum != 0 {
						tmpPeer.LatestHandshake += time.Second * time.Duration(tmpNum)
						tmpNum = 0
					}
				}
			}

		case "transfer":
		case "persistent keepalive":
		}

	}
	if tmp != nil {
		if tmpPeer != nil {
			tmp.Peers = append(tmp.Peers, tmpPeer)
		}
		interfaces = append(interfaces, tmp)
	}
	return
}
