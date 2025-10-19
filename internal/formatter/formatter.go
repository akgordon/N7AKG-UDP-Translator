package formatter

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// MessageType represents the type of source message
type MessageType string

const (
	MessageTypeWSJTX   MessageType = "wsjt-x"
	MessageTypeFldigi  MessageType = "fldigi"
	MessageTypeJS8Call MessageType = "js8call"
	MessageTypeVarAC   MessageType = "varac"
	MessageTypeN1MM    MessageType = "n1mm"
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
	messageLower := strings.ToLower(message)

	// Filter out obvious binary protocol messages (contain significant non-printable characters)
	nonPrintableCount := 0
	for _, b := range []byte(message) {
		if b < 32 && b != 9 && b != 10 && b != 13 { // Exclude tab, LF, CR
			nonPrintableCount++
		}
	}
	// If more than 10% of the message is non-printable, likely binary protocol
	if len(message) > 0 && float64(nonPrintableCount)/float64(len(message)) > 0.1 {
		return MessageTypeGeneral // Will be ignored
	}

	// N1MM detection - N1MM Logger Plus sends XML contactinfo messages (check first as it's most specific)
	if strings.Contains(messageLower, "<contactinfo") || strings.Contains(messageLower, "<contestname>") ||
		strings.Contains(messageLower, "<mycall>") || strings.Contains(messageLower, "n1mm") ||
		(strings.Contains(messageLower, "app=") && strings.Contains(messageLower, "n1mm")) {
		return MessageTypeN1MM
	}

	// VarAC detection - VarAC can send ADIF format or JSON format
	if strings.Contains(messageLower, "varac") || strings.Contains(messageLower, "var-ac") ||
		strings.Contains(messageLower, "\"app\":\"varac\"") || strings.Contains(messageLower, "<app>varac</app>") ||
		(strings.Contains(messageLower, "<mode:") && strings.Contains(messageLower, "vara")) ||
		(strings.Contains(messageLower, "<submode:") && strings.Contains(messageLower, "vara")) ||
		(strings.Contains(messageLower, "{") && strings.Contains(messageLower, "\"call\"") && strings.Contains(messageLower, "\"freq")) {
		return MessageTypeVarAC
	}

	// WSJT-X sends both binary protocol messages and ADIF log messages
	// Only process ADIF log messages (which contain proper ADIF field tags)
	// Binary protocol messages should be ignored even if they contain "WSJT-X"
	if (strings.Contains(messageLower, "<call:") && !strings.Contains(messageLower, "vara")) ||
		(strings.Contains(messageLower, "wsjt-x") && strings.Contains(messageLower, "<") && strings.Contains(messageLower, ":") && strings.Contains(messageLower, ">")) {
		return MessageTypeWSJTX
	}

	// Fldigi might have specific markers
	if strings.Contains(messageLower, "fldigi") {
		return MessageTypeFldigi
	}

	// JS8Call detection
	if strings.Contains(messageLower, "js8call") || strings.Contains(messageLower, "js8") {
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
	case MessageTypeVarAC:
		return f.parseVarAC(message)
	case MessageTypeN1MM:
		return f.parseN1MM(message)
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

	// Check if this is a binary protocol message (contains non-printable characters)
	// Binary messages should be ignored, not parsed as QSOs
	for _, b := range []byte(message) {
		if b < 32 && b != 9 && b != 10 && b != 13 { // Allow tab, LF, CR
			return nil, fmt.Errorf("binary protocol message detected, ignoring")
		}
	}

	// Also check if message lacks proper ADIF structure
	if !strings.Contains(message, "<call:") && !strings.Contains(message, "<CALL:") {
		return nil, fmt.Errorf("not a valid ADIF QSO message")
	}

	qso := &QSO{
		DateTime: time.Now(), // Default fallback
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

	// Parse date and time fields
	qsoDateRegex := regexp.MustCompile(`<qso_date:\d+>(\d{8})`)
	timeOnRegex := regexp.MustCompile(`<time_on:\d+>(\d{4,6})`)

	var qsoDate, timeOn string
	if match := qsoDateRegex.FindStringSubmatch(message); len(match) > 1 {
		qsoDate = match[1]
	}
	if match := timeOnRegex.FindStringSubmatch(message); len(match) > 1 {
		timeOn = match[1]
	}

	// If we have both date and time, parse them
	if qsoDate != "" && timeOn != "" {
		// Ensure time is 6 digits (HHMMSS), pad with zeros if needed
		if len(timeOn) == 4 {
			timeOn = timeOn + "00" // Add seconds if only HHMM
		} else if len(timeOn) == 5 {
			timeOn = timeOn + "0" // Add final second digit if only HHMMS
		}

		dateTimeStr := qsoDate + timeOn
		if len(dateTimeStr) >= 14 { // YYYYMMDDHHMMSS
			if t, err := time.Parse("20060102150405", dateTimeStr); err == nil {
				qso.DateTime = t
			}
		}
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

// parseVarAC parses VarAC format messages (both ADIF and JSON formats)
func (f *Formatter) parseVarAC(message string) (*QSO, error) {
	// VarAC can send messages in two formats:
	// 1. ADIF format: <command:3>Log<parameters:267><CALL:5>n7akg <MODE:7>DYNAMIC <SUBMODE:7>VARA HF...
	// 2. JSON format: {"app":"VarAC","call":"W1ABC","freq":"14.105","mode":"VARA"...}

	qso := &QSO{
		DateTime: time.Now(),
		Mode:     "VARA", // Default VarAC mode
	}

	// Check if it's ADIF format (contains ADIF field tags like <CALL:5>)
	if strings.Contains(message, "<CALL:") && strings.Contains(message, "<EOR>") {
		return f.parseADIF(message)
	}

	// Parse JSON-like format
	if strings.Contains(message, "{") && strings.Contains(message, "}") {
		// Extract callsign
		callRegex := regexp.MustCompile(`"call"\s*:\s*"([A-Z0-9/]+)"`)
		if match := callRegex.FindStringSubmatch(message); len(match) > 1 {
			qso.Callsign = match[1]
		}

		// Extract frequency
		freqRegex := regexp.MustCompile(`"freq(?:uency)?"\s*:\s*"?(\d+\.?\d*)"?`)
		if match := freqRegex.FindStringSubmatch(message); len(match) > 1 {
			qso.Frequency = match[1]
		}

		// Extract mode
		modeRegex := regexp.MustCompile(`"mode"\s*:\s*"([^"]+)"`)
		if match := modeRegex.FindStringSubmatch(message); len(match) > 1 {
			qso.Mode = match[1]
		}

		// Extract band
		bandRegex := regexp.MustCompile(`"band"\s*:\s*"([^"]+)"`)
		if match := bandRegex.FindStringSubmatch(message); len(match) > 1 {
			qso.Band = match[1]
		}

		// Extract RST sent
		rstSentRegex := regexp.MustCompile(`"rst_sent"\s*:\s*"([^"]+)"`)
		if match := rstSentRegex.FindStringSubmatch(message); len(match) > 1 {
			qso.RST_Sent = match[1]
		}

		// Extract RST received
		rstRcvdRegex := regexp.MustCompile(`"rst_r(?:cvd|eceived)"\s*:\s*"([^"]+)"`)
		if match := rstRcvdRegex.FindStringSubmatch(message); len(match) > 1 {
			qso.RST_Rcvd = match[1]
		}

		// Extract timestamp if available
		timestampRegex := regexp.MustCompile(`"timestamp"\s*:\s*"([^"]+)"`)
		if match := timestampRegex.FindStringSubmatch(message); len(match) > 1 {
			// Try to parse the timestamp
			if t, err := time.Parse("2006-01-02 15:04:05", match[1]); err == nil {
				qso.DateTime = t
			} else if t, err := time.Parse("2006-01-02T15:04:05Z", match[1]); err == nil {
				qso.DateTime = t
			}
		}
	} else {
		// Fallback to text parsing for non-JSON VarAC messages
		// VarAC might also send plain text messages like "QSO with W1ABC on 14.105 VARA"

		// Look for callsign pattern (multiple formats)
		callRegex := regexp.MustCompile(`(?i)(?:qso\s+(?:with\s+|completed\s+with\s+)|call[:\s]+)([A-Z0-9/]+)`)
		if match := callRegex.FindStringSubmatch(message); len(match) > 1 {
			qso.Callsign = strings.ToUpper(match[1])
		} else {
			// Fallback: look for any valid callsign in the message
			fallbackRegex := regexp.MustCompile(`\b([A-Z0-9]{1,3}[0-9][A-Z0-9]{0,3}[A-Z])\b`)
			if match := fallbackRegex.FindStringSubmatch(strings.ToUpper(message)); len(match) > 1 {
				qso.Callsign = match[1]
			}
		}

		// Look for frequency (more specific pattern to avoid matching callsign numbers)
		freqRegex := regexp.MustCompile(`(?:on\s+|@\s+|freq[:\s]+)(\d+\.?\d*)\s*(?:MHz|khz)?`)
		if match := freqRegex.FindStringSubmatch(message); len(match) > 1 {
			qso.Frequency = match[1]
		} else {
			// Fallback: look for standalone frequency
			freqRegex2 := regexp.MustCompile(`\b(\d{1,2}\.\d{3})\b`)
			if match := freqRegex2.FindStringSubmatch(message); len(match) > 1 {
				qso.Frequency = match[1]
			}
		}

		// Look for VARA mode indicators
		if strings.Contains(strings.ToUpper(message), "VARA") {
			if strings.Contains(strings.ToUpper(message), "VARA HF") {
				qso.Mode = "VARA HF"
			} else if strings.Contains(strings.ToUpper(message), "VARA FM") {
				qso.Mode = "VARA FM"
			} else {
				qso.Mode = "VARA"
			}
		}
	}

	// If we have frequency but no band, derive the band
	if qso.Frequency != "" && qso.Band == "" {
		if freq, err := strconv.ParseFloat(qso.Frequency, 64); err == nil {
			qso.Band = FrequencyToBand(freq)
		}
	}

	// Set default RST if not provided
	if qso.RST_Sent == "" {
		qso.RST_Sent = "599"
	}
	if qso.RST_Rcvd == "" {
		qso.RST_Rcvd = "599"
	}

	if qso.Callsign == "" {
		return nil, fmt.Errorf("no callsign found in VarAC message: %s", message)
	}

	return qso, nil
}

// parseADIF parses ADIF format messages (used by VarAC and others)
func (f *Formatter) parseADIF(message string) (*QSO, error) {
	qso := &QSO{
		DateTime: time.Now(),
	}

	// Extract ADIF fields using regex
	// ADIF format: <FIELD:length>value
	adifFields := make(map[string]string)

	// Regex to match ADIF field format: <FIELD_NAME:length>value
	fieldRegex := regexp.MustCompile(`<([A-Z_]+):(\d+)>([^<]{0,})`)
	matches := fieldRegex.FindAllStringSubmatch(message, -1)

	for _, match := range matches {
		if len(match) >= 4 {
			fieldName := match[1]
			lengthStr := match[2]
			value := match[3]

			// Parse the length and extract the correct amount of characters
			if length, err := strconv.Atoi(lengthStr); err == nil && len(value) >= length {
				adifFields[fieldName] = value[:length]
			}
		}
	}

	// Map ADIF fields to QSO struct
	if call, exists := adifFields["CALL"]; exists {
		qso.Callsign = strings.ToUpper(call)
	}

	if mode, exists := adifFields["MODE"]; exists {
		qso.Mode = mode
	} else if submode, exists := adifFields["SUBMODE"]; exists {
		qso.Mode = submode
	}

	if band, exists := adifFields["BAND"]; exists {
		qso.Band = band
	}

	if rstSent, exists := adifFields["RST_SENT"]; exists {
		qso.RST_Sent = rstSent
	}

	if rstRcvd, exists := adifFields["RST_RCVD"]; exists {
		qso.RST_Rcvd = rstRcvd
	}

	// Parse date and time
	if qsoDate, dateExists := adifFields["QSO_DATE"]; dateExists {
		if timeOn, timeExists := adifFields["TIME_ON"]; timeExists {
			// ADIF date format: YYYYMMDD, time format: HHMMSS
			dateTimeStr := qsoDate + timeOn
			if len(dateTimeStr) >= 13 { // YYYYMMDDHHMMSS
				if t, err := time.Parse("20060102150405", dateTimeStr); err == nil {
					qso.DateTime = t
				}
			} else if len(dateTimeStr) >= 11 { // YYYYMMDDHHMM
				if t, err := time.Parse("200601021504", dateTimeStr); err == nil {
					qso.DateTime = t
				}
			}
		}
	}

	// Set defaults if not provided
	if qso.RST_Sent == "" {
		qso.RST_Sent = "+00"
	}
	if qso.RST_Rcvd == "" {
		qso.RST_Rcvd = "+00"
	}

	// If we don't have a band but we have frequency, try to derive it
	if qso.Band == "" && qso.Frequency != "" {
		if freq, err := strconv.ParseFloat(qso.Frequency, 64); err == nil {
			qso.Band = FrequencyToBand(freq)
		}
	}

	if qso.Callsign == "" {
		return nil, fmt.Errorf("no callsign found in ADIF message")
	}

	return qso, nil
}

// parseN1MM parses N1MM Logger Plus XML format messages
func (f *Formatter) parseN1MM(message string) (*QSO, error) {
	// N1MM Logger Plus sends XML contactinfo messages
	// Example: <contactinfo app="N1MM Logger Plus" timestamp="2023-10-12 14:30:00"><contestname>GENERAL</contestname><mycall>W1ABC</mycall><band>20m</band><rxfreq>14.074</rxfreq><call>VK1DEF</call><mode>FT8</mode><snt>-05</snt><rcv>-12</rcv></contactinfo>

	qso := &QSO{
		DateTime: time.Now(),
	}

	// Parse XML-like format using regex (lighter than full XML parsing for this use case)

	// Extract callsign (the contacted station)
	callRegex := regexp.MustCompile(`<call>([^<]+)</call>`)
	if match := callRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Callsign = strings.TrimSpace(match[1])
	}

	// Extract frequency
	freqRegex := regexp.MustCompile(`<rxfreq>([^<]+)</rxfreq>`)
	if match := freqRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Frequency = strings.TrimSpace(match[1])
	}
	// Fallback to txfreq if rxfreq not found
	if qso.Frequency == "" {
		txFreqRegex := regexp.MustCompile(`<txfreq>([^<]+)</txfreq>`)
		if match := txFreqRegex.FindStringSubmatch(message); len(match) > 1 {
			qso.Frequency = strings.TrimSpace(match[1])
		}
	}

	// Extract mode
	modeRegex := regexp.MustCompile(`<mode>([^<]+)</mode>`)
	if match := modeRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Mode = strings.TrimSpace(match[1])
	}

	// Extract band
	bandRegex := regexp.MustCompile(`<band>([^<]+)</band>`)
	if match := bandRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Band = strings.TrimSpace(match[1])
	}

	// Extract RST sent (N1MM uses <snt> tag)
	rstSentRegex := regexp.MustCompile(`<snt>([^<]+)</snt>`)
	if match := rstSentRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.RST_Sent = strings.TrimSpace(match[1])
	}

	// Extract RST received (N1MM uses <rcv> tag)
	rstRcvdRegex := regexp.MustCompile(`<rcv>([^<]+)</rcv>`)
	if match := rstRcvdRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.RST_Rcvd = strings.TrimSpace(match[1])
	}

	// Extract timestamp if available
	timestampRegex := regexp.MustCompile(`timestamp="([^"]+)"`)
	if match := timestampRegex.FindStringSubmatch(message); len(match) > 1 {
		// Try to parse the timestamp
		if t, err := time.Parse("2006-01-02 15:04:05", match[1]); err == nil {
			qso.DateTime = t
		} else if t, err := time.Parse("2006-01-02T15:04:05Z", match[1]); err == nil {
			qso.DateTime = t
		}
	}

	// Extract exchange information
	exchangeRegex := regexp.MustCompile(`<exchange1?>([^<]+)</exchange1?>`)
	if match := exchangeRegex.FindStringSubmatch(message); len(match) > 1 {
		qso.Exchange = strings.TrimSpace(match[1])
	}

	// If we have frequency but no band, derive the band
	if qso.Frequency != "" && qso.Band == "" {
		if freq, err := strconv.ParseFloat(qso.Frequency, 64); err == nil {
			qso.Band = FrequencyToBand(freq)
		}
	}

	// Set default RST if not provided
	if qso.RST_Sent == "" {
		qso.RST_Sent = "599"
	}
	if qso.RST_Rcvd == "" {
		qso.RST_Rcvd = "599"
	}

	if qso.Callsign == "" {
		return nil, fmt.Errorf("no callsign found in N1MM message: %s", message)
	}

	return qso, nil
}

// parseGeneral attempts to parse a general format message
func (f *Formatter) parseGeneral(message string) (*QSO, error) {
	// Immediately reject binary protocol messages to reduce spam
	for _, b := range []byte(message) {
		if b < 32 && b != 9 && b != 10 && b != 13 { // Allow tab, LF, CR
			return nil, fmt.Errorf("binary protocol message, ignoring")
		}
	}

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
