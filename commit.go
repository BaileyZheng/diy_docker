package main

import(
	"fmt"
	"github.com/BaileyZheng/diy_docker/container"
	"os/exec"
	log "github.com/Sirupsen/logrus"
)

func commitContainer(containerName,imageName string){
	mntUrl:=fmt.Sprintf(container.MntUrl,containerName)
	mntUrl+="/"
	imageTar:=container.RootUrl+"/"+imageName+".tar"
	if _,err:=exec.Command("tar","-czf",imageTar,"-C",mntUrl,".").CombinedOutput();err!=nil{
		log.Errorf("Tar folder %s error %v",mntUrl,err)
	}
}
