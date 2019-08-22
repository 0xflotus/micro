// package network implements micro network node
package network

import (
	"strings"
	"time"

	"github.com/micro/cli"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/network"
	"github.com/micro/go-micro/network/resolver"
	"github.com/micro/go-micro/network/resolver/dns"
	"github.com/micro/go-micro/network/resolver/http"
	"github.com/micro/go-micro/network/resolver/registry"
	"github.com/micro/go-micro/router"
	"github.com/micro/go-micro/tunnel"
	"github.com/micro/go-micro/util/log"
)

var (
	// Name of the network service
	Name = "go.micro.network"
	// Address is the tunnel address
	Address = ":8084"
	// Tunnel is the name of the tunnel
	Tunnel = "tun:0"
	// Resolver is network resolver
	Resolver = "dns"
)

// run runs the micro server
func run(ctx *cli.Context, srvOpts ...micro.Option) {
	// Init plugins
	for _, p := range Plugins() {
		p.Init(ctx)
	}

	if len(ctx.GlobalString("server_name")) > 0 {
		Name = ctx.GlobalString("server_name")
	}
	if len(ctx.String("address")) > 0 {
		Address = ctx.String("address")
	}
	if len(ctx.String("tunnel_id")) > 0 {
		Tunnel = ctx.String("tunnel_id")
		// We need host:port for the Endpoint value in the proxy
		parts := strings.Split(Tunnel, ":")
		if len(parts) == 1 {
			Tunnel = Tunnel + ":0"
		}
	}
	var nodes []string
	if len(ctx.String("server")) > 0 {
		nodes = strings.Split(ctx.String("server"), ",")
	}

	if len(ctx.String("resolver")) > 0 {
		Resolver = ctx.String("resolver")
	}
	var res resolver.Resolver
	switch Resolver {
	case "dns":
		res = &dns.Resolver{}
	case "http":
		res = &http.Resolver{}
	case "registry":
		res = &registry.Resolver{}
	}

	// create a tunnel
	tun := tunnel.NewTunnel(
		tunnel.Address(Address),
		tunnel.Nodes(nodes...),
	)

	// local tunnel router
	rtr := router.NewRouter(
		router.Network(Name),
	)

	// creaate new network
	net := network.NewNetwork(
		network.Name(Name),
		network.Address(Address),
		network.Tunnel(tun),
		network.Router(rtr),
		network.Resolver(res),
	)

	// Initialise service
	service := micro.NewService(
		micro.Name(Name),
		micro.RegisterTTL(time.Duration(ctx.GlobalInt("register_ttl"))*time.Second),
		micro.RegisterInterval(time.Duration(ctx.GlobalInt("register_interval"))*time.Second),
		micro.Server(net.Server()),
	)

	// initialize router
	rtr.Init(
		router.Id(service.Server().Options().Id),
		router.Registry(service.Client().Options().Registry),
	)

	if err := service.Run(); err != nil {
		log.Log("Network %s failed: %v", Name, err)
	}
}

func Commands(options ...micro.Option) []cli.Command {
	command := cli.Command{
		Name:  "network",
		Usage: "Run the micro network node",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "address",
				Usage:  "Set the micro network address :8084",
				EnvVar: "MICRO_NETWORK_ADDRESS",
			},
			cli.StringFlag{
				Name:   "tunnel_id",
				Usage:  "Id of the tunnel used as the internal dial/listen address.",
				EnvVar: "MICRO_TUNNEL_ID",
			},
			cli.StringFlag{
				Name:   "server",
				Usage:  "Set the micro network server address. This can be a comma separated list.",
				EnvVar: "MICRO_NETWORK_SERVER",
			},
			cli.StringFlag{
				Name:   "resolver",
				Usage:  "Set the micro network resolver. This can be a comma separated list.",
				EnvVar: "MICRO_NETWORK_RESOLVER",
			},
		},
		Action: func(ctx *cli.Context) {
			run(ctx, options...)
		},
	}

	for _, p := range Plugins() {
		if cmds := p.Commands(); len(cmds) > 0 {
			command.Subcommands = append(command.Subcommands, cmds...)
		}

		if flags := p.Flags(); len(flags) > 0 {
			command.Flags = append(command.Flags, flags...)
		}
	}

	return []cli.Command{command}
}
