package wemo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func (s *SwitchService) handleSetup(w http.ResponseWriter, r *http.Request) {
	log.Printf("RX: setup request from: %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "text/xml")

	writeSetupXML(w, s.name, s.uuid, s.serial)
}

func writeSetupXML(w io.Writer, name, uuid, serial string) {
	fmt.Fprint(w, `<?xml version="1.0"?>`)
	fmt.Fprint(w, `<root xmlns="urn:Belkin:device-1-0">`)
	fmt.Fprint(w, `<specVersion><major>1</major><minor>0</minor></specVersion>`)
	fmt.Fprint(w, `<device>`)
	fmt.Fprint(w, `<deviceType>urn:Belkin:device:controllee:1</deviceType>`)
	fmt.Fprintf(w, `<friendlyName>%s</friendlyName>`, name)
	fmt.Fprint(w, `<manufacturer>Belkin International Inc.</manufacturer>`)
	fmt.Fprint(w, `<manufacturerURL>http://www.belkin.com</manufacturerURL>`)
	fmt.Fprint(w, `<modelDescription>Belkin Plugin Socket 1.0</modelDescription>`)
	fmt.Fprint(w, `<modelName>Socket</modelName>`)
	fmt.Fprint(w, `<modelNumber>1</modelNumber>`)
	fmt.Fprint(w, `<modelURL>http://www.belkin.com/plugin/</modelURL>`)
	fmt.Fprint(w, `<modelName>Socket</modelName>`)
	fmt.Fprint(w, `<modelNumber>1</modelNumber>`)
	fmt.Fprint(w, `<modelDescription>Belkin Plugin Socket 1.0</modelDescription>`)
	fmt.Fprintf(w, `<UDN>uuid:%s</UDN>`, uuid)
	fmt.Fprintf(w, `<serialNumber>%s</serialNumber>`, serial)
	fmt.Fprint(w, `<binaryState>1</binaryState>`)
	fmt.Fprint(w, `<serviceList>`)
	fmt.Fprint(w, `<service>`)
	fmt.Fprint(w, `<serviceType>urn:Belkin:service:basicevent:1</serviceType>`)
	fmt.Fprint(w, `<serviceId>urn:Belkin:serviceId:basicevent1</serviceId>`)
	fmt.Fprint(w, `<controlURL>/upnp/control/basicevent1</controlURL>`)
	fmt.Fprint(w, `<eventSubURL>/upnp/event/basicevent1</eventSubURL>`)
	fmt.Fprint(w, `<SCPDURL>/eventservice.xml</SCPDURL>`)
	fmt.Fprint(w, `</service>`)
	fmt.Fprint(w, `</serviceList>`)
	fmt.Fprint(w, `</device>`)
	fmt.Fprint(w, `</root>`)
}

func (s *SwitchService) handleUpnpControlBasicEvent1(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	s.mu.Lock()

	w.Header().Set("Content-Type", "text/xml")

	if bytes.Contains(body, []byte("SetBinaryState")) {
		var state bool
		ctx := r.Context()
		switch {
		case bytes.Contains(body, []byte("<BinaryState>1</BinaryState>")):
			log.Printf(`RX: turn on. name="%s"`, s.name)
			state = runCallback(ctx, s.onCallback, s.state)
			sendRelayState(w, state, "Set")
		case bytes.Contains(body, []byte("<BinaryState>0</BinaryState>")):
			log.Printf(`RX: turn off. name="%s"`, s.name)
			state = runCallback(ctx, s.offCallback, s.state)
			sendRelayState(w, state, "Set")
		}
		s.state = state
	}

	if bytes.Contains(body, []byte("GetBinaryState")) {
		log.Printf(`RX: sync binary state. name="%s" state="%t"`, s.name, s.state)
		sendRelayState(w, s.state, "Get")
	}
	s.mu.Unlock()
}

func runCallback(ctx context.Context, command func(ctx context.Context, state bool) bool, state bool) bool {
	if command == nil {
		return !state
	}
	return command(ctx, state)
}

func sendRelayState(w io.Writer, state bool, method string) {
	fmt.Fprint(w, `<?xml version="1.0"?>`)
	fmt.Fprint(w, `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">`)
	fmt.Fprint(w, `<s:Body>`)
	fmt.Fprintf(w, `<u:%sBinaryStateResponse xmlns:u="urn:Belkin:service:basicevent:1">`, method)
	if state {
		fmt.Fprint(w, `<BinaryState>1</BinaryState>`)
	} else {
		fmt.Fprint(w, `<BinaryState>0</BinaryState>`)
	}
	fmt.Fprintf(w, `</u:%sBinaryStateResponse>`, method)
	fmt.Fprint(w, `</s:Body>`)
	fmt.Fprint(w, `</s:Envelope>`)
}

func handleEventService(w http.ResponseWriter, r *http.Request) {
	log.Printf("EventService request from: %s", r.RemoteAddr)

	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprint(w, `<?xml version="1.0"?>
<scpd xmlns="urn:Belkin:service-1-0">
<actionList>
  <action>
    <name>SetBinaryState</name>
    <argumentList>
      <argument>
        <retval/>
        <name>BinaryState</name>
        <relatedStateVariable>BinaryState</relatedStateVariable>
        <direction>in</direction>
      </argument>
    </argumentList>
  </action>
  <action>
    <name>GetBinaryState</name>
    <argumentList>
      <argument>
        <retval/>
        <name>BinaryState</name>
        <relatedStateVariable>BinaryState</relatedStateVariable>
        <direction>out</direction>
      </argument>
    </argumentList>
  </action>
</actionList>
<serviceStateTable>
  <stateVariable sendEvents="yes">
    <name>BinaryState</name>
    <dataType>Boolean</dataType>
    <defaultValue>0</defaultValue>
  </stateVariable>
  <stateVariable sendEvents="yes">
    <name>level</name>
    <dataType>string</dataType>
    <defaultValue>0</defaultValue>
  </stateVariable>
</serviceStateTable>
</scpd>`)
}
