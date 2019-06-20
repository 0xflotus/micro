package server

import (
	"fmt"
	"os"
	"time"

	"github.com/micro/cli"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry/gossip"
	"github.com/micro/go-micro/router"
	"github.com/micro/go-micro/server"
	"github.com/micro/go-micro/transport"
	"github.com/micro/go-micro/transport/grpc"
	"github.com/micro/go-micro/util/log"
)

var (
	// Name of the server microservice
	Name = "go.micro.server"
	// Address is the router microservice bind address
	Address = ":8083"
	// Router is the router bind address a.k.a. gossip bind address
	Router = ":9093"
	// Network is the micro network bind address
	Network = ":9094"
)

// srv is micro server
type srv struct {
	// router is micro router
	router router.Router
	// network is micro network server
	network server.Server
}

// newServer creates new micro server and returns it
func newServer(s micro.Service, r router.Router) *srv {
	// NOTE: this will end up being QUIC transport once it gets stable
	t := grpc.NewTransport(transport.Addrs(Network))
	n := server.NewServer(server.Transport(t))

	return &srv{
		router:  r,
		network: n,
	}
}

// start starts the micro server.
func (s *srv) start() error {
	log.Log("[server] starting micro server")

	// start the router
	if err := s.router.Start(); err != nil {
		return fmt.Errorf("failed to start router: %v", err)
	}

	log.Logf("[server] router successfully started: \n%s", s.router)
	log.Logf("[server] initial routing table: \n%s", s.router.Table())

	return nil
}

// stop stops the micro server
func (s *srv) stop() error {
	log.Log("[server] attempting to stop server")

	// stop the router
	if err := s.router.Stop(); err != nil {
		return fmt.Errorf("failed to stop router: %v", err)
	}

	log.Log("[server] router successfully stopped")

	return nil
}

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
	if len(ctx.String("router")) > 0 {
		Router = ctx.String("router")
	}
	if len(ctx.String("network")) > 0 {
		Network = ctx.String("network")
	}

	// Initialise service
	service := micro.NewService(
		micro.Name(Name),
		micro.Address(Address),
		micro.RegisterTTL(time.Duration(ctx.GlobalInt("register_ttl"))*time.Second),
		micro.RegisterInterval(time.Duration(ctx.GlobalInt("register_interval"))*time.Second),
	)

	// create new router
	r := router.NewRouter(
		router.ID(service.Server().Options().Id),
		router.Address(Address),
		router.GossipAddress(Router),
		router.NetworkAddress(Network),
		router.LocalRegistry(service.Client().Options().Registry),
		router.NetworkRegistry(gossip.NewRegistry(gossip.Address(Router), gossip.Advertise(Router))),
	)

	// create new server and start it
	s := newServer(service, r)

	// start the server
	if err := s.start(); err != nil {
		log.Logf("[server] error starting server: %v", err)
		os.Exit(1)
	}

	// Run the server as a micro service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}

	// stop the server
	if err := s.stop(); err != nil {
		log.Logf("[server] error stopping server: %v", err)
		os.Exit(1)
	}

	log.Logf("[server] successfully stopped")
}

func Commands(options ...micro.Option) []cli.Command {
	command := cli.Command{
		Name:  "server",
		Usage: "Run the micro network server",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "address",
				Usage:  "Set the micro server address :8083",
				EnvVar: "MICRO_SERVER_ADDRESS",
			},
			cli.StringFlag{
				Name:   "router",
				Usage:  "Set the micro router address :9093",
				EnvVar: "MICRO_ROUTER_ADDRESS",
			},
			cli.StringFlag{
				Name:   "network",
				Usage:  "Set the micro network address :9094",
				EnvVar: "MICRO_NETWORK_ADDRESS",
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
