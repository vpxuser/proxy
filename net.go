package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
)

func WildcardFQDN(fqdn string) string {
	fqdnSplit := strings.Split(fqdn, ".")
	if len(fqdnSplit) < 2 {
		return fqdn
	}
	return "*." + strings.Join(fqdnSplit[len(fqdnSplit)-2:], ".")
}

func IsDomain(host string) bool {
	matched, err := regexp.MatchString(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`, host)
	if err != nil {
		return false
	}
	return matched
}

func Socks5Handshake(conn net.Conn) (string, string, error) {
	buf := make([]byte, 262)
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return "", "", err
	}

	nMethods := int(buf[1])
	if _, err := io.ReadFull(conn, buf[:nMethods]); err != nil {
		return "", "", err
	}

	if _, err := conn.Write([]byte{0x05, 0x00}); err != nil {
		return "", "", err
	}

	if _, err := io.ReadFull(conn, buf[:4]); err != nil {
		return "", "", err
	}
	if buf[1] != 0x01 {
		conn.Write([]byte{0x05, 0x07, 0x00, 0x01})
		return "", "", fmt.Errorf("socks5 handshake: not supported method %d", buf[1])
	}

	var host, port string
	switch buf[3] {
	case 0x01:
		if _, err := io.ReadFull(conn, buf[:6]); err != nil {
			return "", "", err
		}
		host, port = net.IP(buf[:4]).String(), strconv.Itoa(int(buf[4])<<8|int(buf[5]))
	case 0x03:
		if _, err := io.ReadFull(conn, buf[:1]); err != nil {
			return "", "", err
		}

		domainLen := int(buf[0])
		if _, err := io.ReadFull(conn, buf[:domainLen+2]); err != nil {
			return "", "", err
		}

		host, port = string(buf[:domainLen]), strconv.Itoa(int(binary.BigEndian.Uint16(buf[domainLen:domainLen+2])))
	default:
		return "", "", fmt.Errorf("socks5 handshake: not supported address type %d", buf[3])
	}

	if _, err := conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0}); err != nil {
		return "", "", err
	}

	return host, port, nil
}
