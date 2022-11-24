package mcom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parsePrinterName(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(parsePrinterName("device for PDF: cups-pdf:/"), "PDF")
	assert.Equal(parsePrinterName("device for test: socket://111.1.1.11"), "test")
	assert.Equal(parsePrinterName("device for test_dnssd: dnssd://Canon%20MF230._ipp._tcp.local/?uuid=00000000-ffff-1111-8888-74bfc01e784d"), "test_dnssd")
}
