package manager

import (
	"strconv"
	"net/http"
	"github.com/labstack/echo"
	"github.com/seveirbian/gear/pkg"
	"github.com/seveirbian/gear/types"
)

var ()

func handleNodes(c echo.Context) error {
	var resp string

	mgr.NodesMu.RLock()
 
	for _, node := range(mgr.Nodes) {
		resp = resp + strconv.FormatUint(node.ID, 10) + ":" + node.IP + ":" + node.Port + ";"
	}

	defer mgr.NodesMu.RUnlock()
	
	return c.String(http.StatusOK, resp)
}

func handleJoin(c echo.Context) error {
	mIP := c.Param("IP")
	mPort := c.Param("Port")

	id := pkg.CreateIdFromIP(mIP)

	// 对Nodes数据结构加读锁
	mgr.NodesMu.RLock()

	_, ok := mgr.Nodes[id]

	// 解锁
	mgr.NodesMu.RUnlock()

	if !ok {
		// 对Nodes数据结构加写锁
		mgr.NodesMu.Lock()

		mgr.Nodes[id] = types.Node{
			ID:   id,
	        IP:   mIP,
	        Port: mPort,
		}

		// 解锁
		mgr.NodesMu.Unlock()
	}

	var EtcdNFS = types.Config{
		Etcd: types.Etcd{
			IP: mgr.Etcd.IP, 
			Port: mgr.Etcd.Port, 
		}, 
		NFS: types.NFS{
			IP: mgr.NFS.IP, 
			Path: mgr.NFS.Path, 
		},
	}

	return c.JSON(http.StatusOK, EtcdNFS)
}












