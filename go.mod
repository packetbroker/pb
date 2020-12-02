module go.packetbroker.org/pb

go 1.15

replace go.packetbroker.org/api/v3 => ../../go.packetbroker.org/api/v3

replace go.packetbroker.org/api/routing => ../../go.packetbroker.org/api/routing

replace go.packetbroker.org/api/iam => ../../go.packetbroker.org/api/iam

require (
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2
	github.com/spf13/pflag v1.0.5
	go.packetbroker.org/api/iam v0.0.0-00010101000000-000000000000
	go.packetbroker.org/api/routing v0.0.0-00010101000000-000000000000
	go.packetbroker.org/api/v3 v3.1.1
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb // indirect
	golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58
	golang.org/x/sys v0.0.0-20201201145000-ef89a241ccb3 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20201202151023-55d61f90c1ce // indirect
	google.golang.org/grpc v1.33.2
	google.golang.org/protobuf v1.25.0
	htdvisser.dev/exp/clicontext v1.1.0
	mvdan.cc/gofumpt v0.0.0-20200709182408-4fd085cb6d5f
)
