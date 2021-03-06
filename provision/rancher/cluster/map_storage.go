package cluster

import (
	"errors"
	"sync"
	//	"net"
	"time"
	//"math"
)

var (
	ErrNoSuchNode            = errors.New("No such node in storage")
	ErrNoSuchContainer       = errors.New("No such container in storage")
	ErrNoSuchImage           = errors.New("No such image in storage")
	ErrDuplicatedNodeAddress = errors.New("Node address shouldn't repeat")
)

type MapStorage struct {
	cMap map[string]string
	//iMap    map[string]*Image
	nodes   []Node
	nodeMap map[string]*Node
	ipindex map[string]*IPIndex
	cMut    sync.Mutex
	iMut    sync.Mutex
	nMut    sync.Mutex
	ipMut   sync.Mutex
}

func (s *MapStorage) StoreContainerByName(containerID, Name string) error {
	s.cMut.Lock()
	defer s.cMut.Unlock()
	if s.cMap == nil {
		s.cMap = make(map[string]string)
	}
	s.cMap[Name] = containerID
	return nil
}

func (s *MapStorage) RetrieveContainerByName(Name string) (string, error) {
	s.cMut.Lock()
	defer s.cMut.Unlock()
	container, ok := s.cMap[Name]
	if !ok {
		return "", ErrNoSuchContainer
	}
	return container, nil
}

func (s *MapStorage) StoreContainer(containerID, hostID string) error {
	s.cMut.Lock()
	defer s.cMut.Unlock()
	if s.cMap == nil {
		s.cMap = make(map[string]string)
	}
	s.cMap[containerID] = hostID
	return nil
}

func (s *MapStorage) RetrieveContainer(containerID string) (string, error) {
	s.cMut.Lock()
	defer s.cMut.Unlock()
	host, ok := s.cMap[containerID]
	if !ok {
		return "", ErrNoSuchContainer
	}
	return host, nil
}

func (s *MapStorage) RemoveContainer(containerID string) error {
	s.cMut.Lock()
	defer s.cMut.Unlock()
	delete(s.cMap, containerID)
	return nil
}

func (s *MapStorage) RetrieveContainers() ([]Container, error) {
	s.cMut.Lock()
	defer s.cMut.Unlock()
	entries := make([]Container, 0, len(s.cMap))
	for k, v := range s.cMap {
		entries = append(entries, Container{Id: k, Host: v})
	}
	return entries, nil
}

func (s *MapStorage) updateNodeMap() {
	s.nodeMap = make(map[string]*Node)
	for i := range s.nodes {
		s.nodeMap[s.nodes[i].Address] = &s.nodes[i]
	}
}

func (s *MapStorage) StoreNode(node Node) error {
	s.nMut.Lock()
	defer s.nMut.Unlock()
	for _, n := range s.nodes {
		if n.Address == node.Address {
			return ErrDuplicatedNodeAddress
		}
	}
	if node.Metadata == nil {
		node.Metadata = make(map[string]string)
	}
	s.nodes = append(s.nodes, node)
	s.updateNodeMap()
	return nil
}

func deepCopyNode(n Node) Node {
	newMap := map[string]string{}
	for k, v := range n.Metadata {
		newMap[k] = v
	}
	n.Metadata = newMap
	return n
}

func (s *MapStorage) RetrieveNodes() ([]Node, error) {
	s.nMut.Lock()
	defer s.nMut.Unlock()
	dst := make([]Node, len(s.nodes))
	for i := range s.nodes {
		dst[i] = deepCopyNode(s.nodes[i])
	}
	return dst, nil
}

func (s *MapStorage) RetrieveNode(address string) (Node, error) {
	s.nMut.Lock()
	defer s.nMut.Unlock()
	if s.nodeMap == nil {
		s.nodeMap = make(map[string]*Node)
	}
	node, ok := s.nodeMap[address]
	if !ok {
		return Node{}, ErrNoSuchNode
	}
	return deepCopyNode(*node), nil
}

func (s *MapStorage) UpdateNode(node Node) error {
	s.nMut.Lock()
	defer s.nMut.Unlock()
	if s.nodeMap == nil {
		s.nodeMap = make(map[string]*Node)
	}
	_, ok := s.nodeMap[node.Address]
	if !ok {
		return ErrNoSuchNode
	}
	*s.nodeMap[node.Address] = node
	return nil
}

