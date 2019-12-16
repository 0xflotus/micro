package store

import (
	"strings"
	"time"

	"github.com/micro/cli"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/config/options"
	"github.com/micro/go-micro/store"
	pb "github.com/micro/go-micro/store/service/proto"
	"github.com/micro/go-micro/util/log"
	"github.com/micro/micro/store/handler"

	"github.com/micro/go-micro/store/memory"
	"github.com/micro/go-micro/store/postgresql"
)

var (
	// Name of the tunnel service
	Name = "go.micro.store"
	// Address is the tunnel address
	Address = ":8002"
	// Backend is the implementation of the store
	Backend = "memory"
	// Nodes is passed to the underlying backend
	Nodes = []string{"localhost"}
	// Namespace is passed to the underlying backend if set.
	Namespace = ""
)

// run runs the micro server
func run(ctx *cli.Context, srvOpts ...micro.Option) {
	log.Name("store")

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
	if len(ctx.String("backend")) > 0 {
		Backend = ctx.String("backend")
	}
	if len(ctx.String("nodes")) > 0 {
		Nodes = strings.Split(ctx.String("nodes"), ",")
	}
	if len(ctx.String("namespace")) > 0 {
		Namespace = ctx.String("namespace")
	}

	// Initialise service
	service := micro.NewService(
		micro.Name(Name),
		micro.RegisterTTL(time.Duration(ctx.GlobalInt("register_ttl"))*time.Second),
		micro.RegisterInterval(time.Duration(ctx.GlobalInt("register_interval"))*time.Second),
	)

	newStore := &handler.Store{}
	opts := []options.Option{store.Nodes(Nodes...)}
	if len(Namespace) > 0 {
		opts = append(opts, store.Namespace(Namespace))
	}

	switch Backend {
	case "memory":
		newStore.Store = memory.NewStore(opts...)
	case "postgresql":
		opts = append(opts, options.WithValue("store.sql.driver", "postgres"))
		if s, err := postgresql.New(opts...); err != nil {
			log.Fatal(err)
		} else {
			newStore.Store = s
		}
	default:
		log.Fatalf("%s is not an implemented store")
	}

	pb.RegisterStoreHandler(service.Server(), newStore)

	// start the service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

// Commands is the cli interface for the store service
func Commands(options ...micro.Option) []cli.Command {
	command := cli.Command{
		Name:  "store",
		Usage: "Run the micro store service",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "address",
				Usage:  "Set the micro tunnel address :8002",
				EnvVar: "MICRO_SERVER_ADDRESS",
			},
			cli.StringFlag{
				Name:   "backend",
				Usage:  "Set the backend for the micro store",
				EnvVar: "MICRO_STORE_BACKEND",
				Value:  "memory",
			},
			cli.StringFlag{
				Name:   "nodes",
				Usage:  "Comma separated list of Nodes to pass to the store backend",
				EnvVar: "MICRO_STORE_NODES",
			},
			cli.StringFlag{
				Name:   "namespace",
				Usage:  "Namespace to pass to the store backend",
				EnvVar: "MICRO_STORE_NAMESPACE",
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
