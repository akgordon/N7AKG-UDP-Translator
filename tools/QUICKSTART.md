# Quick Start Guide - WSJT-X Simulator

## Basic Usage

1. **Quick test with defaults:**
   ```powershell
   .\tools\wsjtx_simulator.exe
   ```

2. **Use configuration file:**
   ```powershell
   .\tools\wsjtx_simulator.exe --config .\tools\simulator-config.yaml
   ```

3. **Override specific settings:**
   ```powershell
   .\tools\wsjtx_simulator.exe --config .\tools\simulator-config.yaml -interval 2
   ```

## Common Test Scenarios

### Test Different Bands

**20m FT8 (most common):**
```powershell
.\tools\wsjtx_simulator.exe -mode FT8 -frequency 14074000 -band 20m
```

**40m FT8:**
```powershell
.\tools\wsjtx_simulator.exe -mode FT8 -frequency 7074000 -band 40m
```

**17m FT8:**
```powershell
.\tools\wsjtx_simulator.exe -mode FT8 -frequency 18100000 -band 17m
```

### Test Different Modes

**FT4 on 20m:**
```powershell
.\tools\wsjtx_simulator.exe -mode FT4 -frequency 14080000 -band 20m
```

**MSK144 on 6m:**
```powershell
.\tools\wsjtx_simulator.exe -mode MSK144 -frequency 50276000 -band 6m
```

### Rapid Testing

**Send messages every 2 seconds:**
```powershell
.\tools\wsjtx_simulator.exe -interval 2
```

**Send messages every 10 seconds:**
```powershell
.\tools\wsjtx_simulator.exe -interval 10
```

### Network Testing

**Send to remote relay:**
```powershell
.\tools\wsjtx_simulator.exe -addr 192.168.1.100 -port 2333
```

**Send to custom port:**
```powershell
.\tools\wsjtx_simulator.exe -port 2334
```

## Customizing Messages

Edit `simulator-config.yaml` to change:

- **Your callsign:** `radio.callsign: "W1ABC"`
- **Your grid:** `radio.grid: "FN31pr"`
- **Remote station:** `messages.remote_callsign: "K2XYZ"`
- **Signal strength:** `messages.signal_report: "-10"`

## Typical Testing Workflow

1. **Start the UDP Logger Relay in one terminal:**
   ```powershell
   .\udp-logger-relay.exe --verbose
   ```

2. **Run simulator in another terminal:**
   ```powershell
   .\tools\wsjtx_simulator.exe
   ```

3. **Watch both terminals** to verify messages are sent and received

4. **Press Ctrl+C** in simulator to stop and see statistics

## Configuration File Template

Create custom config files for different test scenarios:

**`test-40m.yaml`** - Test 40m band
```yaml
target:
  address: "127.0.0.1"
  port: 2333
timing:
  interval: 3
radio:
  mode: "FT8"
  callsign: "W1TEST"
  grid: "FN42"
  frequency: "7074000"
  band: "40m"
messages:
  remote_callsign: "K2ABC"
  remote_grid: "FN31"
  signal_report: "-12"
```

**`test-ft4.yaml`** - Test FT4 mode
```yaml
target:
  address: "127.0.0.1"
  port: 2333
timing:
  interval: 4
radio:
  mode: "FT4"
  callsign: "W1TEST"
  grid: "FN42"
  frequency: "14080000"
  band: "20m"
messages:
  remote_callsign: "K2ABC"
  remote_grid: "FN31"
  signal_report: "-18"
```

Then use them:
```powershell
.\tools\wsjtx_simulator.exe --config test-40m.yaml
.\tools\wsjtx_simulator.exe --config test-ft4.yaml
```
