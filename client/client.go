package client

import (
	// "os"
	"fmt"
	"time"
	"sync"
	"net/url"
	// "os/exec"
	// "syscall"
	"net/http"
	// "os/signal"
	"io/ioutil"
	// "encoding/json"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/seveirbian/gear/types"
	"github.com/seveirbian/gear/pkg"
)

var (
	logger = logrus.WithField("gear", "client")
)

var (
	port = "2020"
	cli Client = Client{}
)

type Client struct {
	Self  types.Node

	Echo  *echo.Echo

	EnableP2p bool

	Manager types.Node
	Monitor types.Node

	NodesMu sync.RWMutex
	Nodes map[uint64]types.Node
}

func Init(managerIP string, managerPort string, monitorIP string, monitorPort string, enbaleP2p bool) (*Client, error) {
	// 1. create new echo instance
	e := echo.New()

	// 2. add routes
	e.GET("/info", handleInfo)
	e.POST("/get/:CID", handleGet)
	e.POST("/download/:CID", handleDownload)
	e.POST("/upload", handleUpload)

	e.POST("/recorded/:IMAGE", handleRecorded)
	e.POST("/record/:IMAGE/:CID", handleRecord)
	e.POST("/report/:IMAGE", handleReport)

	// 3. get self's IP
	ip := pkg.GetSelfIp()

	// 4. create self ID
	id := pkg.CreateIdFromIP(ip)

	// 5. fill the manager's fileds
	cli.Self = types.Node{
		ID:   id,
		IP:   ip,
		Port: port,
	}
	cli.Echo = e
	cli.Manager = types.Node{
		ID: pkg.CreateIdFromIP(managerIP), 
		IP: managerIP, 
		Port: managerPort, 
	}

	cli.Monitor = types.Node{
		ID: pkg.CreateIdFromIP(monitorIP), 
		IP: monitorIP, 
		Port: monitorPort, 
	}
	cli.EnableP2p = enbaleP2p

	return &cli, nil
}

func (c *Client) Start() {
	if c.EnableP2p {
		// 将自己的信息添加到manager中并且从manager中获取集群其他节点、etcd和nfs信息
		c.joinCluster()
		fmt.Println(c.Nodes)

		// 一个协程来定期(60秒)更新集群节点信息
		go updateNodes(c)
	}

	c.start()
}

func (c *Client) start() {
	c.Echo.Logger.Fatal(c.Echo.Start(":" + c.Self.Port))
}

func (c *Client) joinCluster() {
	// 1. join to the p2p cluster
	err := c.join()
	if err != nil {
		logger.Fatal("Fail to join the cluster...")
	}

	// 2. contact to manager node to get other nodes' info
	nodes, err := c.getClusterNodes()
	if err != nil {
		logger.Fatal("Fail to get the cluster nodes info...")
	}

	c.Nodes = map[uint64]types.Node{}

	for _, node := range(nodes) {
		c.Nodes[node.ID] = node
	}
}

func (c *Client) join() error {
	_, err := http.PostForm("http://"+c.Manager.IP+":"+c.Manager.Port+"/join/"+c.Self.IP+"/"+c.Self.Port, url.Values{})

	return err
}

func (c *Client) getClusterNodes() ([]types.Node, error) {
	resp, err := http.Get("http://"+c.Manager.IP+":"+c.Manager.Port+"/nodes")
	if err != nil {
		logger.Fatal("Fail to get the cluster info...")
		return []types.Node{}, err
	}

	nodesInStirng, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatal("Fail to read response's body...")
		return []types.Node{}, err
	}

	nodes := pkg.GetNodes(string(nodesInStirng))

	return nodes, nil
}

func updateNodes(c *Client) {
	for {
		nodes, err := c.getClusterNodes()
		if err != nil {
			logger.Fatal("Fail to get cluster nodes...")
		}

		c.NodesMu.Lock()
		for _, node := range(nodes) {
			if _, ok := c.Nodes[node.ID]; !ok {
				c.Nodes[node.ID] = node
			}
		}
		c.NodesMu.Unlock()

		time.Sleep(60*time.Second)
	}
}










