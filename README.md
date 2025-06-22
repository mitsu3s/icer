# ICER

ICER is a tool designed for sending ICMP error messages to a specified target.
It can be used for testing and debugging network configurations, simulating network
errors, or analyzing how systems handle ICMP error responses.

> [!CAUTION]
> Use with caution, as improper use may violate network policies or disrupt services.

## Requirement

| Language/FrameWork | Version |
| :----------------- | ------: |
| Go                 |  1.23.3 |

## Installation

Install ICER, follow the steps below:

```sh
# Install this repository
$ git clone git@github.com:mitus3s/icer.git

# Move to the command directory
$ cd cmd

# Build the executable
$ go build -o icer

```

## Usage

Run the icer with the following syntax:

```sh
$ ./icer [flags/commands]
```

### Flags

- `-t`, `--type`
  Specify the ICMP Type (e.g., 3 for Destination Unreachable, 5 for Redirect Message, 11 for Time Exceeded).

- `-c`, `--code`
  Specify the ICMP Code corresponding to the Type.

### Example

Send ICMP Redirect Message (Redirect Datagram for the Network):

```sh
# ICMP Redirect (Type=5, Code=0)
$ ./icer -t 5 -c 0
```

### Configuration

The IP address settings can be modified in the configuration file located at [`data/config.yaml`](https://github.com/mitsu3s/icer/blob/main/data/config.yaml).

```yaml
# data/config.yaml
real_ip: # IP used for sending/receiving
  src_ip: "192.168.10.10"
  dst_ip: "192.168.20.20"

error_ip: # IP causing the error
  src_ip: "10.10.10.10"
  dst_ip: "20.20.20.20"
```

### Other Command

- `help`
  Show usage information
- `version`
  Show icer version

> [!IMPORTANT]
> Since the operation has been checked only on Linux, operation on other operating systems is not guaranteed.

## Contributing

Thank you for considering contributing to our source code! We deeply appreciate any contributions, no matter how small, and we are truly grateful for your help.

If you encounter any issues, please feel free to [icer issues](https://github.com/mitsu3s/icer/issues) in this repository. We are happy to assist you!

## LICENSE

[MIT License](./LICENSE)
