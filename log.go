package main

import(
	"fmt"
	"github.com/BaileyZheng/diy_docker/container"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

func logContainer(containerName string){
	dirURL:=fmt.Sprintf(container.DefaultInfoLocation,containerName)
	logFileLocation:=dirURL+container.ContainerLogFile
	file,err:=os.Open(logFileLocation)
	defer file.Close()
	if err != nil{
		log.Errorf("Log container open file %s error %v", logFileLocation,err)
		return
	}
	content,err:=ioutil.ReadAll(file)
	if err != nil{
		log.Errorf("Log container read file error %v", err)
		return
	}
	fmt.Fprint(os.Stdout,string(content))
}
