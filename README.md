# icer

`icer` is a command-line tool designed for crafting and sending precise ICMP error messages (such as Destination Unreachable, Redirect, or Time Exceeded) to a specified target. 

It is useful for testing and debugging network configurations, simulating network errors, or validating how routers and application firewalls handle ICMP error responses.

> [!CAUTION]
> Use with caution. Improper use may violate network policies, disrupt services, or trigger intrusion detection systems. You generally need `root` or `CAP_NET_RAW` capabilities to send raw packets.

> [!NOTE]
> ICMP error messages (Type 3, 5, 11) are **one-way notifications** and do not have acknowledgment responses like TCP. This tool measures whether packets were successfully submitted to the OS network stack, not whether they reached the destination.

## Requirements

- Go 1.25+

## Installation

You can install `icer` easily via `go install`:

```sh
go install github.com/mitsu3s/icer/cmd/icer@latest
```

Once installed, verify it's working:

```sh
icer version
```

## Usage

`icer` features a modern subcommand architecture:

```sh
icer [command] [flags]
```

### Supported Commands

- `send` - Send ICMP error packets
- `list-types` - Show supported ICMP types and codes
- `init` - Generate a sample configuration file
- `version` - Show the current version

### Examples

**1. CLI-only Sending (No config file required)**

You can send a one-off ICMP error message by specifying the outer IP headers (for the ICMP packet itself) and the inner IP headers (the original packet being quoted in the ICMP error body).

```sh
# Send ICMP Destination Unreachable (Type 3, Code 0)
sudo icer send -t 3 -c 0 \
  --src 192.168.1.1 --dst 192.168.1.10 \
  --error-src 10.0.0.1 --error-dst 10.0.0.10
```

**2. Send Multiple Packets**

```sh
# Send 5 packets, 1 second apart
sudo icer send -t 3 -c 1 --count 5 --interval 1s \
  --src 192.168.1.1 --dst 192.168.1.10 \
  --error-src 10.0.0.1 --error-dst 10.0.0.10
```

**3. Dry Run Mode**

Simulate sending without actually opening a raw socket (useful if running without root privileges):

```sh
icer send -t 5 -c 0 --dry-run -v \
  --src 192.168.1.1 --dst 192.168.1.10 \
  --error-src 10.0.0.1 --error-dst 10.0.0.10
```

**4. Check Supported ICMP Types**

```sh
icer list-types

# Output:
# Type 3: Destination Unreachable
#   Indicates that a destination is unreachable
#   Code 0: Network Unreachable
#   Code 1: Host Unreachable
# ...
```

## Configuration Profiles

If you frequently test against the same targets, you can use configuration profiles.

1. Generate a sample config:
   ```sh
   icer init
   ```
   *This creates `~/.config/icer/config.yaml`.*

2. Use a profile to send packets:
   ```sh
   sudo icer send -t 3 -c 0 -p default
   # Or for a specific profile in your config:
   sudo icer send -t 3 -c 0 -p web-servers
   ```

A profile config file allows defining complex targets, for example:

```yaml
profiles:
  - name: "multi-target"
    description: "Hit multiple servers at once"
    targets:
      - real_ip: { src_ip: "192.168.1.1", dst_ip: "192.168.1.10" }
        error_ip: { src_ip: "10.0.0.1", dst_ip: "10.0.0.10" }
      - real_ip: { src_ip: "192.168.1.1", dst_ip: "192.168.1.11" }
        error_ip: { src_ip: "10.0.0.1", dst_ip: "10.0.0.11" }
```

## Troubleshooting

- **"socket: operation not permitted"**: You must run `icer` as root (`sudo icer ...`) or grant raw socket capabilities (`setcap cap_net_raw+ep $(which icer)`).
- **"unknown or unsupported type"**: Run `icer list-types` to see which type and code combinations are allowed.

## Contributing

Thank you for considering contributing to our source code! We deeply appreciate any contributions, no matter how small, and we are truly grateful for your help.

If you encounter any issues, please feel free to [open an issue](https://github.com/mitsu3s/icer/issues) in this repository.

## LICENSE

[MIT License](./LICENSE)
