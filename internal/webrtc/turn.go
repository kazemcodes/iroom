package webrtc

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/pion/turn/v5"
)

type TURNServer struct {
	server   *turn.Server
	publicIP string
	port     int
	secret   string
}

func NewTURNServer(publicIP string, port int, secret string) *TURNServer {
	return &TURNServer{
		publicIP: publicIP,
		port:     port,
		secret:   secret,
	}
}

func (ts *TURNServer) Start() error {
	udpListener, err := net.ListenPacket("udp4", fmt.Sprintf("0.0.0.0:%d", ts.port))
	if err != nil {
		return fmt.Errorf("failed to create TURN listener on port %d: %w", ts.port, err)
	}

	s, err := turn.NewServer(turn.ServerConfig{
		Realm: "iroom",
		AuthHandler: func(ra *turn.RequestAttributes) (string, []byte, bool) {
			key := turn.GenerateAuthKey(ra.Username, "iroom", ts.secret)
			return ra.Username, key, true
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(ts.publicIP),
					Address:      "0.0.0.0",
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to start TURN server: %w", err)
	}

	ts.server = s
	slog.Info("TURN server started", "addr", fmt.Sprintf("%s:%d", ts.publicIP, ts.port))
	return nil
}

func (ts *TURNServer) Close() {
	if ts.server != nil {
		ts.server.Close()
	}
}

func (ts *TURNServer) GetURL() string {
	return fmt.Sprintf("turn:%s:%d", ts.publicIP, ts.port)
}
