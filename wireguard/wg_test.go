package wireguard

import (
	"fmt"
	"testing"
)

const tes1 = `interface: rehtt
  public key: 111
  private key: (hidden)
  listening port: 51820

peer: 222
  preshared key: (hidden)
  endpoint: 1.1.1.1:51820
  allowed ips: 10.3.3.0/24, 192.168.100.1/32
  latest handshake: 1 day, 14 hours, 5 minutes, 46 seconds ago
  transfer: 36.05 MiB received, 2.18 GiB sent
  persistent keepalive: every 25 seconds`

func TestParseWg(t *testing.T) {
	for _, v := range ParseWg([]byte(tes1)) {
		fmt.Println(v.Name, v.ListenPort, v.PublicKey)
		for _, peer := range v.Peers {
			fmt.Printf("%+v\n", peer)
			fmt.Println(peer.Transfer.Received.MB())
		}
		fmt.Println()
	}
}
