# Examples

This directory contains example programs that demonstrate UDP Logger Relay functionality.

## VarAC Demo

**File:** `varac_demo.go`

Demonstrates VarAC message detection and parsing with various message formats:

```bash
go run varac_demo.go
```

### What it shows:

1. **Full JSON Format** - Complete VarAC UDP broadcast with all fields
2. **Plain Text Format** - Human-readable QSO messages
3. **Simple Completion** - Basic VarAC completion notifications
4. **Minimal JSON** - Reduced field sets from some VarAC configurations
5. **VARA FM** - VHF/UHF VARA FM examples

### Output includes:

- Message type detection
- Parsed QSO data (callsign, frequency, mode, band, RST)
- Generated N1MM Logger Plus XML format
- Error handling for malformed messages

This is useful for:
- Testing VarAC integration before deployment
- Understanding message format variations
- Troubleshooting parsing issues
- Learning the formatter API

## Adding New Examples

To add a new example:

1. Create a new `.go` file in this directory
2. Import the necessary internal packages
3. Add the example binary to `.gitignore`
4. Document it in this README and the main README

Example template:

```go
package main

import (
    "fmt"
    "github.com/akgordon/UDP-Logger-Relay/internal/formatter"
)

func main() {
    fmt.Println("Your Example Demo")
    f := formatter.New("CALL", "OP", "CONTEST")
    // Your demo code here
}
```