package main

import (
	"github.com/BaileyZheng/diy_docker/container"
	"github.com/BaileyZheng/diy_docker/cgroups/subsystems"
	"github.com/BaileyZheng/diy_docker/cgroups"
	log "github.com/Sirupsen/logrus"
	"os"
	"strings"
	"math/rand"
	"time"
	"encoding/json"
	"strconv"
	"fmt"
)

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig, containerName, volume, imageName string){
	if containerName==""{
		containerName=randStringBytes(10)
	}
	if imageName==""{
		imageName="busybox"
	}
	parent,writePipe:=container.NewParentProcess(tty,containerName,volume,imageName)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err:= parent.Start(); err!=nil{
		log.Error(err)
	}
	containerName,err:=recordContainerInfo(parent.Process.Pid,comArray,containerName,volume)
	if err!=nil{
		log.Errorf("Record container info error %v", err)
		return
	}
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	sendInitCommand(comArray,writePipe)
	if tty {
		parent.Wait()
		deleteContainerInfo(containerName)
		container.DeleteWorkSpace(volume,containerName)
	}
}

func sendInitCommand(comArray []string, writePipe *os.File){
	command := strings.Join(comArray," ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

func randStringBytes(n int) string{
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b:=make([]byte,n)
	for i:=range b{
		b[i]=letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func recordContainerInfo(containerPID int, commandArray []string, containerName, volume string) (string, error){
	id:=randStringBytes(10)
	createTime:=time.Now().Format("2006-01-02 03:04:05")
	command := strings.Join(commandArray,"")
	if containerName == "" {
		containerName=id
	}
	containerInfo:=&container.ContainerInfo{
		Id:id,
		Pid:strconv.Itoa(containerPID),
		Name:containerName,
		Command:command,
		CreatedTime:createTime,
		Status:container.RUNNING,
		Volume:volume,
	}
	
	jsonBytes, err := json.Marshal(containerInfo)
	if err!=nil{
		log.Errorf("Record container info error %v",err)
		return "",err
	}
	jsonStr:=string(jsonBytes)
	dirUrl:=fmt.Sprintf(container.DefaultInfoLocation,containerName)
	if err:=os.MkdirAll(dirUrl,0622);err!=nil{
		log.Errorf("Mkdir error %s error %v",dirUrl,err)
		return "",err
	}
	fileName:=dirUrl+"/"+container.ConfigName
	file,err:=os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("Create file %s error %v",fileName,err)
		return "",err
	}
	if _,err:=file.WriteString(jsonStr);err!=nil{
		log.Errorf("File write string error %v",err)
		return "",err
	}
	return containerName,nil
}

func deleteContainerInfo(containerId string){
	dirURL:=fmt.Sprintf(container.DefaultInfoLocation,containerId)
	if err:=os.RemoveAll(dirURL);err!=nil{
		log.Errorf("Remove dir %s error %v",dirURL,err)
	}
}