func (s *MapStorage) RetrieveNodesByMetadata(metadata map[string]string) ([]Node, error) {
	s.nMut.Lock()
	defer s.nMut.Unlock()
	filteredNodes := []Node{}
	for _, node := range s.nodes {
		for key, value := range metadata {
			nodeVal, ok := node.Metadata[key]
			if ok && nodeVal == value {
				filteredNodes = append(filteredNodes, node)
			}
		}
	}
	return filteredNodes, nil
}

func (s *MapStorage) RemoveNode(addr string) error {
	s.nMut.Lock()
	defer s.nMut.Unlock()
	index := -1
	for i, node := range s.nodes {
		if node.Address == addr {
			index = i
		}
	}
	if index < 0 {
		return ErrNoSuchNode
	}
	copy(s.nodes[index:], s.nodes[index+1:])
	s.nodes = s.nodes[:len(s.nodes)-1]
	s.updateNodeMap()
	return nil
}

func (s *MapStorage) LockNodeForHealing(address string, isFailure bool, timeout time.Duration) (bool, error) {
	s.nMut.Lock()
	defer s.nMut.Unlock()
	n, present := s.nodeMap[address]
	if !present {
		return false, ErrNoSuchNode
	}
	now := time.Now().UTC()
	if n.Healing.LockedUntil.After(now) {
		return false, nil
	}
	n.Healing.LockedUntil = now.Add(timeout)
	n.Healing.IsFailure = isFailure
	return true, nil
}

func (s *MapStorage) ExtendNodeLock(address string, timeout time.Duration) error {
	s.nMut.Lock()
	defer s.nMut.Unlock()
	n, present := s.nodeMap[address]
	if !present {
		return ErrNoSuchNode
	}
	now := time.Now().UTC()
	n.Healing.LockedUntil = now.Add(timeout)
	return nil
}

func (s *MapStorage) UnlockNode(address string) error {
	s.nMut.Lock()
	defer s.nMut.Unlock()
	n, present := s.nodeMap[address]
	if !present {
		return ErrNoSuchNode
	}
	n.Healing = HealingData{}
	return nil
}

/*
func (s *MapStorage) StoreImage(repo, id, host string) error {
	s.iMut.Lock()
	defer s.iMut.Unlock()
	if s.iMap == nil {
		s.iMap = make(map[string]*Image)
	}
	img, _ := s.iMap[repo]
	if img == nil {
		img = &Image{Repository: repo, History: []ImageHistory{}}
		s.iMap[repo] = img
	}
	hasId := false
	for _, entry := range img.History {
		if entry.ImageId == id && entry.Node == host {
			hasId = true
			break
		}
	}
	if !hasId {
		img.History = append(img.History, ImageHistory{Node: host, ImageId: id})
	}
	img.LastNode = host
	img.LastId = id
	return nil
}

func (s *MapStorage) RetrieveImage(repo string) (Image, error) {
	s.iMut.Lock()
	defer s.iMut.Unlock()
	image, ok := s.iMap[repo]
	if !ok {
		return Image{}, ErrNoSuchImage
	}
	if len(image.History) == 0 {
		return Image{}, ErrNoSuchImage
	}
	return *image, nil
}

func (s *MapStorage) RemoveImage(repo, id, host string) error {
	s.iMut.Lock()
	defer s.iMut.Unlock()
	image, ok := s.iMap[repo]
	if !ok {
		return ErrNoSuchImage
	}
	newHistory := []ImageHistory{}
	for _, entry := range image.History {
		if entry.ImageId != id || entry.Node != host {
			newHistory = append(newHistory, entry)
		}
	}
	image.History = newHistory
	return nil
}

func (s *MapStorage) RetrieveImages() ([]Image, error) {
	s.iMut.Lock()
	defer s.iMut.Unlock()
	images := make([]Image, 0, len(s.iMap))
	for _, img := range s.iMap {
		images = append(images, *img)
	}
	return images, nil
}
*/
type IPIndex struct {
	Ip     string
	Subnet string
	Index  uint
}
