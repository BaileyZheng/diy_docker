package container

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

func NewWorkSpace(volume,imageName,containerName string){
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName,imageName)
	if volume!=""{
		volumeURLs:=strings.Split(volume,":")
		length:=len(volumeURLs)
		if length==2&&volumeURLs[0]!="" && volumeURLs[1]!=""{
			MountVolume(volumeURLs,containerName)
			log.Infof("NewWorkSpace volume urls %q",volumeURLs)
		}else{
			log.Infof("volume parameter input is not correct.")
		}
	}
}

func PathExists(path string) (bool,error){
	_,err:=os.Stat(path)
	if err==nil{
		return true,nil
	}
	if os.IsNotExist(err){
		return false,nil
	}
	return false,err
}

func CreateReadOnlyLayer(imageName string) error{
	unTarFolderUrl:=RootUrl+"/"+imageName+"/"
	imageUrl:=RootUrl+"/"+imageName+".tar"
	exist,err:=PathExists(unTarFolderUrl)
	if err!=nil{
		log.Infof("Fail to judge whether dir %s exists. %v",unTarFolderUrl,err)
		return err
	}
	if !exist{
		if err:=os.MkdirAll(unTarFolderUrl,0622);err!=nil{
			log.Infof("Mkdir %s error %v",unTarFolderUrl,err)
			return err
		}
		if _,err:=exec.Command("tar","-xvf",imageUrl,"-C",unTarFolderUrl).CombinedOutput();err!=nil{
			log.Infof("Untar dir %s error %v",unTarFolderUrl,err)
			return err
		}
	}
	return nil
}

func CreateWriteLayer(containerName string){
	writeURL:=fmt.Sprintf(WriteLayerUrl,containerName)
	if err:=os.MkdirAll(writeURL,0777);err!=nil{
		log.Infof("Mkdir write layer dir %s error. %v",writeURL,err)
	}
}

func MountVolume(volumeURLs []string,containerName string) error{
	parentURL:=volumeURLs[0]
	if exist,_:=PathExists(parentURL);!exist{
		if err:=os.Mkdir(parentURL,0777);err!=nil{
			log.Infof("Mkdir parent dir %s error. %v", parentURL,err)
		}
	}
	containerURL:=volumeURLs[1]
	mntURL:=fmt.Sprintf(MntUrl,containerName)
	containerVolumeUrl:=mntURL+containerURL
	if exist,_:=PathExists(containerVolumeUrl);!exist{
		if err:=os.Mkdir(containerVolumeUrl,0777);err!=nil{
			log.Infof("Mkdir container dir %s error. %v",containerVolumeUrl,err)
		}
	}
	dirs:="dirs="+parentURL
	_,err:=exec.Command("mount","-t","aufs","-o",dirs,"none",containerVolumeUrl).CombinedOutput()
	if err!=nil{
		log.Infof("Mount volume failed. %v",err)
		return err
	}
	return nil
}

func CreateMountPoint(containerName, imageName string) error{
	mntURL:=fmt.Sprintf(MntUrl,containerName)
	if err:=os.MkdirAll(mntURL,0777);err!=nil{
		log.Infof("Mkdir mountpoint dir %s error. %v",mntURL,err)
		return err
	}
	tmpWriteLayer:=fmt.Sprintf(WriteLayerUrl,containerName)
	tmpImageLocation:=RootUrl+"/"+imageName
	mnturl:=fmt.Sprintf(MntUrl,containerName)
	dirs:="dirs="+tmpWriteLayer+":"+tmpImageLocation
	cmd:=exec.Command("mount","-t","aufs","-o",dirs,"none",mnturl)
	if err:=cmd.Run();err!=nil{
		log.Infof("run command for creating mount point %s of dirs %s failed %v",mnturl, dirs,err)
		return err
	}
	return nil
}

func DeleteWorkSpace(volume, containerName string){
	if volume!=""{
		volumeUrls:=strings.Split(volume,":")
		length:=len(volumeUrls)
		if length==2&&volumeUrls[0]!=""&&volumeUrls[1]!=""{
			DeleteMountPointWithVolume(volumeUrls[1],containerName)
		}else{
			DeleteMountPoint(containerName)
		}
	}else{
		DeleteMountPoint(containerName)
	}
	DeleteWriteLayer(containerName)
}

func DeleteMountPoint(containerName string) error{
	mntUrl:=fmt.Sprintf(MntUrl,containerName)
	if _,err:=exec.Command("umount",mntUrl).CombinedOutput();err!=nil{
		log.Errorf("Unmount %s error %v",mntUrl,err)
		return err
	}
	if err:=os.RemoveAll(mntUrl);err!=nil{
		log.Errorf("Remove mountpoint dir %s error %v",mntUrl,err)
		return err
	}
	return nil
}

func DeleteMountPointWithVolume(volumeUrl, containerName string) error{
	mntUrl:=fmt.Sprintf(MntUrl,containerName)
	containerVolumeUrl:=mntUrl+"/"+volumeUrl
	if _,err:=exec.Command("umount",containerVolumeUrl).CombinedOutput();err!=nil{
		log.Errorf("Unmount volume %s failed. %v",containerVolumeUrl,err)
		return err
	}
	DeleteMountPoint(containerName)
	return nil
}

func DeleteWriteLayer(containerName string){
	writeUrl:=fmt.Sprintf(WriteLayerUrl,containerName)
	if err:=os.RemoveAll(writeUrl);err!=nil{
		log.Errorf("Remove writeLayer dir %s error. %v",writeUrl,err)
	}
}
