package formatter

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// MessageType represents the type of source message
type MessageType string

const (
	MessageTypeWSJTX   MessageType = "wsjt-x"
	MessageTypeFldigi  MessageType = "fldigi"
	MessageTypeJS8Call MessageType = "js8call"
	MessageTypeGeneral MessageType = "general"
)

// QSO represents a QSO record
type QSO struct {
	Callsign  string
	Frequency string
	Mode      string
	RST_Sent  string
	RST_Rcvd  string
	DateTime  time.Time
	Band      string
	Exchange  string
}

// N1MMContactInfo represents the N1MM Logger Plus contact info XML structure
type N1MMContactInfo struct {
	XMLName       xml.Name `xml:"contactinfo"`
	App           string   `xml:"app,attr"`
	Timestamp     string   `xml:"timestamp,attr"`
	Contest       string   `xml:"contestname"`
	Station       string   `xml:"mycall"`
	Band          string   `xml:"band"`
	RXFreq        string   `xml:"rxfreq"`
	TXFreq        string   `xml:"txfreq"`
	Operator      string   `xml:"operator"`
	Mode          string   `xml:"mode"`
	Call          string   `xml:"call"`
	CountryPrefix string   `xml:"countryprefix"`
	WPXPrefix     string   `xml:"wpxprefix"`
	StationPrefix string   `xml:"stationprefix"`
	Continent     string   `xml:"continent"`
	SentNr        string   `xml:"snt"`
	RcvdNr        string   `xml:"rcv"`
	GridSquare    string   `xml:"gridsquare"`
	Exchange      string   `xml:"exchange1"`
	Section       string   `xml:"section"`
	Comment       string   `xml:"comment"`
	Qth           string   `xml:"qth"`
	Name          string   `xml:"name"`
	Power         string   `xml:"power"`
	MiscText      string   `xml:"misctext"`
	Zone          string   `xml:"zone"`
	Prec          string   `xml:"prec"`
	CK            string   `xml:"ck"`
	IsMult1       string   `xml:"ismult1"`
	IsMult2       string   `xml:"ismult2"`
	IsMult3       string   `xml:"ismult3"`
	Points        string   `xml:"points"`
	Radionr       string   `xml:"radionr"`
	RoverLocation string   `xml:"roverlocation"`
	RadioUsed     string   `xml:"RadioUsed"`
}

// Formatter handles message format conversion
type Formatter struct {
	station  string
	operator string
	contest  string
}

// New creates a new formatter instance
func New(station, operator, contest string) *Formatter {
	return &Formatter{
		station:  station,
		operator: operator,
		contest:  contest,
	}
}

// DetectMessageType attempts to detect the source message type
func (f *Formatter) DetectMessageType(message string) MessageType {
	message = strings.ToLower(message)

	// WSJT-X typically sends ADIF-like messages or specific format
	if strings.Contains(message, "wsjt-x") || strings.Contains(message, "<call:") {
		return MessageTypeWSJTX
	}

	// Fldigi might have specific markers
	if strings.Contains(message, "fldigi") {
		return MessageTypeFldigi
	}

	// JS8Call detection
	if strings.Contains(message, "js8call") || strings.Contains(message, "js8") {
		return MessageTypeJS8Call
	}

	return MessageTypeGeneral
}

// ParseMessage attempts to parse the incoming message and extract QSO information
func (f *Formatter) ParseMessage(message string, msgType MessageType) (*QSO, error) {
	switch msgType {
	case MessageTypeWSJTX:
		return f.parseWSJTX(message)
	case MessageTypeFldigi:
		return f.parseFldigi(message)
	case MessageTypeJS8Call:
		return f.parseJS8Call(message)
	default:
		return f.parseGeneral(message)
	}
}

// FormatForN1MM converts a QSO to N1MM Logger Plus XML format
func (f *Formatter) FormatForN1MM(qso *QSO) (string, error) {
	contact := N1MMContactInfo{
		App:       "UDP-Logger-Relay",
		Timestamp: qso.DateTime.Format("2006-01-02 15:04:05"),
		Contest:   f.contest,
		Station:   f.station,
		Band:      qso.Band,
		RXFreq:    qso.Frequency,
		TXFreq:    qso.Frequency,
		Operator:  f.operator,
		Mode:      qso.Mode,
		Call:      qso.Callsign,
		SentNr:    qso.RST_Sent,
		RcvdNr:    qso.RST_Rcvd,
		Exchange:  qso.Exchange,
		Radionr:   "1",
	}

	xmlData, err := xml.MarshalIndent(contact, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal XML: %w", err)
	}

	return string(xmlData), nil
}

