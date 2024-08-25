# DG-citrus

A server to act as a bridge between the [DG-LAB App](https://www.dungeon-lab.com/) and third party controller clients.

Compared to the [official implementation](https://github.com/DG-LAB-OPENSOURCE/DG-LAB-OPENSOURCE), this implementation has the following advantages:

- Adds an HTTP API protocol to provide a more flexible way to interact with the device, especially for restricted coding environments
  - eg. [VRChat Udon](https://creators.vrchat.com/worlds/udon/)
- Provides unlimited device bindings, allowing one-to-many and many-to-one control relationships
- Supports mixed controller client types for a single device, allowing complex control scenarios
  - eg. A device can be controlled by VRChat world buttons (using HTTP) and avatar interactions (using WebSocket through OSC) simultaneously
- Fully parses the protocol data into the level of a single pulse parameter, friendly for further development based on it
- High performance, low latency, and low resource consumption compared to other implementations

## Quick Start

```bash
git clone git@github.com:TundraWork/DG-citrus.git
cd DG-citrus
go mod tidy
./build.sh
cp config.example.yaml config.yaml
./output/bootstrap.sh
```

## Usage

### Configuration

`config.yaml`

- `HostName`: The public accessible hostname of the server
- `Port`: The port your server will listen on, for both HTTP and WebSocket connections
- `UseSecureWebsocket`: Whether to use secure WebSocket connections (wss://), otherwise use insecure connections (ws://)
- `AllowInsecureClientId`: Whether to allow clients to connect without a valid client ID, if this is set to `true`, the server will use only the IP address of a client to identify it. Useful for restricted coding environments.

### Websocket API

The websocket API is compatible with the [official implementation](https://github.com/DG-LAB-OPENSOURCE/DG-LAB-OPENSOURCE).

- DG-LAB App connections: `wss://<hostname>:<port>/app/<client ID>`
- Third party controller client connections: `wss://<hostname>:<port>/v1/ws`

### HTTP API

- Register a client: `GET /v1/register`
- Get DG-LAB App binding qrcode: `GET /v1/bind?clientId=<client ID>`
- Send a command to all bound devices: `GET /v1/command?clientId=<client ID>&message=<message field in official protocol>`
- Heartbeat: `GET /v1/heartbeat?clientId=<client ID>`

## License

DG-citrus is licensed under the [MIT License](LICENSE).
