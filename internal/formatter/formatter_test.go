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
		{`{"app":"VarAC","call":"W1ABC"}`, MessageTypeVarAC},
		{"VarAC QSO completed", MessageTypeVarAC},
		{"var-ac message", MessageTypeVarAC},
		{"<app>varac</app>", MessageTypeVarAC},
		{`<contactinfo app="N1MM Logger Plus"><call>W1ABC</call></contactinfo>`, MessageTypeN1MM},
		{"<contestname>ARRL-DX</contestname>", MessageTypeN1MM},
		{"<mycall>K1ABC</mycall><band>20m</band>", MessageTypeN1MM},
		{`app="N1MM Logger Plus"`, MessageTypeN1MM},
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

func TestParseVarAC(t *testing.T) {
	formatter := New("TEST", "OP", "GENERAL")

	// Test JSON format VarAC message
	jsonMessage := `{"app":"VarAC","call":"W1ABC","freq":"14.105","mode":"VARA HF","timestamp":"2023-10-12 14:30:00","rst_sent":"599","rst_rcvd":"599","band":"20m"}`
	qso, err := formatter.parseVarAC(jsonMessage)

	if err != nil {
		t.Fatalf("parseVarAC failed: %v", err)
	}

	if qso.Callsign != "W1ABC" {
		t.Errorf("Expected callsign W1ABC, got %s", qso.Callsign)
	}

	if qso.Frequency != "14.105" {
		t.Errorf("Expected frequency 14.105, got %s", qso.Frequency)
	}

	if qso.Mode != "VARA HF" {
		t.Errorf("Expected mode VARA HF, got %s", qso.Mode)
	}

	if qso.Band != "20m" {
		t.Errorf("Expected band 20m, got %s", qso.Band)
	}

	if qso.RST_Sent != "599" {
		t.Errorf("Expected RST sent 599, got %s", qso.RST_Sent)
	}

	if qso.RST_Rcvd != "599" {
		t.Errorf("Expected RST rcvd 599, got %s", qso.RST_Rcvd)
	}

	// Test plain text format
	textMessage := "QSO with VK2XYZ on 14.105 VARA"
	qso2, err := formatter.parseVarAC(textMessage)

	if err != nil {
		t.Fatalf("parseVarAC text format failed: %v", err)
	}

	if qso2.Callsign != "VK2XYZ" {
		t.Errorf("Expected callsign VK2XYZ, got %s", qso2.Callsign)
	}

	if qso2.Frequency != "14.105" {
		t.Errorf("Expected frequency 14.105, got %s", qso2.Frequency)
	}

	if qso2.Mode != "VARA" {
		t.Errorf("Expected mode VARA, got %s", qso2.Mode)
	}

	// Test minimal JSON format
	minimalMessage := `{"call":"EA1ABC","freq":"7.105"}`
	qso3, err := formatter.parseVarAC(minimalMessage)

	if err != nil {
		t.Fatalf("parseVarAC minimal format failed: %v", err)
	}

	if qso3.Callsign != "EA1ABC" {
		t.Errorf("Expected callsign EA1ABC, got %s", qso3.Callsign)
	}

	if qso3.Band != "40m" {
		t.Errorf("Expected band 40m (derived from frequency), got %s", qso3.Band)
	}
}

func TestParseN1MM(t *testing.T) {
	formatter := New("TEST", "OP", "GENERAL")

	// Test full N1MM XML format message
	n1mmMessage := `<contactinfo app="N1MM Logger Plus" timestamp="2023-10-12 14:30:00"><contestname>ARRL-DX-CW</contestname><mycall>W1ABC</mycall><band>20m</band><rxfreq>14.035</rxfreq><txfreq>14.035</txfreq><operator>K1XYZ</operator><mode>CW</mode><call>VK1DEF</call><snt>599</snt><rcv>599</rcv><exchange1>VK</exchange1></contactinfo>`
	qso, err := formatter.parseN1MM(n1mmMessage)

	if err != nil {
		t.Fatalf("parseN1MM failed: %v", err)
	}

	if qso.Callsign != "VK1DEF" {
		t.Errorf("Expected callsign VK1DEF, got %s", qso.Callsign)
	}

	if qso.Frequency != "14.035" {
		t.Errorf("Expected frequency 14.035, got %s", qso.Frequency)
	}

	if qso.Mode != "CW" {
		t.Errorf("Expected mode CW, got %s", qso.Mode)
	}

	if qso.Band != "20m" {
		t.Errorf("Expected band 20m, got %s", qso.Band)
	}

	if qso.RST_Sent != "599" {
		t.Errorf("Expected RST sent 599, got %s", qso.RST_Sent)
	}

	if qso.RST_Rcvd != "599" {
		t.Errorf("Expected RST rcvd 599, got %s", qso.RST_Rcvd)
	}

	if qso.Exchange != "VK" {
		t.Errorf("Expected exchange VK, got %s", qso.Exchange)
	}

	// Test minimal N1MM format
	minimalMessage := `<contactinfo><call>JA1ABC</call><mode>SSB</mode><rxfreq>14.205</rxfreq></contactinfo>`
	qso2, err := formatter.parseN1MM(minimalMessage)

	if err != nil {
		t.Fatalf("parseN1MM minimal format failed: %v", err)
	}

	if qso2.Callsign != "JA1ABC" {
		t.Errorf("Expected callsign JA1ABC, got %s", qso2.Callsign)
	}

	if qso2.Mode != "SSB" {
		t.Errorf("Expected mode SSB, got %s", qso2.Mode)
	}

	if qso2.Band != "20m" {
		t.Errorf("Expected band 20m (derived from frequency), got %s", qso2.Band)
	}

	// Test N1MM format with only txfreq
	txFreqMessage := `<contactinfo><call>EA1ABC</call><txfreq>7.035</txfreq><mode>CW</mode></contactinfo>`
	qso3, err := formatter.parseN1MM(txFreqMessage)

	if err != nil {
		t.Fatalf("parseN1MM txfreq format failed: %v", err)
	}

	if qso3.Frequency != "7.035" {
		t.Errorf("Expected frequency 7.035 (from txfreq), got %s", qso3.Frequency)
	}

	if qso3.Band != "40m" {
		t.Errorf("Expected band 40m (derived from txfreq), got %s", qso3.Band)
	}

	// Test N1MM timestamp parsing - should be in UTC
	timestampMessage := `<contactinfo app="N1MM Logger Plus" timestamp="2025-11-19 01:36:37"><call>WB4WOJ</call><mode>CW</mode><band>14</band></contactinfo>`
	qso4, err := formatter.parseN1MM(timestampMessage)

	if err != nil {
		t.Fatalf("parseN1MM timestamp format failed: %v", err)
	}

	// Verify the timestamp is parsed as UTC
	expectedTime := time.Date(2025, 11, 19, 1, 36, 37, 0, time.UTC)
	if !qso4.DateTime.Equal(expectedTime) {
		t.Errorf("Expected timestamp %v (UTC), got %v (location: %v)", expectedTime, qso4.DateTime, qso4.DateTime.Location())
	}

	// Verify the formatted output preserves UTC time
	formattedXML, err := formatter.FormatForN1MM(qso4)
	if err != nil {
		t.Fatalf("FormatForN1MM failed: %v", err)
	}

	// Check that the output contains the UTC timestamp
	if !strings.Contains(formattedXML, "2025-11-19 01:36:37") {
		t.Errorf("Expected formatted XML to contain UTC timestamp '2025-11-19 01:36:37', got: %s", formattedXML)
	}
}
