package types

import ()

var ()

type Node struct {
	ID   uint64
	IP   string
	Port string
}

type Etcd struct {
	IP string `json:"ip"`
	Port string `json:"port"`
}

type NFS struct {
	IP string `json:"ip"`
	Path string `json:"path"`
}

type Config struct {
	Etcd Etcd `json:"etcd"`
	NFS NFS	`json:"nfs"`
}

// etcd 相关的结构体
type EtcdNode struct {
	Key string `json:"key"`
	Value string `json:"value"`
	ModifiedIndex int `json:"modifiedIndex"`
	CreatedIndex int `json:"createdIndex"`
}

type Response struct {
	Action string `json:"action"`
	EtcdNode EtcdNode `json:"node"`
}

type Image struct {
	Repository string
	Tag string
}

type MonitorFile struct {
	Hash string
	RelativePath string
}