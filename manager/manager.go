package manager

import (
    "fmt"
    "sync"
    "time"
    "net"
    "net/http"
    "path/filepath"
    "github.com/labstack/echo"
    "github.com/sirupsen/logrus"
    "github.com/seveirbian/gear/pkg"
    "github.com/seveirbian/gear/types"
)

var (
    logger = logrus.WithField("gear", "manager")
)

var (
    GearPath             = "/var/lib/gear/"
    GearGzipPath         = filepath.Join(GearPath, "gzip")
    GearStoragePath      = filepath.Join(GearPath, "storage")
)

var (
    port = "2019"
    mgr Manager = Manager{}
)

type Manager struct {
	Self types.Node

    Echo *echo.Echo

    NodesMu sync.RWMutex
    Nodes map[uint64]types.Node

    MonitorIp string
    MonitorPort string
}

func Init() (*Manager, error) {
    // 1. create echo instance
    e := echo.New()

    // 2. add routes
    e.GET("/nodes", handleNodes)
    e.POST("/join/:IP/:Port", handleJoin)
    e.POST("/pull/:CID", handlePull)
    e.POST("/query/:CID", handleQuery)
    e.POST("/push/:CID", handlePush)

    e.POST("/prefetch", handlePreFetch)

    // e.POST("/report/:IMAGE", handleReport)

    // 3. get self's IP
    ip := pkg.GetSelfIp()

    // 4. create self ID
    id := pkg.CreateIdFromIP(ip)

    // 5. fill the manager's fileds
    mgr.Self = types.Node{
            ID:   id,
            IP:   ip,
            Port: port,
        }
    mgr.Echo = e
    mgr.NodesMu = sync.RWMutex{}
    mgr.Nodes = map[uint64]types.Node{}

    // 6. Monitor's ip and port
    mgr.MonitorIp = pkg.GetSelfIp()
    mgr.MonitorPort = "2021"

    return &mgr, nil
} 

func (m *Manager) Start() {
    fmt.Printf("Manager IP: %s\n", m.Self.IP)

    // 需要一个协程定期查询所有节点是否在线
    go updateNodes(m)

    m.start()
}

func (m *Manager) start() {
    m.Echo.Logger.Fatal(m.Echo.Start(":" + m.Self.Port))
}

func updateNodes(m *Manager) {
    for {
        nodes, err := m.askNodes()
        if err != nil {
            logger.Fatal("Fail to ask cluster nodes...")
        }

        m.NodesMu.Lock()
        for _, node := range(nodes) {
            delete(m.Nodes, node.ID)    
        }
        m.NodesMu.Unlock()

        time.Sleep(60*time.Second)
    }
}

func (m *Manager) askNodes() ([]types.Node, error) {
    var quitedNodes = []types.Node{}

    for _, node := range(m.Nodes) {
        c := http.Client{
            Transport: &http.Transport{
                Dial: func(netw, addr string) (net.Conn, error) {
                    // deadline := time.Now().Add(25 * time.Second)
                    c, err := net.DialTimeout(netw, addr, time.Second*5)
                    if err != nil {
                        return nil, err
                    }
                    // c.SetDeadline(deadline)
                    return c, nil
                },
            },
        }

        _, err := c.Get("http://"+node.IP+":"+node.Port+"/info")
        if err != nil {
            quitedNodes = append(quitedNodes, node)
        }
    }

    return quitedNodes, nil
}



