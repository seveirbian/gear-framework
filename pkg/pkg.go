package pkg

import (
	"io"
	"os"
	"fmt"
	"net"
	"strings"
	"strconv"
	"hash/fnv"
	"crypto/md5"
	"github.com/sirupsen/logrus"
	"github.com/seveirbian/gear/types"
)

var (

)

func GetSelfIp() string {
	conn, err := net.Dial("udp", "google.com:80")
    if err != nil {
        logrus.Fatal("Fail to dial google.com:80")
    }
    defer conn.Close()

	return strings.Split(conn.LocalAddr().String(), ":")[0]
}

func CreateIdFromIP(ip string) uint64 {
	f := fnv.New64()
	f.Write([]byte(ip))
	return f.Sum64()
}

func GetNodes(nodesInString string) []types.Node {
	var nodes = []types.Node{}

	nodesSlices := strings.Split(nodesInString, ";")
	for _, nodesSlice := range(nodesSlices[:len(nodesSlices)-1]) {
		nodeInfo := strings.Split(nodesSlice, ":")
		id, err := strconv.ParseUint(nodeInfo[0], 10, 64)
		if err != nil {
			logrus.Fatal("Fail to conv uint64...")
		}
		
		nodes = append(nodes, types.Node{
			ID: id, 
			IP: nodeInfo[1], 
			Port: nodeInfo[2], 
		})
	}

	return nodes
}

func Hash(s string) uint64 {
	f := fnv.New64()
	f.Write([]byte(s))
	return f.Sum64()
}

func HashAFileInMD5(path string) string {
	f, err := os.Open(path)
	if err != nil {
		logrus.Fatal("Fail to open the file needs to be md5ed...")
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		logrus.Fatal("Fail to copy file to md5...")
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}










