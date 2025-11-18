# WSJT-X UDP Simulator

A testing tool that simulates WSJT-X UDP broadcasts to test the UDP Logger Relay application.

## Purpose

This tool is designed to help debug and test the UDP Logger Relay by generating realistic WSJT-X-style UDP messages. Use this when you want to verify that your relay is properly listening for and processing UDP packets without needing to run the actual WSJT-X application.

## Building

From the root of the project:

```powershell
go build -o tools/wsjtx_simulator.exe tools/wsjtx_simulator.go
```

Or from the tools directory:

```powershell
cd tools
go build -o wsjtx_simulator.exe wsjtx_simulator.go
```

## Usage

Basic usage (sends to localhost:2333):

```powershell
.\wsjtx_simulator.exe
```

### Command Line Options

- `-addr` - Target address (default: "127.0.0.1")
- `-port` - Target port (default: 2333)
- `-interval` - Seconds between messages (default: 5)
- `-mode` - Operating mode (default: "FT8")
- `-callsign` - Callsign to use (default: "W1TEST")
- `-frequency` - Frequency in Hz (default: "14074000")
- `-band` - Band designation (default: "20m")

### Examples

Send messages every 3 seconds:

```powershell
.\wsjtx_simulator.exe -interval 3
```

Simulate FT4 mode on 40m:

```powershell
.\wsjtx_simulator.exe -mode FT4 -frequency 7074000 -band 40m
```

Send to a specific IP and port:

```powershell
.\wsjtx_simulator.exe -addr 192.168.1.100 -port 2334
```

Use your own callsign:

```powershell
.\wsjtx_simulator.exe -callsign W1ABC
```

## Testing the UDP Logger Relay

1. **Start the UDP Logger Relay** in one terminal:
   ```powershell
   .\udp-logger-relay.exe --verbose
   ```

2. **Run the simulator** in another terminal:
   ```powershell
   .\tools\wsjtx_simulator.exe
   ```

3. **Watch the output** - You should see:
   - Simulator showing messages being sent
   - Relay showing messages being received and forwarded

## Message Format

The simulator cycles through typical WSJT-X message types:

1. CQ call: `CQ W1TEST FN42`
2. Reply: `W1TEST K2ABC FN42`
3. Signal report: `K2ABC W1TEST -15`
4. Confirmation: `W1TEST K2ABC RRR`
5. Final: `K2ABC W1TEST 73`

Each message includes:
- Timestamp (HHMMSS format)
- Frequency (in Hz)
- Mode (FT8, FT4, etc.)
- Message content

## Troubleshooting

### Simulator sends but relay doesn't receive

1. **Check firewall settings** - Windows Firewall may be blocking UDP
2. **Verify the port** - Make sure both are using the same port (default 2333)
3. **Check the address** - Relay should listen on `0.0.0.0` to receive from any interface
4. **Enable verbose mode** on the relay to see detailed logging

### "Address already in use" error

Another application (possibly WSJT-X) is already using port 2333. Either:
- Close the other application
- Use a different port: `-port 2334`

## Network Testing

To verify UDP connectivity, you can use PowerShell:

```powershell
# Check what's listening on UDP port 2333
Get-NetUDPEndpoint | Where-Object LocalPort -eq 2333
```

To test with netcat (if installed):

```powershell
# Listen for UDP on port 2333
nc -u -l 2333
```

Then run the simulator to send to that port.
