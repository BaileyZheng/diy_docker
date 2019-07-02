package network

import(
	"net"
	"os"
	log "github.com/Sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"os/exec"
	"github.com/BaileyZheng/diy_docker/container"
	"fmt"
	"encoding/json"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type Network struct{
	Name string
	IpRange *net.IPNet
	Driver string
}

type Endpoint struct{
	ID string `json:"id"`
	Device netlink.Veth `json:"dev"`
	IPAddress net.IP `json:"ip"`
	MacAddress net.HardwareAddr `json:"mac"`
	PortMapping []string `json:"portmapping"`
	Network *Network
}

type NetworkDriver interface{
	Name() string
	Create(subnet string,name string) (*Network,error)
	Connect(network *Network,endpoint *Endpoint) error
}

var (
	defaultNetworkPath="/var/run/mydocker/network/network/"
	drivers=map[string]NetworkDriver{}
	networks=map[string]*Network{}
)

func Init() error{
	var bridgeDriver=BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()]=&bridgeDriver
	if _,err:=os.Stat(defaultNetworkPath);err!=nil{
		if os.IsNotExist(err){
			os.MkdirAll(defaultNetworkPath,0644)
		}else{
			return err
		}
	}
	filepath.Walk(defaultNetworkPath,func(nwPath string,info os.FileInfo,err error) error{
		if info.IsDir(){
			return nil
		}
		_,nwName:=path.Split(nwPath)
		nw:=&Network{
			Name:nwName,
		}
		if err:=nw.load(nwPath);err!=nil{
			log.Errorf("load network %s error %v",nwPath,err)
		}
		networks[nwName]=nw
		return nil
	})
	return nil
}

func CreateNetwork(driver,subnet,name string) error{
	_,cidr,_:=net.ParseCIDR(subnet)
	gatewayIp,err:=ipAllocator.Allocate(cidr)
	if err!=nil{
		log.Errorf("ipallocator allocate error %v",err)
		return err
	}
	cidr.IP=gatewayIp
	nw,err:=drivers[driver].Create(cidr.String(),name)
	if err!=nil{
		log.Errorf("create network error %v",err)
		return err
	}
	return nw.dump(defaultNetworkPath)
}

func (nw *Network) dump(dumpPath string) error{
	if _,err:=os.Stat(dumpPath);err!=nil{
		if os.IsNotExist(err){
			os.MkdirAll(dumpPath,0644)
		}else{
			return err
		}
	}
	nwPath:=path.Join(dumpPath,nw.Name)
	nwFile,err:=os.OpenFile(nwPath,os.O_TRUNC|os.O_WRONLY|os.O_CREATE,0644)
	if err!=nil{
		log.Errorf("open file %s error: %v",nwPath,err)
		return err
	}
	defer nwFile.Close()
	nwJson,err:=json.Marshal(nw)
	if err!=nil{
		log.Errorf("json marshal error %v",err)
		return err
	}
	if _,err:=nwFile.Write(nwJson);err!=nil{
		log.Errorf("write error %v",err)
		return err
	}
	return nil
}

func (nw *Network) load(dumpPath string) error{
	nwConfigFile,err:=os.Open(dumpPath)
	defer nwConfigFile.Close()
	if err!=nil{
		log.Errorf("open file %s error %v",dumpPath,err)
		return err
	}
	nwJson:=make([]byte,2000)
	n,err:=nwConfigFile.Read(nwJson)
	if err!=nil{
		log.Errorf("read file %s error %v",dumpPath,err)
		return err
	}
	if err:=json.Unmarshal(nwJson[:n],nw);err!=nil{
		log.Errorf("unmarshal json error %v",err)
		return err
	}
	return nil
}

func Connect(networkName string, cinfo *container.ContainerInfo) error{
	network,exist:=networks[networkName]
	log.Errorf("network name %s, network ipRange %s, network driver %s",network.Name,network.IpRange.String(),network.Driver)
	if !exist{
		return fmt.Errorf("network %s does not exist",networkName)
	}
	ip,err:=ipAllocator.Allocate(network.IpRange)
	if err!=nil{
		log.Errorf("allocate ip error %v",err)
		return err
	}else{
		log.Infof("allocate ip %s",ip)
	}
	ep:=&Endpoint{
		ID: fmt.Sprintf("%s-%s",cinfo.Id,networkName),
		IPAddress: ip,
		Network: network,
		PortMapping: cinfo.PortMapping, 
	}
	if err=drivers[network.Driver].Connect(network,ep);err!=nil{
		log.Errorf("driver connect network and endpoint error %v",err)
		return err
	}
	if err=configEndpointIpAddressAndRoute(ep,cinfo);err!=nil{
		log.Errorf("config endpoint ip address and route error %v",err)
		return err
	}
	return configPortMapping(ep,cinfo)
}

func configEndpointIpAddressAndRoute(ep *Endpoint,cinfo *container.ContainerInfo) error{
	peerLink,err:=netlink.LinkByName(ep.Device.PeerName)
	if err!=nil{
		return fmt.Errorf("config endpoint error %v",err)
	}
	defer enterContainerNetns(&peerLink,cinfo)()
	interfaceIP:=*ep.Network.IpRange
	interfaceIP.IP=ep.IPAddress
	if err=setInterfaceIp(ep.Device.PeerName,interfaceIP.String());err!=nil{
		return fmt.Errorf("set interface %s ip %s error %v",ep.Device.PeerName,ep.IPAddress,err)
	}
	if err=setInterfaceUP(ep.Device.PeerName);err!=nil{
		return fmt.Errorf("set interface up error %v",err)
	}
	if err=setInterfaceUP("lo");err!=nil{
		return fmt.Errorf("set interface lo up error %v",err)
	}
	_,cidr,_:=net.ParseCIDR("0.0.0.0/0")
	defaultRoute:=&netlink.Route{
		LinkIndex:peerLink.Attrs().Index,
		Gw:ep.Network.IpRange.IP,
		Dst:cidr,
	}
	if err=netlink.RouteAdd(defaultRoute);err!=nil{
		return fmt.Errorf("add route error %v",err)
	}
	return nil
}

func enterContainerNetns(enLink *netlink.Link,cinfo *container.ContainerInfo) func(){
	f,err:=os.OpenFile(fmt.Sprintf("/proc/%s/ns/net",cinfo.Pid),os.O_RDONLY,0)
	if err!=nil{
		log.Errorf("get container net namespace error %v",err)
	}
	nsFD:=f.Fd()
	runtime.LockOSThread()
	if err=netlink.LinkSetNsFd(*enLink,int(nsFD));err!=nil{
		log.Errorf("set link setns error %v",err)
	}
	origns,err:=netns.Get()
	if err!=nil{
		log.Errorf("get currentnetns error %v",err)
	}
	if err=netns.Set(netns.NsHandle(nsFD));err!=nil{
		log.Errorf("netns set error %v",err)
	}
	return func(){
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func configPortMapping(ep *Endpoint,cinfo *container.ContainerInfo) error{
	for _,pm:=range ep.PortMapping{
		portMapping:=strings.Split(pm,":")
		if len(portMapping)!=2{
			log.Errorf("port mapping format error %v",pm)
			continue
		}
		iptablesCmd:=fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",portMapping[0],ep.IPAddress.String(),portMapping[1])
		cmd:=exec.Command("iptables",strings.Split(iptablesCmd," ")...)
		output,err:=cmd.Output()
		if err!=nil{
			log.Errorf("iptables Output %v",output)
			continue
		}
	}
	return nil
}
