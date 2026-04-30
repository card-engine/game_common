package sever

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/card-engine/game_common/gamehub/common"
	"github.com/card-engine/game_common/gamehub/inout"
	"github.com/card-engine/game_common/gamehub/jdb"
	"github.com/card-engine/game_common/gamehub/jili"
	"github.com/card-engine/game_common/gamehub/spribe"
	"github.com/card-engine/game_common/gamehub/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"

	google_grpc "google.golang.org/grpc"
)

type GameApiServer struct {
	app *fiber.App
	log *log.Helper

	serverAddr string //服务器绑定的地址
	router     types.Router

	endpoint *url.URL
	lis      net.Listener
}

// 没有大厅类的游戏
func InitGameApiServer(gameBrand types.GameBrand, // 游戏品牌
	serverName string, // 服务器名称
	gameName string, // 游戏名称
	serverAddr string, // 服务器绑定的地址
	tableMatcherType types.TableMatcherType, // 配桌算法
	roomCreator types.RoomCreator, // 房间创建器
	rdb *redis.Client, // redis 客户端
	apiGrpcConn *google_grpc.ClientConn,
	rtpGrpcConn *google_grpc.ClientConn, // rtp 客户端
	logger log.Logger) *GameApiServer {
	// 没有大厅的使用自建无大厅创建器
	return InitGameApiServerWithLobby(gameBrand, serverName, gameName, serverAddr, tableMatcherType, roomCreator, common.NewNoLobbyCreator(tableMatcherType), rdb, apiGrpcConn, rtpGrpcConn, logger)
}

// 有大厅的游戏的服务器
func InitGameApiServerWithLobby(
	gameBrand types.GameBrand, // 游戏品牌
	serverName string, // 服务器名称
	gameName string, // 游戏名称
	serverAddr string, // 服务器绑定的地址
	tableMatcherType types.TableMatcherType, // 配桌算法
	roomCreator types.RoomCreator, // 房间创建器
	lobbyCreator types.LobbyCreator, // 大厅创建器
	rdb *redis.Client, // redis 客户端
	apiGrpcConn *google_grpc.ClientConn,
	rtpGrpcConn *google_grpc.ClientConn, // rtp 客户端
	logger log.Logger) *GameApiServer {
	app := fiber.New()

	s := &GameApiServer{
		app: app,
		log: log.NewHelper(logger),

		serverAddr: serverAddr,
	}

	roomManager := common.NewRoomManager(gameBrand, roomCreator, tableMatcherType, logger)

	var lobby types.LobbyImp = nil
	if lobbyCreator != nil {
		lobby = lobbyCreator.CreateLobby(roomManager)
	}

	switch gameBrand {
	case types.GameBrand_Inout:
		s.router = inout.NewInoutRouter(gameName, app, rdb, apiGrpcConn, rtpGrpcConn, roomManager, lobby, logger)
	case types.GameBrand_Spribe:
		s.router = spribe.NewSpribeRouter(gameName, app, rdb, apiGrpcConn, rtpGrpcConn, roomManager, lobby, logger)
	case types.GameBrand_Jdb:
		s.router = jdb.NewJdbRouter(gameName, app, rdb, apiGrpcConn, rtpGrpcConn, roomManager, lobby, logger)
	case types.GameBrand_Jili:
		s.router = jili.NewJiliRouter(gameName, app, rdb, apiGrpcConn, rtpGrpcConn, roomManager, lobby, logger)
	default:
		s.log.Fatalf("gameBrand %v not support", gameBrand)
	}

	s.route()
	return s
}

func (s *GameApiServer) route() {
	if s.router == nil {
		s.log.Fatalf("router is nil")
	}

	s.router.Route()
}

func (s *GameApiServer) Start(ctx context.Context) error {
	if err := s.ensureListener(); err != nil {
		return err
	}

	go func() {
		if err := s.app.Listener(s.lis); err != nil {
			log.Fatalf("Listen failed: %v", err)
		}
	}()

	return nil
}

func normalizeListenAddr(addr string) string {
	a := strings.TrimSpace(addr)
	if a == "" {
		return ":0"
	}
	// If user only provides host without port, auto-pick a random port.
	// Examples: "127.0.0.1" -> "127.0.0.1:0", "0.0.0.0" -> "0.0.0.0:0"
	if !strings.Contains(a, ":") {
		return net.JoinHostPort(a, "0")
	}
	return a
}

func (s *GameApiServer) Stop(ctx context.Context) error {
	if s.app != nil {
		s.app.Shutdown()
		s.app = nil
	}
	s.lis = nil

	return nil
}

// 实现 Endpointer 接口
func (s *GameApiServer) Endpoint() (*url.URL, error) {
	if s.endpoint != nil {
		return s.endpoint, nil
	}

	// Ensure we have a real port (handles ":0" and empty addr).
	if err := s.ensureListener(); err != nil {
		return nil, err
	}

	host, port, err := s.boundHostPort()
	if err != nil {
		return nil, err
	}

	if host == "" || host == "0.0.0.0" || host == "::" {
		if ip := realLocalIP(); ip != "" {
			host = ip
		}
	}

	s.endpoint = &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, port),
	}
	return s.endpoint, nil
}

func (s *GameApiServer) ensureListener() error {
	if s.lis != nil {
		return nil
	}
	addr := normalizeListenAddr(s.serverAddr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.lis = ln
	// Replace with the actual bound address (handles ":0" etc.)
	s.serverAddr = ln.Addr().String()
	s.endpoint = nil
	return nil
}

func (s *GameApiServer) boundHostPort() (string, string, error) {
	// Use the final bound address (Start() overwrites it with ln.Addr().String()).
	host, port, err := splitHostPort(s.serverAddr)
	if err != nil {
		return "", "", fmt.Errorf("invalid serverAddr %q: %w", s.serverAddr, err)
	}
	if port == "" {
		return "", "", fmt.Errorf("missing port in serverAddr %q", s.serverAddr)
	}
	return host, port, nil
}

func splitHostPort(addr string) (host string, port string, err error) {
	// Handles:
	// - ":8080"
	// - "0.0.0.0:8080"
	// - "127.0.0.1:8080"
	// - "[::]:8080"
	// - "localhost:8080"
	if addr == "" {
		return "", "", fmt.Errorf("empty address")
	}
	return net.SplitHostPort(addr)
}

func realLocalIP() string {
	// Align with Kratos gRPC endpoint selection:
	// pick a global-unicast IP from the smallest interface index,
	// prefer IPv4 when available.
	return pickKratosLikeIP()
}

func isValidIP(addr string) bool {
	ip := net.ParseIP(addr)
	return ip != nil && ip.IsGlobalUnicast() && !ip.IsInterfaceLocalMulticast()
}

func pickKratosLikeIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	var (
		minIndex = 0
		ips      = make([]net.IP, 0, 1)
	)
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Index >= minIndex && len(ips) != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}
			if ip == nil {
				continue
			}
			if isValidIP(ip.String()) {
				minIndex = iface.Index
				ips = append(ips, ip)
				// Prefer IPv4 when possible.
				if ip.To4() != nil {
					break
				}
				continue
			}
		}
	}
	if len(ips) != 0 {
		return ips[len(ips)-1].String()
	}
	return ""
}
