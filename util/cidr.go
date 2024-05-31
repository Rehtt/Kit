package util

import (
	"fmt"
	"math/big"
	"net"
)

func IpToBigInt(ip net.IP) *big.Int {
	ip = ip.To4()
	return big.NewInt(0).SetBytes(ip)
}

func BigIntToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes())
}

// CIDR 计算
func CalculateCIDR(startIP, endIP string) (string, error) {
	start := net.ParseIP(startIP)
	if start == nil {
		return "", fmt.Errorf("invalid start IP address: %s", startIP)
	}
	end := net.ParseIP(endIP)
	if end == nil {
		return "", fmt.Errorf("invalid end IP address: %s", endIP)
	}

	startInt := IpToBigInt(start)
	endInt := IpToBigInt(end)

	diff := big.NewInt(0).Sub(endInt, startInt)
	ones := 32 - diff.BitLen()

	mask := net.CIDRMask(ones, 32)
	network := start.Mask(mask)
	cidr := fmt.Sprintf("%s/%d", network.String(), ones)

	return cidr, nil
}
