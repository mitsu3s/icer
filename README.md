# ICER

ICER is a tool designed for sending ICMP error messages to a specified target.
It can be used for testing and debugging network configurations, simulating network
errors, or analyzing how systems handle ICMP error responses. The tool provides
flexibility and precision, allowing users to specify various parameters for ICMP
message construction and delivery.

> [!CAUTION]
> Use with caution, as improper use may violate network policies or disrupt services.


## Requirement

| Language/FrameWork | Version |
| :----------------- | ------: |
| Go                 |  1.22.9 |

## Installation
To install ICER, follow the steps below:
```sh
# Install this repository
$ git clone git@github.com:mitus3s/icer

# Move to the command directory
$ cd icer/cmd

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

### Other Command
- `help`
Show usage information
- `version`
Show icer version


## LICENSE

[MIT License](./LICENSE)
