package apserver

// #include <dns_sd.h>
// #include <arpa/inet.h>
import "C"

import "errors"
import "fmt"
import "unsafe"

// Keep this around so we can deregister them at the end.
var services []C.DNSServiceRef

// Register a DNS Service of name, service type stype, with txt record specified, on
// a certain port. This gets added to a list of registered services which can be
// deregistered by calling ServiceDeregister(). I'm fairly sure that closing the program
// causes the socket connection to mDNSResponder to be closed, which will also deregister
// the services.
func ServiceRegister(name, stype string, txt map[string]string, port uint16) error {
	txtblob, err := mapToBytes(txt)
	if err != nil {
		return err
	}

	// Network order is big-endian
	bport := (port&0xff)<<8 + port>>8

	var ref C.DNSServiceRef
	dnserr := C.DNSServiceRegister(
		&ref,                        // DNSServiceRegister
		0,                           // DNSServiceFlags
		0,                           // Interface (0 => all interfaces)
		C.CString(name),             // Service name (const char *)
		C.CString(stype),            // Service type (const char *)
		nil,                         // Domain (const char *) empty => default domain
		nil,                         // Host (const char *), empty => machine default host
		C.uint16_t(bport),           // Port, in network byte order
		C.uint16_t(len(txtblob)),    // TXT Record length
		unsafe.Pointer(&txtblob[0]), // TXT Record
		nil, // Callback on register/fail (TODO: do something here)
		nil) // Application Context
	if dnserr != C.kDNSServiceErr_NoError {
		return errors.New("Could not register service")
	}

	services = append(services, ref)
	return nil
}

// Deregister all previously allocated services.
func ServiceDeregister() {
	for _, ref := range services {
		C.DNSServiceRefDeallocate(ref)
	}
}

// Convert a map to the TXT record format: 1 byte (length) followed by
// the character data (string)
func mapToBytes(txt map[string]string) ([]byte, error) {
	blob := make([]byte, 0)
	for key, val := range txt {
		line := key + "=" + val
		if len(line) > 0xff {
			return nil, errors.New(fmt.Sprintf("Line \"%s\" greater than 255 bytes", line))
		}
		blob = append(blob, byte(len(line)))
		blob = append(blob, []byte(line)...)
	}

	return blob, nil
}
