package container

import(
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
	"fmt"
)

type ContainerInfo struct{
	Pid	string `json:"pid"`
	Id	string `json:"id"`
	Name	string `json:"name"`
	Command string `json:"command"`
	CreatedTime string `json:"createTime"`
	Status	string `json:"status"`
	Volume  string `json:"volume"`
}

var(
	RUNNING		string="running"
	STOP		string="stopped"
	Exit		string="exited"
	DefaultInfoLocation string="/var/run/mydocker/%s/"
	ConfigName	string="config.json"
	ContainerLogFile string="container.log"
	RootUrl 	string="/root"
	MntUrl		string="/root/mnt/%s"
	WriteLayerUrl	string="/root/writeLayer/%s"
)

func NewParentProcess(tty bool, containerName, volume, imageName string) (*exec.Cmd, *os.File){
	readPipe, writePipe, err := NewPipe()
	if err != nil{
		log.Errorf("New pipe error %v", err)
		return nil,nil
	}
	cmd:=exec.Command("/proc/self/exe","init")
	cmd.SysProcAttr=&syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS|syscall.CLONE_NEWPID|syscall.CLONE_NEWNS|syscall.CLONE_NEWNET|syscall.CLONE_NEWIPC,
	}
	if tty{
		cmd.Stdin=os.Stdin
		cmd.Stdout=os.Stdout
		cmd.Stderr=os.Stderr
	}else{
		dirURL:=fmt.Sprintf(DefaultInfoLocation,containerName)
		if err:=os.MkdirAll(dirURL,0622);err!=nil{
			log.Errorf("NewParentProcess mkdir %s error %v",dirURL,err)
			return nil,nil
		}
		stdLogFilePath:=dirURL+ContainerLogFile
		stdLogFile,err:=os.Create(stdLogFilePath)
		if err!=nil{
			log.Errorf("NewParentProcess create file %s error %v",stdLogFilePath,err)
			return nil,nil
		}
		cmd.Stdout=stdLogFile
	}
	cmd.ExtraFiles=[]*os.File{readPipe}
	NewWorkSpace(volume,imageName,containerName)
	cmd.Dir=fmt.Sprintf(MntUrl,containerName)
	return cmd,writePipe
}

func NewPipe()(*os.File, *os.File, error){
	read,write,err := os.Pipe()
	if err!=nil{
		return nil,nil,err
	}
	return read,write,nil
}
