package network

import(
	log "github.com/Sirupsen/logrus"
	"fmt"
	"net"
	"github.com/vishvananda/netlink"
	"os/exec"
	"strings"
)

type BridgeNetworkDriver struct{
}

func (d *BridgeNetworkDriver) Name() string{
	return "bridge"
}

func (d *BridgeNetworkDriver) Create(subnet string,name string) (*Network,error){
	ip,ipRange,_:=net.ParseCIDR(subnet)
	ipRange.IP=ip
	n:=&Network{
		Name: name,
		IpRange: ipRange,
		Driver: d.Name(),
	}
	if err:=d.initBridge(n);err!=nil{
		log.Errorf("init bridge error %v",err)
		return nil,err
	}
	return n,nil
}

func (d *BridgeNetworkDriver) initBridge(network *Network) error{
	bridgeName:=network.Name
	if err:=createBridgeInterface(bridgeName);err!=nil{
		log.Errorf("create bridge interface err %v",err)
		return err
	}
	if err:=setInterfaceIp(bridgeName,network.IpRange.String());err!=nil{
		log.Errorf("assign address %s on bridge %s error %v",network.IpRange,bridgeName,err)
		return err
	}
	if err:=setInterfaceUP(bridgeName);err!=nil{
		log.Errorf("set bridge %s up error %v",bridgeName,err)
		return err
	}
	if err:=setupIPTables(bridgeName,network.IpRange);err!=nil{
		log.Errorf("set up iptables for bridge %s error %v",bridgeName,err)
		return err
	}
	return nil
}

func createBridgeInterface(bridgeName string) error{
	if _,err:=net.InterfaceByName(bridgeName);err==nil||!strings.Contains(err.Error(),"no such network interface"){
		return err
	}
	la:=netlink.NewLinkAttrs()
	la.Name=bridgeName
	br:=&netlink.Bridge{LinkAttrs:la}
	if err:=netlink.LinkAdd(br);err!=nil{
		return fmt.Errorf("create bridge %s error %v",bridgeName,err)
	}
	return nil
}

func setInterfaceIp(bridgeName, rawIp string) error{
	iface,err:=netlink.LinkByName(bridgeName)
	if err!=nil{
		return fmt.Errorf("get interface %s error %v",bridgeName,err)
	}
	ipNet,err:=netlink.ParseIPNet(rawIp)
	if err!=nil{
		return fmt.Errorf("rawIP: %s",rawIp)
	}
	addr:=&netlink.Addr{IPNet:ipNet}
	return netlink.AddrAdd(iface,addr)
}

func setInterfaceUP(bridgeName string) error{
	iface,err:=netlink.LinkByName(bridgeName)
	if err!=nil{
		return fmt.Errorf("get interface error %v",err)
	}
	if err:=netlink.LinkSetUp(iface);err!=nil{
		return fmt.Errorf("link set up error %v",err)
	}
	return nil
}

func setupIPTables(bridgeName string, subnet *net.IPNet) error{
	iptablesCmd:=fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(),bridgeName)
	if output,err:=exec.Command("iptables",strings.Split(iptablesCmd," ")...).Output();err!=nil{
		log.Errorf("exec command iptables error %v, output %v",err,output)
		return err
	}
	return nil
}

func (d *BridgeNetworkDriver) Connect(network *Network, ep *Endpoint) error{
	bridgeName:=network.Name
	br,err:=netlink.LinkByName(bridgeName)
	if err!=nil{
		log.Errorf("find bridge by name err %v",err)
		return err
	}
	la:=netlink.NewLinkAttrs()
	la.Name=ep.ID[:5]
	la.MasterIndex=br.Attrs().Index
	ep.Device=netlink.Veth{
		LinkAttrs:la,
		PeerName:"cif-"+ep.ID[:5],
	}
	if err=netlink.LinkAdd(&ep.Device);err!=nil{
		return fmt.Errorf("add endpoint device error %v",err)
	}
	if err=netlink.LinkSetUp(&ep.Device);err!=nil{
		return fmt.Errorf("set up endpoint device error %v",err)
	}
	return nil
}

func (d *BridgeNetworkDriver) Delete(network Network) error{
	bridgeName:=network.Name
	br,err:=netlink.LinkByName(bridgeName)
	if err!=nil{
		return err
	}
	return netlink.LinkDel(br)
}
