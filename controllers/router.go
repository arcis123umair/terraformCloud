package controllers

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func Router()  {
	r := mux.NewRouter()

	//Checking Cloud Credentials
	r.HandleFunc("/api/terraform/checkCredentials", CheckCloudCredentials).Methods("POST")

	//EC2 Instance
	r.HandleFunc("/api/terraform/createEc2Instance", CreateEc2Instance).Methods("POST")
	r.HandleFunc("/api/terraform/createDockerInstance", CreateEc2InstanceWithDocker).Methods("POST")
	r.HandleFunc("/api/terraform/listInstances/{userid}", ListAllEc2InstanceInRegion).Methods("GET")
	r.HandleFunc("/api/terraform/getInstanceDetail/{userid}/{instance_id}", GetAnEc2InstanceDetails).Methods("GET")
	r.HandleFunc("/api/terraform/destroyInstance/{userid}", DestroyEc2Instance).Methods("DELETE")
	r.HandleFunc("/api/terraform/destroyDockerInstance/{userid}", DestroyEc2DockerInstance).Methods("DELETE")

	//Azure Vm's
	r.HandleFunc("/api/terraform/createAzureVM", CreateAzureVm).Methods("POST")
	r.HandleFunc("/api/terraform/createAzureDockerVM", CreateAzureDockerVm).Methods("POST")
	r.HandleFunc("/api/terraform/listAzureVM/{userId}", ListAllAzureVM).Methods("GET")
	r.HandleFunc("/api/terraform/getAzureVMDetails/{userId}/{resourceGroup}/{vmName}", GetAzureVMDetails).Methods("GET")
	r.HandleFunc("/api/terraform/destroyAzureVM/{userId}", DestroyAzureVm).Methods("DELETE")
	r.HandleFunc("/api/terraform/destroyAzureDockerVM/{userId}", DestroyAzureDockerVm).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":28080", r))
}
