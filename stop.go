package main

import(
	"fmt"
	"github.com/BaileyZheng/diy_docker/container"
	"encoding/json"
	"strconv"
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"syscall"
	"os"
)

func stopContainer(containerName string){
	pid,err:=getContainerPidByName(containerName)
	if err!=nil{
		log.Errorf("Get container pid by name %s error %v",containerName,err)
		return
	}
	pidInt,err:=strconv.Atoi(pid)
	if err!=nil{
		log.Errorf("conver pid from string to int error %v",err)
		return
	}
	if err:=syscall.Kill(pidInt,syscall.SIGTERM);err!=nil{
		log.Errorf("Stop container %s error %v",containerName,err)
		return
	}
	containerInfo,err:=getContainerInfoByName(containerName)
	if err!=nil{
		log.Errorf("get container %s info error %v",containerName,err)
		return
	}
	containerInfo.Status=container.STOP
	containerInfo.Pid=""
	newContentBytes,err:=json.Marshal(containerInfo)
	if err!=nil{
		log.Errorf("json marshal %s error %v",containerInfo,err)
		return
	}
	dirURL:=fmt.Sprintf(container.DefaultInfoLocation,containerName)
	configFilePath:=dirURL+container.ConfigName
	if err:=ioutil.WriteFile(configFilePath,newContentBytes,0622);err!=nil{
		log.Errorf("write file %S error %v",configFilePath,err)
	}
}

func removeContainer(containerName string){
	containerInfo,err:=getContainerInfoByName(containerName)
	if err!=nil{
		log.Errorf("get container %s info error %v",containerName,err)
		return
	}
	if containerInfo.Status!=container.STOP{
		log.Errorf("Couldn't remove running container")
		return
	}
	dirURL:=fmt.Sprintf(container.DefaultInfoLocation,containerName)
	if err:=os.RemoveAll(dirURL);err!=nil{
		log.Errorf("remove file %s error %v",dirURL,err)
	}
}

func getContainerInfoByName(containerName string) (*container.ContainerInfo,error){
	dirURL:=fmt.Sprintf(container.DefaultInfoLocation,containerName)
	configFilePath:=dirURL+container.ConfigName
	contentBytes,err:=ioutil.ReadFile(configFilePath)
	if err!=nil{
		log.Errorf("read file %s error %v",configFilePath,err)
		return nil,err
	}
	var containerInfo container.ContainerInfo
	if err:=json.Unmarshal(contentBytes,&containerInfo); err!=nil{
		log.Errorf("getContainerInfoByName unmarshal error %v",err)
		return nil,err
	}
	return &containerInfo,nil
}
