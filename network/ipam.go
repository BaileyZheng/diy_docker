package network

import(
	"path"
	"os"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"strings"
	"net"
)

const ipamDefaultAllocatorPath="/var/run/mydocker/network/ipam/subnet.json"

type IPAM struct{
	SubnetAllocatorPath string
	Subnets *map[string]string
}

var ipAllocator=&IPAM{
	SubnetAllocatorPath:ipamDefaultAllocatorPath,
}

func (ipam *IPAM) dump() error{
	ipamConfigFileDir,_:=path.Split(ipam.SubnetAllocatorPath)
	if _,err:=os.Stat(ipamConfigFileDir);err!=nil{
		if os.IsNotExist(err){
			os.MkdirAll(ipamConfigFileDir,0644)
		}else{
			return err
		}
	}
	subnetConfigFile,err:=os.OpenFile(ipam.SubnetAllocatorPath,os.O_TRUNC|os.O_WRONLY|os.O_CREATE,0644)
	defer subnetConfigFile.Close()
	if err!=nil{
		log.Errorf("open subnetAllocatorPath error %v",err)
		return err
	}
	ipamConfigJson,err:=json.Marshal(ipam.Subnets)
	if err!=nil{
		log.Errorf("json marshal error %v",err)
		return err
	}
	_,err=subnetConfigFile.Write(ipamConfigJson)
	if err!=nil{
		log.Errorf("write subnetAllocatorPath error %v",err)
		return err
	}
	return nil
}

func (ipam *IPAM) load() error{
	if _,err:=os.Stat(ipam.SubnetAllocatorPath);err!=nil{
		if os.IsNotExist(err){
			return nil
		}else{
			return err
		}
	}
	subnetConfigFile,err:=os.Open(ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()
	if err!=nil{
		log.Errorf("open file error %v",err)
		return err
	}
	subnetJson:=make([]byte,2000)
	n,err:=subnetConfigFile.Read(subnetJson)
	if err!=nil{
		log.Errorf("read file error %v",err)
		return err
	}
	if err:=json.Unmarshal(subnetJson[:n],ipam.Subnets);err!=nil{
		log.Errorf("json unmarshal error %v",err)
		return err
	}
	return nil
}

func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP,err error){
	ipam.Subnets=&map[string]string{}
	if err:=ipam.load();err!=nil{
		log.Errorf("load allocation info error %v",err)
	}
	one,size:=subnet.Mask.Size()
	if _,exist:=(*ipam.Subnets)[subnet.String()];!exist{
		(*ipam.Subnets)[subnet.String()]=strings.Repeat("0",1<<uint8(size-one))
	}
	for c:=range((*ipam.Subnets)[subnet.String()]){
		if (*ipam.Subnets)[subnet.String()][c]=='0'{
			ipalloc:=[]byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c]='1'
			(*ipam.Subnets)[subnet.String()]=string(ipalloc)
			ip=subnet.IP
			for t:=uint(4);t>0;t-=1{
				[]byte(ip)[4-t]+=uint8(c>>((t-1)*8))
			}
			ip[3]+=1
			break
		}
	}
	ipam.dump()
	return
}
