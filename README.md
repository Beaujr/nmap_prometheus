# nmap_prometheus

A client/server application set for exporting nmap results to prometheus metrics.

### Metrics
```
# Example
home_detector_device{ip="192.168.1.5",mac="31:AB:CF:34:B1:2L",name="Beaus Phone"} 1
home_detector_device_lastseen{ip="192.168.1.5",mac="31:AB:CF:34:B1:2L",name="Beaus Phone"} 1.592911572e+09
```
One of each of the above metrics will be posted for each device found on the network.

```home_detector_device```  is 1 if ``away=false`` and 0 if ``away=true``
```home_detector_device_lastseen``` is a unix timestamp of the last time it was reported to the server.


## Design
## GOOS & GOARCH
Docker builds are multiplatform for armv7 and amd64

Tested and working on RaspberryPi 3 and Intel Nuc

### Client
The client is a gRPC/nmap client application which communicates with the server.

Made possible due to [Ullaakut/nmap](https://github.com/Ullaakut/nmap)
#### Flags
```bash
    -server=<nmap_prometheus_server>:<port>
    -subnet=<your subnet range>
```

### Server
The Server is a GRPC server which accepts and logs the payloads as prometheus metrics.
```bash
 --timeout <number of seconds since last reported used determine device away>
```


Currently all detected devices will be saved to a config/devices.yaml file.

The initial plan was for the server to talk to:
 - [Assistant Relay](https://greghesp.github.io/assistant-relay/docs/introduction/)
 - [go-fcm-server](https://github.com/Beaujr/go-fcm-server)

This will be supported in limited capacity with the following flags.
```bash
  --assistant=https://<assistant_relay_url>
  --assistantUser=<your_assistant_relay_user>
  --fcm=https://<go_fcm_server>/fcm/send/<topic>
```
### Devices.yaml
#### Example
```yaml
- id:
    ip: 192.168.1.110
    mac: 31:AB:CF:34:B1:2L
  lastseen: 1592902392
  away: false
  name: Beau
  person: true
  command: ""
  smart: false
```

#### Fields
| Name | Type | Description | Example |
|---|---|---|---|
|Id   | Object  | Mac & Ip Address| see above |
| lastSeen  | int64  | Unix Timestamp of the last time the device was reported | 1592902392 |
|  away | bool  | Device hasn't been reported for > the servers --timeout flag| true |
| name  | string  | Device Name, defaults to devices Mac / Ip | TV |
| person  |  bool | if set to true will update | false |
| command  | string  | Command to get state from Assistant relay | ```Is The TV On?``` |
| smart  |  bool | If true will used ```command``` string to get state from Assistant Relay | ```true``` |

#### Debug
```bash
   --debug=true
```
This disables the server trying to talk to other services.

The initial plan was for the server to talk to:
 - [Assistant Relay](https://greghesp.github.io/assistant-relay/docs/introduction/)
 - [go-fcm-server](https://github.com/Beaujr/go-fcm-server)

This will be supported in limited capacity with the following flags.
```bash
  --assistant=https://<assistant_relay_url>
  --assistantUser=<your_assistant_relay_user>
  --fcm=https://<go_fcm_server>/fcm/send/<topic>
```

## Installation

- Clone repository
- Change ```-subnet=<your_desired_subnet_range>```
```yaml
   version: '2'
   services:
     client:
       image: beaujr/nmap_prometheus:client-latest
       command: '-server=localhost:50051 -subnet=192.168.1.100-255'
       network_mode: host
     server:
       image: beaujr/nmap_prometheus:server-latest
       ports:
         - 50051:50051
         - 2112:2112
       volumes:
         - ./config:/config
       command: '-debug=true'
```

- run `docker-compose up`
- open http://localhost:2112/metrics and see the prometheus metrics
