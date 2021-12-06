package wemo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func startUPnPService(ctx context.Context, services ...*SwitchService) error {
	serverAddr, err := net.ResolveUDPAddr("udp", "239.255.255.250:1900")
	if err != nil {
		return fmt.Errorf("resolve UDP Addr: %s", err)
	}

	serverConn, err := net.ListenMulticastUDP("udp", nil, serverAddr)
	if err != nil {
		return fmt.Errorf("listen muiticast UDP: %s", err)
	}

	go func() {
		<-ctx.Done()
		serverConn.Close()
	}()

	buf := make([]byte, 1024)

	for {
		n, addr, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return context.Canceled
			}
			return fmt.Errorf("read from UDP: %s", err)
		}

		body := buf[0:n]

		if bytes.Contains(body, []byte("M-SEARCH")) {
			if bytes.Contains(body, []byte("urn:Belkin:device:**")) ||
				bytes.Contains(body, []byte("ssdp:all")) ||
				bytes.Contains(body, []byte("upnp:rootdevice")) {

				log.Printf("RX: UDP belkin UPnP request from: %s", addr)

				for _, s := range services {
					var d net.Dialer
					conn, err := d.DialContext(ctx, "udp", addr.String())
					if err == nil {
						localAddr := conn.LocalAddr().(*net.UDPAddr)
						if err := writeDeviceSearchResponse(conn, localAddr.IP.String(), s.port, s.uuid); err != nil {
							log.Printf("failed to write response: %s", err)
						}
						conn.Close()
					}
				}
			}
		}
	}
}

func writeDeviceSearchResponse(w io.Writer, host, port, id string) error {
	b := &bytes.Buffer{}
	fmt.Fprint(b, "HTTP/1.1 200 OK\r\n")
	fmt.Fprint(b, "CACHE-CONTROL: max-age=86400\r\n")
	fmt.Fprint(b, "DATE: Sat, 26 Nov 2016 04:56:29 GMT\r\n")
	fmt.Fprint(b, "EXT:\r\n")
	fmt.Fprintf(b, "LOCATION: http://%s:%s/setup.xml\r\n", host, port)
	fmt.Fprint(b, `OPT: "http://schemas.upnp.org/upnp/1/0/"; ns=01`)
	fmt.Fprint(b, "\r\n")
	fmt.Fprintf(b, "01-NLS: %s\r\n", id)
	fmt.Fprint(b, "SERVER: Unspecified, UPnP/1.0, Unspecified\r\n")
	fmt.Fprint(b, "ST: urn:Belkin:device:**\r\n")
	fmt.Fprintf(b, "USN: uuid:%s::urn:Belkin:device:controllee:1\r\n", id)
	fmt.Fprint(b, "X-User-Agent: redsonic\r\n\r\n")
	_, err := w.Write(b.Bytes())
	return err
}
