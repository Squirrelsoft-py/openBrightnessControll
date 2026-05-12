# Brightness Control

Automatically adjusts external monitor brightness based on local sunrise and sunset times. Brightness ramps smoothly through a configurable transition window centred on each event. Uses `ddcutil` to send DDC/CI commands over I2C.

## How it works

- Calculates today's sunrise/sunset for your coordinates each morning
- Every 60 seconds (configurable), computes the target brightness
- During the transition window the value is linearly interpolated between day and night brightness
- Applies changes only when the value actually changes (no unnecessary DDC writes)

## Dependencies

### Debian / Ubuntu

```bash
# Install Go (1.22+)
sudo apt install golang-go

# Install ddcutil and the I2C kernel module
sudo apt install ddcutil i2c-tools

# Load the I2C kernel module
sudo modprobe i2c-dev

# Make i2c-dev load on boot
echo "i2c-dev" | sudo tee /etc/modules-load.d/i2c-dev.conf

# Allow your user to access I2C devices without sudo
sudo usermod -aG i2c "$USER"
# Log out and back in for the group change to take effect
```

Verify `ddcutil` works before proceeding:

```bash
ddcutil detect
ddcutil getvcp 10   # should print current brightness
```

## Build

```bash
git clone <repo-url>
cd brightness-control
go build -o brightness .
```

## Configuration

Copy `config.toml` and edit to match your location and preferences:

```toml
[location]
latitude         = 52.52
longitude        = 13.41
timezone         = "Europe/Berlin"
name             = "Berlin"

[brightness]
day              = 95   # percent, daytime target
night            = 10   # percent, nighttime target

[transition]
# Total window centred on sunrise/sunset.
# 60 min → ramp from 30 min before to 30 min after each event.
duration_minutes = 60

[schedule]
check_interval_seconds = 60
```

## Usage

```bash
# Run once (set brightness now and exit)
./brightness --once

# Dry run (log brightness without calling ddcutil)
./brightness --dry-run

# Specify a custom config path
./brightness --config /path/to/config.toml
```

## systemd user service

1. Install the binary and config to the expected locations:

```bash
mkdir -p ~/.local/bin ~/.config/brightness-control
cp brightness ~/.local/bin/
cp config.toml ~/.config/brightness-control/
# edit ~/.config/brightness-control/config.toml with your coordinates
```

2. Install and enable the service:

```bash
mkdir -p ~/.config/systemd/user
cp brightness.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now brightness.service
```

3. Check status or restart after a config change:

```bash
systemctl --user status brightness.service
# or use the helper script:
bash restart.sh
```

Logs are written to the systemd journal:

```bash
journalctl --user -u brightness.service -f
```
