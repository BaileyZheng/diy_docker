package nsenter

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

__attribute__((constructor)) void enter_namespace(void){
	char *mydocker_pid;
	mydocker_pid=getenv("mydocker_pid");
	if(mydocker_pid){
	}else{
		return;
	}
	char *mydocker_cmd;
	mydocker_cmd=getenv("mydocker_cmd");
	if(mydocker_cmd){
	}else{
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[]={"ipc","uts","net","pid","mnt"};
	for(i=0;i<5;i++){
		sprintf(nspath,"/proc/%s/ns/%s",mydocker_pid,namespaces[i]);
		int fd=open(nspath,O_RDONLY);
		if(setns(fd,0)==-1){
		}
		close(fd);
	}
	int res=system(mydocker_cmd);
	exit(0);
	return;
}
*/
import "C"
