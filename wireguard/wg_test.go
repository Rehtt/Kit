package wireguard

import (
	"fmt"
	"testing"
)

const tes1 = `interface: rehtt
  public key: 111111111111111111111111111111111
  private key: (hidden)
  listening port: 51820

peer: 111111111111111111111111111111111
  preshared key: (hidden)
  endpoint: 1.1.1.1:51820
  allowed ips: 10.3.3.0/24, 192.168.100.1/32
  latest handshake: 2 minutes, 2 seconds ago
  transfer: 108.62 KiB received, 82.80 KiB sent
  persistent keepalive: every 25 seconds`

func TestParseWg(t *testing.T) {
	for _, v := range ParseWg([]byte(tes1)) {
		fmt.Println(v.Name, v.ListenPort, v.PublicKey)
		for _, peer := range v.Peers {
			fmt.Println(peer.PublicKey, peer.LatestHandshake, peer.Allowed)
		}
		fmt.Println()
	}
}
