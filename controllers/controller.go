package controllers

import (
	"github.com/gorilla/mux"
	"net/http"
	"terraformCloud/dao"
	"terraformCloud/models"
)

func CheckCloudCredentials(w http.ResponseWriter,r *http.Request) {
	var creds models.CloudInfo
	dao.GetCredentials(creds,w,r)
}

func CreateEc2Instance(w http.ResponseWriter, r *http.Request) {
	var ec2 models.Ec2Instance
	dao.CreateEc2Instance(ec2,w,r)
}

func CreateEc2InstanceWithDocker(w http.ResponseWriter, r *http.Request)  {
	var ec2 models.Ec2Instance
	dao.CreateEc2InstanceWithScriptFile(ec2,w,r)
}

func ListAllEc2InstanceInRegion(w http.ResponseWriter,r *http.Request)  {
	id := mux.Vars(r)
	dao.ListEc2Instances(id["userid"], w,r)
}

func GetAnEc2InstanceDetails(w http.ResponseWriter, r *http.Request)  {
	id := mux.Vars(r)
	instance := mux.Vars(r)
	dao.GetEc2InstanceDetail(id["userid"], instance["instance_id"], w,r)
}

func DestroyEc2Instance(w http.ResponseWriter, r *http.Request)  {
	id := mux.Vars(r)
	dao.DestroyInstance(id["userid"],w,r)
}

func DestroyEc2DockerInstance(w http.ResponseWriter, r *http.Request)  {
	id := mux.Vars(r)
	dao.DestroyDockerInstance(id["userid"],w,r)
}
//Azure

func CreateAzureVm(w http.ResponseWriter,r *http.Request)  {
	var vm models.AzureVM
	dao.CreateAzureVm(vm, w,r)
}

func CreateAzureDockerVm(w http.ResponseWriter,r *http.Request)  {
	var vm models.AzureVM
	dao.CreateAzureVmWithDocker(vm,w,r)
}

func ListAllAzureVM(w http.ResponseWriter,r *http.Request)  {
	id := mux.Vars(r)
	dao.ListAllAzureVM(id["userId"], w,r)
}
//
func GetAzureVMDetails(w http.ResponseWriter, r *http.Request)  {
	id := mux.Vars(r)
	vmname := mux.Vars(r)
	rg := mux.Vars(r)
	dao.GetAzureVMDetails(id["userId"], rg["resourceGroup"],vmname["vmName"], w,r)
}

func DestroyAzureDockerVm(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)
	dao.DestroyAzureDockerVM(id["userId"], w,r)
}

func DestroyAzureVm(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)
	dao.DestroyAzureVM(id["userId"], w,r)
}
