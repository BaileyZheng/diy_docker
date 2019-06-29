package network

import(
	"net"
	"os"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
	"path"
)

type Network struct{
	Name string
	IpRange *net.IPNet
	Driver string
}

type NetworkDriver interface{
	Name() string
	Create(subnet string,name string) (*Network,error)
}

var (
	defaultNetworkPath="/var/run/mydocker/network/network/"
	drivers=map[string]NetworkDriver{}
	networks=map[string]*Network{}
)

func Init() error{
	var bridgeDriver=BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()]=&bridgeDriver
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
