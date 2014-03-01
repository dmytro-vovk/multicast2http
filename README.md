Multicast2http
==============

Server to convert multicast/unicast streams into http based unicast.

Configuration
-------------

There are two configuration files in JSON format.

`sources.json` lists stream sources:
```js
{
	"/public-path": {
		"source": "udp://1.2.3.4:1234",
		"interface": "eth3",
		"set": 1
	},
	"/another-source": {
		"source": "http://1.2.3.4/source-path",
		"set": 2
	},
	...
}
```

`/public-path` is public path that will be server by the server.

`source` contains address of traffic source, either udp multicast or http unicast. For udp it must contain protocol, hostname, and port (`udp://1.2.3.4:56`). For http, it must be fully qualified URL (`http://source.domain/path`).

`interface` specifies network interface to listen for incoming udp traffic. Makes sense only for multicast sources.

`set` integer id of the access group source belongs to.

`networks.json` lists networks that can have access to streams:
```json
{
	"192.168.0.0/24": [
		1,
		2
	],
	...
}
```

`192.168.0.0/24` masked network ip.

`[ 1, 2 ]` array of integer ids of sets the network is allowed to access.

Command line arguments
----------------------

`sources` type string -- File with URL to source mappings.
  
`networks` type string -- File with networks to sets mappings.

`listen` type string -- Ip:port to listen for clients.

`fake-stream` type string -- Fake stream to return to non authorized clients.

`enable-web-controls` type boolean -- Whether to enable controls via special paths.

Example: `Streamer -listen=":7979" -sources="sources.json" -networks="networks.json" -fake-stream="fake.ts" -enable-web-controls=true >> /var/log/streamer.log 2>&1`

Web controls
------------

Special URLs can be enabled by `-enable-web-controls=true` command line argument. If enabled, two special URLs are available:

`http://your-server.url/server-status` -- Shows service uptime and number of connected clients.

`http://your-server.url/reload-config` -- Reloads config files.

Console controls
----------------

Server can handle HUP OS signal to reload configs. Configs are parsed and if they are valid, they are applied.