// parseWSJTX parses WSJT-X format messages
func (f *Formatter) parseWSJTX(message string) (*QSO, error) {
	// Example WSJT-X ADIF format: <call:6>VK1ABC<band:3>20m<mode:4>FT8<rst_sent:3>-05<rst_rcvd:3>-12<qso_date:8>20231012<time_on:6>123000<eor>
	qso := &QSO{
		DateTime: time.Now(),
	}

	// Parse ADIF-style fields
	callRegex := regexp.MustCompile(`<call:\d+>([A-Z0-9/]+)`)
	if match := callRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Callsign = match[1]
	}

	bandRegex := regexp.MustCompile(`<band:\d+>(\d+m)`)
	if match := bandRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Band = match[1]
	}

	modeRegex := regexp.MustCompile(`<mode:\d+>(\w+)`)
	if match := modeRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Mode = match[1]
	}

	rstSentRegex := regexp.MustCompile(`<rst_sent:\d+>([\-\+]?\d+)`)
	if match := rstSentRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.RST_Sent = match[1]
	}

	rstRcvdRegex := regexp.MustCompile(`<rst_rcvd:\d+>([\-\+]?\d+)`)
	if match := rstRcvdRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.RST_Rcvd = match[1]
	}

	freqRegex := regexp.MustCompile(`<freq:\d+>(\d+\.?\d*)`)
	if match := freqRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Frequency = match[1]
	}

	if qso.Callsign == "" {
		return nil, fmt.Errorf("no callsign found in message")
	}

	return qso, nil
}

// parseFldigi parses Fldigi format messages
func (f *Formatter) parseFldigi(message string) (*QSO, error) {
	// Implement Fldigi-specific parsing logic here
	return f.parseGeneral(message)
}

// parseJS8Call parses JS8Call format messages
func (f *Formatter) parseJS8Call(message string) (*QSO, error) {
	// Implement JS8Call-specific parsing logic here
	return f.parseGeneral(message)
}

// parseGeneral attempts to parse a general format message
func (f *Formatter) parseGeneral(message string) (*QSO, error) {
	// Simple regex-based parsing for common formats
	// This is a fallback parser that tries to extract basic information

	qso := &QSO{
		DateTime: time.Now(),
		Mode:     "DATA", // Default mode
	}

	// Look for callsign pattern (basic ham radio callsign regex)
	callRegex := regexp.MustCompile(`\b([A-Z0-9]{1,3}[0-9][A-Z0-9]{0,3}[A-Z])\b`)
	if match := callRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Callsign = match[1]
	}

	// Look for frequency (MHz format)
	freqRegex := regexp.MustCompile(`(\d+\.?\d*)\s*MHz`)
	if match := freqRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Frequency = match[1]
	}

	// Look for band
	bandRegex := regexp.MustCompile(`(\d+)m\b`)
	if match := bandRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Band = match[1] + "m"
	}

	// Look for mode
	modeRegex := regexp.MustCompile(`\b(FT8|FT4|PSK31|RTTY|CW|SSB|LSB|USB|AM|FM)\b`)
	if match := modeRegex.FindStringSubmatch(strings.ToUpper(message)); len(match) > 1 {
		qso.Mode = match[1]
	}

	if qso.Callsign == "" {
		return nil, fmt.Errorf("no callsign found in message: %s", message)
	}

	return qso, nil
}

// FrequencyToBand converts frequency in MHz to amateur band designation
func FrequencyToBand(freqMHz float64) string {
	switch {
	case freqMHz >= 1.8 && freqMHz <= 2.0:
		return "160m"
	case freqMHz >= 3.5 && freqMHz <= 4.0:
		return "80m"
	case freqMHz >= 7.0 && freqMHz <= 7.3:
		return "40m"
	case freqMHz >= 14.0 && freqMHz <= 14.35:
		return "20m"
	case freqMHz >= 21.0 && freqMHz <= 21.45:
		return "15m"
	case freqMHz >= 28.0 && freqMHz <= 29.7:
		return "10m"
	case freqMHz >= 50.0 && freqMHz <= 54.0:
		return "6m"
	case freqMHz >= 144.0 && freqMHz <= 148.0:
		return "2m"
	case freqMHz >= 420.0 && freqMHz <= 450.0:
		return "70cm"
	default:
		return "UNK"
	}
}
