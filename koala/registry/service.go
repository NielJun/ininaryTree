package registry

// abstract of service
type Service struct {
	// name of the service
	Name string `json:"name"`
	// service nodes
	Nodes []*Node `json:"nodes"`
}

// the abstract node
type Node struct {
	Id   string `json:"id"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
