package az

import "github.com/aykhans/bsky-feedgen/pkg/generator"

var users = generator.Users{
	// Invalid
	"did:plc:5zww7zorx2ajw7hqrhuix3ba": false,
	"did:plc:c4vhz47h566t2ntgd7gtawen": false,
	"did:plc:lc7j7xdq67gn7vc6vzmydfqk": false,
	"did:plc:msian4dqa2rqalf3biilnf3m": false,
	"did:plc:gtosalycg7snvodjhsze35jm": false,

	// Valid
	"did:plc:jbt4qi6psd7rutwzedtecsq7": true,
	"did:plc:yzgdpxsklrmfgqmjghdvw3ti": true,
	"did:plc:g7ebgiai577ln3avsi2pt3sn": true,
	"did:plc:phtq2rhgbwipyx5ie3apw44j": true,
	"did:plc:jfdvklrs5n5qv7f25v6swc5h": true,
	"did:plc:u5ez5w6qslh6advti4wyddba": true,
	"did:plc:cs2cbzojm6hmx5lfxiuft3mq": true,
}
