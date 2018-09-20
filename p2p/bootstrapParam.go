package p2p

//Kind of predicatable bootstrapper because it's generates known peers by only using an ip address and port domain info
//peerid's will be autogenerated in a predictable way (seed = port number)
//It will use a Kademlia engine to select closest peers

type BootstrapParam struct {
	PortStart int
	PortEnd   int
	Address   string
}
