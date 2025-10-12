package formatter

import (
	"strings"
	"testing"
	"time"
)

func TestDetectMessageType(t *testing.T) {
	formatter := New("TEST", "OP", "GENERAL")

	tests := []struct {
		message  string
		expected MessageType
	}{
		{"<call:6>VK1ABC<mode:3>FT8<eor>", MessageTypeWSJTX},
		{"WSJT-X message here", MessageTypeWSJTX},
		{"fldigi message", MessageTypeFldigi},
		{"js8call data", MessageTypeJS8Call},
		{"some random message", MessageTypeGeneral},
	}

	for _, test := range tests {
		result := formatter.DetectMessageType(test.message)
		if result != test.expected {
			t.Errorf("DetectMessageType(%s) = %s; expected %s", test.message, result, test.expected)
		}
	}
}

func TestParseWSJTX(t *testing.T) {
	formatter := New("TEST", "OP", "GENERAL")

	message := "<call:6>VK1ABC<band:3>20m<mode:3>FT8<rst_sent:3>-05<rst_rcvd:3>-12<eor>"
	qso, err := formatter.parseWSJTX(message)

	if err != nil {
		t.Fatalf("parseWSJTX failed: %v", err)
	}

	if qso.Callsign != "VK1ABC" {
		t.Errorf("Expected callsign VK1ABC, got %s", qso.Callsign)
	}

	if qso.Band != "20m" {
		t.Errorf("Expected band 20m, got %s", qso.Band)
	}

	if qso.Mode != "FT8" {
		t.Errorf("Expected mode FT8, got %s", qso.Mode)
	}

	if qso.RST_Sent != "-05" {
		t.Errorf("Expected RST sent -05, got %s", qso.RST_Sent)
	}

	if qso.RST_Rcvd != "-12" {
		t.Errorf("Expected RST rcvd -12, got %s", qso.RST_Rcvd)
	}
}

func TestParseGeneral(t *testing.T) {
	formatter := New("TEST", "OP", "GENERAL")

	message := "QSO with VK1ABC on 14.074 MHz FT8 mode"
	qso, err := formatter.parseGeneral(message)

	if err != nil {
		t.Fatalf("parseGeneral failed: %v", err)
	}

	if qso.Callsign != "VK1ABC" {
		t.Errorf("Expected callsign VK1ABC, got %s", qso.Callsign)
	}

	if qso.Frequency != "14.074" {
		t.Errorf("Expected frequency 14.074, got %s", qso.Frequency)
	}

	if qso.Mode != "FT8" {
		t.Errorf("Expected mode FT8, got %s", qso.Mode)
	}
}

func TestFormatForN1MM(t *testing.T) {
	formatter := New("W1AW", "K1ABC", "TEST-CONTEST")

	qso := &QSO{
		Callsign:  "VK1ABC",
		Frequency: "14.074",
		Mode:      "FT8",
		RST_Sent:  "-05",
		RST_Rcvd:  "-12",
		DateTime:  time.Date(2023, 10, 12, 14, 30, 0, 0, time.UTC),
		Band:      "20m",
		Exchange:  "59 VK",
	}

	n1mmXML, err := formatter.FormatForN1MM(qso)
	if err != nil {
		t.Fatalf("FormatForN1MM failed: %v", err)
	}

	// Check that XML contains expected elements
	if !strings.Contains(n1mmXML, "<call>VK1ABC</call>") {
		t.Error("XML should contain callsign")
	}

	if !strings.Contains(n1mmXML, "<mycall>W1AW</mycall>") {
		t.Error("XML should contain station call")
	}

	if !strings.Contains(n1mmXML, "<contestname>TEST-CONTEST</contestname>") {
		t.Error("XML should contain contest name")
	}

	if !strings.Contains(n1mmXML, "<mode>FT8</mode>") {
		t.Error("XML should contain mode")
	}
}

func TestFrequencyToBand(t *testing.T) {
	tests := []struct {
		freq float64
		band string
	}{
		{1.85, "160m"},
		{3.7, "80m"},
		{7.1, "40m"},
		{14.2, "20m"},
		{21.2, "15m"},
		{28.5, "10m"},
		{52.0, "6m"},
		{146.0, "2m"},
		{435.0, "70cm"},
		{999.0, "UNK"},
	}

	for _, test := range tests {
		result := FrequencyToBand(test.freq)
		if result != test.band {
			t.Errorf("FrequencyToBand(%.2f) = %s; expected %s", test.freq, result, test.band)
		}
	}
}
