package dao

import (
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cdedev/response"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	_ "golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"

	"net/http"
	"os"
	"path"
	"text/template"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"terraformCloud/models"
)

var connectionString = "root:root@tcp(10.100.100.64:3316)/terraform_cloud?parseTime=true"

//var provider string

func GetCredentials(data models.CloudInfo, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, "Check your json values")
	}
	if data.CloudType == "AWS" || data.CloudType == "aws" || data.CloudType == "Aws" {
		err = DemoGetRegion(data)
		if err != nil {
			logrus.Error("ResponseCode:", 2090, "Results:", err)
			response.RespondJSON(w, 2090, nil)
		} else {
			DatabaseConnect(data, w, r)
		}
	} else if data.CloudType == "Azure" || data.CloudType == "azure" || data.CloudType == "AZURE" {
		//err = GetLocationList(data)
		//if err != nil {
		//	logrus.Error("ResponseCode:", 2090, "Results:", err)
		//	response.RespondJSON(w, 2090, nil)
		//} else {
		//	DatabaseConnect(data, w, r)
		//}
		DatabaseConnect(data, w, r)
	} else if data.CloudType == "Gcp" || data.CloudType == "GCP" || data.CloudType == "gcp" {
		DatabaseConnect(data, w, r)
	}
	//err = DemoGetRegion(data)
	//if err != nil {
	//	logrus.Error("ResponseCode:", 2090, "Results:", err)
	//	response.RespondJSON(w, 2090, nil)
	//} else {
	//	DatabaseConnect(data, w, r)
	//}
}

func DatabaseConnect(data models.CloudInfo, w http.ResponseWriter, r *http.Request) {
	var err error
	Database, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		logrus.Error("ResponseCode:", 2091, "Results:", err)
		response.RespondJSON(w, 2091, nil)
	}
	err = Database.AutoMigrate(&models.CloudInfo{})
	if err != nil {
		logrus.Error("ResponseCode:", 2024, "Results:", err)
		response.RespondJSON(w, 2024, nil)
	}
	err = Database.Create(&data).Error
	if err != nil {
		logrus.Error("ResponseCode:", 2092, "Results:", err)
		response.RespondJSON(w, 2092, data.ID)
	} else {
		response.RespondJSON(w, 1000, nil)
	}

}

// Checking Aws Region for verifying aws credentials

func DemoGetRegion(input models.CloudInfo) error {
	mySession, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(input.AccessKey, input.SecretKey, ""),
	})
	if err != nil {
		fmt.Println("err::", err)
	}
	fmt.Println("mySession::", mySession.Config.Region)
	return err
}

// Checking azure credentials for verifying azure credentials
//
//func GetLocationList(data models.CloudInfo) error {
//	cred, err := azidentity.NewClientSecretCredential(data.TenantId, data.ClientId, data.ClientSecret, nil)
//	if err != nil {
//		logrus.Error("ResponseCode:", 2018, "Results:", err)
//	}
//	clientFactory, err := armresources.NewResourceGroupsClient(data.SubscriptionId, cred, nil)
//	if err != nil {
//		fmt.Println("error in connecting resource group client::", err)
//	}
//	_, err = clientFactory.List(nil)
//	if err != nil {
//		fmt.Println("error in getting location list::", err)
//	}
//	return err
//
//}

//Checking GCP credentials for verifying it

//Creating EC2 instance

func CreateEc2Instance(ec2 models.Ec2Instance, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	var data models.CloudInfo
	var ingress []models.AwsIncomingNetworkSecurity
	var egress []models.AwsOutgoingNetworkSecurity

	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)
	err = json.NewDecoder(r.Body).Decode(&ec2)
	if err != nil {
		logrus.Error("ResponseCode:", 2000, "Results:", err)
		response.RespondJSON(w, 2000, nil)
	}
	Database, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		logrus.Error("ResponseCode:", 2091, "Results:", err)
		response.RespondJSON(w, 2091, nil)
	}
	err = Database.Where("id = ?", ec2.ID).Find(&data).Error
	if err != nil {
		logrus.Error("ResponseCode:", 2020, "Results:", err)
		response.RespondJSON(w, 2020, ec2.ID)
	}
	tf1 := path.Join("templates", "ec2-template")
	fmt.Println("template file path: ", tf1)
	temp := template.Must(template.ParseFiles(tf1))

	for i := 0; i < len(ec2.AwsIncomingNetworkSecurity); i++ {
		securityValue := models.AwsIncomingNetworkSecurity{
			FromPort:  ec2.AwsIncomingNetworkSecurity[i].FromPort,
			ToPort:    ec2.AwsIncomingNetworkSecurity[i].ToPort,
			Protocol:  ec2.AwsIncomingNetworkSecurity[i].Protocol,
			CidrBlock: ec2.AwsIncomingNetworkSecurity[i].CidrBlock,
		}
		ingress = append(ingress, securityValue)
	}

	for j := 0; j < len(ec2.AwsOutgoingNetworkSecurity); j++ {
		securityValue := models.AwsOutgoingNetworkSecurity{
			FromPort:  ec2.AwsOutgoingNetworkSecurity[j].FromPort,
			ToPort:    ec2.AwsOutgoingNetworkSecurity[j].ToPort,
			CidrBlock: ec2.AwsOutgoingNetworkSecurity[j].CidrBlock,
		}
		egress = append(egress, securityValue)
	}

	ec2Values := models.Ec2Instance{
		AccessKey: data.AccessKey,
		SecretKey: data.SecretKey,
		Region:    ec2.Region,
		AwsInstance: models.AwsInstance{
			ResourceName:      ec2.AwsInstance.ResourceName,
			InstanceType:      ec2.AwsInstance.InstanceType,
			InstanceImage:     ec2.AwsInstance.InstanceImage,
			InstanceFirstTag:  ec2.AwsInstance.InstanceFirstTag,
			InstanceSecondTag: ec2.AwsInstance.InstanceSecondTag,
		},
		AwsVpc: models.AwsVpc{
			VpcResourceName: ec2.AwsVpc.VpcResourceName,
			VpcCidrBlock:    ec2.AwsVpc.VpcCidrBlock,
			VpcFirstTag:     ec2.AwsVpc.VpcFirstTag,
			VpcSecondTag:    ec2.AwsVpc.VpcSecondTag,
		},
		AwsSubnet: models.AwsSubnet{
			SubnetResourceName:     ec2.AwsSubnet.SubnetResourceName,
			SubnetCidrBlock:        ec2.AwsSubnet.SubnetCidrBlock,
			SubnetAvailabilityZone: ec2.AwsSubnet.SubnetAvailabilityZone,
			SubnetFirstTag:         ec2.AwsSubnet.SubnetFirstTag,
			SubnetSecondTag:        ec2.AwsSubnet.SubnetSecondTag,
		},
		AwsSecurityGroupName:         ec2.AwsSecurityGroupName,
		AwsSecurityGroupResourceName: ec2.AwsSecurityGroupResourceName,
		AwsIncomingNetworkSecurity:   ingress,
		AwsOutgoingNetworkSecurity:   egress,
	}
	fmt.Println("Checking files: ", ec2Values)
	err = os.MkdirAll("./aws/"+ec2.ID+"/ec2", 0755)
	if err != nil {
		fmt.Println("Cannot create directory", err)
	}
	createFile, err := os.Create("./aws/" + ec2.ID + "/ec2/ec2-instance.tf")
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = temp.Execute(createFile, ec2Values)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, nil)
	}

	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.5.5")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}

	workingDir := "./aws/" + ec2.ID + "/ec2"
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}

	err = tf.Init(context.TODO(), tfexec.Upgrade(true))
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}

	err = tf.Apply(context.TODO())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	} else {
		output, err := tf.Output(context.TODO())
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, err)
		} else {
			logrus.Error("ResponseCode:", 1000, "Results:", nil)
			response.RespondJSON(w, 1000, output)
		}
	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)

}

func CreateEc2InstanceWithScriptFile(ec2 models.Ec2Instance, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	var data models.CloudInfo
	var ingress []models.AwsIncomingNetworkSecurity
	var egress []models.AwsOutgoingNetworkSecurity

	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)

	jsonBytes := []byte(r.FormValue("data"))

	if err = json.Unmarshal(jsonBytes, &ec2); err != nil {
		logrus.Error("ResponseCode:", 2000, "Results:", err)
		response.RespondJSON(w, 2000, nil)
	}

	Database, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		logrus.Error("ResponseCode:", 2091, "Results:", err)
		response.RespondJSON(w, 2091, nil)
	}
	err = Database.Where("id = ?", ec2.ID).Find(&data).Error
	if err != nil {
		logrus.Error("ResponseCode:", 2020, "Results:", err)
		response.RespondJSON(w, 2020, ec2.ID)
	}
	err = os.MkdirAll("./aws/"+ec2.ID+"/ec2/dockerTf/dockerData", 0755)
	if err != nil {
		fmt.Println("Cannot create directory", err)
	}

	//Generating ssh files

	//privateKeyPath := "./aws/" + ec2.ID + "/ec2/dockerTf/id_rsa"
	//
	//_, err = exec.Command("/bin/bash", "-c", "ssh-keygen -q -t rsa -N '' -f"+privateKeyPath+"<<<y >/dev/null 2>&1").Output()
	//if err != nil {
	//	logrus.Error("ResponseCode:", 2018, "Results:", err)
	//	response.RespondJSON(w, 2018, nil)
	//}

	//Creating template files

	tf1 := path.Join("templates", "ec2-script-template")
	fmt.Println("template file path: ", tf1)
	temp := template.Must(template.ParseFiles(tf1))

	for i := 0; i < len(ec2.AwsIncomingNetworkSecurity); i++ {
		securityValue := models.AwsIncomingNetworkSecurity{
			FromPort:  ec2.AwsIncomingNetworkSecurity[i].FromPort,
			ToPort:    ec2.AwsIncomingNetworkSecurity[i].ToPort,
			Protocol:  ec2.AwsIncomingNetworkSecurity[i].Protocol,
			CidrBlock: ec2.AwsIncomingNetworkSecurity[i].CidrBlock,
		}
		ingress = append(ingress, securityValue)
	}

	for j := 0; j < len(ec2.AwsOutgoingNetworkSecurity); j++ {
		securityValue := models.AwsOutgoingNetworkSecurity{
			FromPort:  ec2.AwsOutgoingNetworkSecurity[j].FromPort,
			ToPort:    ec2.AwsOutgoingNetworkSecurity[j].ToPort,
			CidrBlock: ec2.AwsOutgoingNetworkSecurity[j].CidrBlock,
		}
		egress = append(egress, securityValue)
	}

	ec2Values := models.Ec2Instance{
		AccessKey: data.AccessKey,
		SecretKey: data.SecretKey,
		Region:    ec2.Region,
		AwsInstance: models.AwsInstance{
			ResourceName:      ec2.AwsInstance.ResourceName,
			InstanceType:      ec2.AwsInstance.InstanceType,
			InstanceImage:     ec2.AwsInstance.InstanceImage,
			InstanceFirstTag:  ec2.AwsInstance.InstanceFirstTag,
			InstanceSecondTag: ec2.AwsInstance.InstanceSecondTag,
		},
		AwsVpc: models.AwsVpc{
			VpcResourceName: ec2.AwsVpc.VpcResourceName,
			VpcCidrBlock:    ec2.AwsVpc.VpcCidrBlock,
			VpcFirstTag:     ec2.AwsVpc.VpcFirstTag,
			VpcSecondTag:    ec2.AwsVpc.VpcSecondTag,
		},
		AwsSubnet: models.AwsSubnet{
			SubnetResourceName:     ec2.AwsSubnet.SubnetResourceName,
			SubnetCidrBlock:        ec2.AwsSubnet.SubnetCidrBlock,
			SubnetAvailabilityZone: ec2.AwsSubnet.SubnetAvailabilityZone,
			SubnetFirstTag:         ec2.AwsSubnet.SubnetFirstTag,
			SubnetSecondTag:        ec2.AwsSubnet.SubnetSecondTag,
		},
		AwsSecurityGroupName:         ec2.AwsSecurityGroupName,
		AwsInternetGatewayName:       ec2.AwsInternetGatewayName,
		AwsSecurityGroupResourceName: ec2.AwsSecurityGroupResourceName,
		AwsKeyId:                     ec2.AwsKeyId,
		AwsKeyName:                   ec2.AwsKeyName,
		Hostname:                     ec2.Hostname,
		AwsIncomingNetworkSecurity:   ingress,
		AwsOutgoingNetworkSecurity:   egress,
	}
	fmt.Println("Checking files: ", ec2Values)

	createFile, err := os.Create("./aws/" + ec2.ID + "/ec2/dockerTf/ec2-instance.tf")
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = temp.Execute(createFile, ec2Values)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, nil)
	}

	tf2 := path.Join("AwsEc2", "install")
	fmt.Println("script file path: ", tf2)
	temp1 := template.Must(template.ParseFiles(tf2))
	scriptValues := models.Ec2Instance{Hostname: ec2.Hostname}
	createScriptFile, err := os.Create("./aws/" + ec2.ID + "/ec2/dockerTf/docker-install.sh")
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = temp1.Execute(createScriptFile, scriptValues)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, nil)
	}

	if ec2.UploadFile == true {

		err = r.ParseMultipartForm(20000) // 10 MB limit
		if err != nil {
			fmt.Println("Error", err)
		}

		// Get files from request body
		files := r.MultipartForm.File["files"]
		if len(files) == 0 {
			fmt.Println("Error", err)
		}

		// Parse each file and return filename and contents
		for _, test := range files {
			// Open file from request body
			f, err := test.Open()
			if err != nil {
				fmt.Println("Error", err)

			}
			// Read file contents
			contents, err := ioutil.ReadAll(f)
			if err != nil {
				fmt.Println("Error", err)
			}

			createF, err := os.Create("./AwsEc2/" + test.Filename)
			if err != nil {
				fmt.Println("Error", err)
			}

			// Write the contents of the request body to the file
			_, err = createF.Write(contents)
			if err != nil {
				fmt.Println("Error", err)
			}

			tf := path.Join("AwsEc2", test.Filename)
			fmt.Println("script file path: ", tf)
			temp3 := template.Must(template.ParseFiles(tf))
			createDynamicFile, err := os.Create("./aws/" + ec2.ID + "/ec2/dockerTf/dockerData/" + test.Filename)
			if err != nil {
				log.Println("create file: ", err)
			}
			err = temp3.Execute(createDynamicFile, nil)
			if err != nil {
				logrus.Error("ResponseCode:", 2018, "Results:", err)
				response.RespondJSON(w, 2018, nil)
			}

		}
	}

	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.5.5")),
	}

	execPath, err := installer.Install(context.TODO())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}

	workingDir := "./aws/" + ec2.ID + "/ec2/dockerTf"
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}

	err = tf.Init(context.TODO(), tfexec.Upgrade(true))
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}

	err = tf.Apply(context.TODO())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	} else {
		output, err := tf.Output(context.TODO())
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, err)
		} else {
			logrus.Error("ResponseCode:", 1000, "Results:", nil)
			response.RespondJSON(w, 1000, output)
		}
	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)

}

func ListEc2Instances(id string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)
	var err error
	var data models.CloudInfo
	Database, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		logrus.Error("ResponseCode:", 2091, "Results:", err)
		response.RespondJSON(w, 2091, nil)
	}
	err = Database.Where("id = ?", id).Find(&data).Error
	if err != nil {
		logrus.Error("ResponseCode:", 2020, "Results:", err)
		response.RespondJSON(w, 2020, id)
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(data.AccessKey, data.SecretKey, ""),
		Region:      aws.String(data.Region),
	})
	svc := ec2.New(sess)
	result, err := svc.DescribeInstances(nil)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	} else {
		logrus.Error("ResponseCode:", 1000, "Results:", nil)
		response.RespondJSON(w, 1000, result)
	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)
}

func GetEc2InstanceDetail(id string, instance_id string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)
	var err error
	var data models.CloudInfo
	Database, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		logrus.Error("ResponseCode:", 2091, "Results:", err)
		response.RespondJSON(w, 2091, nil)
	}
	err = Database.Where("id = ?", id).Find(&data).Error
	if err != nil {
		logrus.Error("ResponseCode:", 2020, "Results:", err)
		response.RespondJSON(w, 2020, id)
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(data.AccessKey, data.SecretKey, ""),
		Region:      aws.String(data.Region),
	})
	svc := ec2.New(sess)
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instance_id),
		},
	}
	result, err := svc.DescribeInstances(params)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	} else {
		logrus.Error("ResponseCode:", 1000, "Results:", nil)
		response.RespondJSON(w, 1000, result)
	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)
}

func DestroyInstance(id string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)
	var err error
	workingDir := "./aws/" + id + "/ec2"
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.5.5")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}
	err = tf.Destroy(context.TODO())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	} else {
		logrus.Error("ResponseCode:", 1000, "Results:", nil)
		response.RespondJSON(w, 1000, nil)
	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)
}

func DestroyDockerInstance(id string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)
	var err error
	workingDir := "./aws/" + id + "/ec2/dockerTf"
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.5.5")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}
	err = tf.Destroy(context.TODO())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	} else {
		logrus.Error("ResponseCode:", 1000, "Results:", nil)
		response.RespondJSON(w, 1000, nil)
	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)
}

//Azure vm creation

// Listing all available Linux images inside your

func CreateAzureVm(vm models.AzureVM, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var err error
	var data models.CloudInfo
	var security []models.AzureNetworkSecurity

	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)
	err = json.NewDecoder(r.Body).Decode(&vm)
	if err != nil {
		logrus.Error("ResponseCode:", 2000, "Results:", err)
		response.RespondJSON(w, 2000, nil)
	}
	Database, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		logrus.Error("ResponseCode:", 2091, "Results:", err)
		response.RespondJSON(w, 2091, nil)
	}
	err = Database.Where("id = ?", vm.ID).Find(&data).Error
	if err != nil {
		logrus.Error("ResponseCode:", 2020, "Results:", err)
		response.RespondJSON(w, 2020, vm.ID)
	}

	if vm.VMType == "Linux" || vm.VMType == "linux" || vm.VMType == "LINUX" {
		tf1 := path.Join("templates", "azure-linux-vm-template")
		fmt.Println("template file path: ", tf1)
		temp := template.Must(template.ParseFiles(tf1))

		for i := 0; i < len(vm.AzureNetworkSecurity); i++ {
			securityValue := models.AzureNetworkSecurity{
				SecurityName:      vm.AzureNetworkSecurity[i].SecurityName,
				SecurityDirection: vm.AzureNetworkSecurity[i].SecurityDirection,
				SecurityAccess:    vm.AzureNetworkSecurity[i].SecurityAccess,
				SecurityPriority:  vm.AzureNetworkSecurity[i].SecurityPriority,
				Protocol:          vm.AzureNetworkSecurity[i].Protocol,
				DestinationPort:   vm.AzureNetworkSecurity[i].DestinationPort,
			}
			security = append(security, securityValue)
		}

		vmValues := models.AzureVM{
			SubscriptionId: data.SubscriptionId,
			ClientId:       data.ClientId,
			ClientSecret:   data.ClientSecret,
			TenantId:       data.TenantId,
			AzureResourceGroup: models.AzureResourceGroup{
				ResourceName:  vm.AzureResourceGroup.ResourceName,
				ResourceGroup: vm.AzureResourceGroup.ResourceGroup,
				Location:      vm.AzureResourceGroup.Location,
			},
			AzureVirtualNetwork: models.AzureVirtualNetwork{
				ResourceName:               vm.AzureVirtualNetwork.ResourceName,
				VirtualNetworkName:         vm.AzureVirtualNetwork.VirtualNetworkName,
				VirtualNetworkAddressSpace: vm.AzureVirtualNetwork.VirtualNetworkAddressSpace,
			},
			AzureSubnet: models.AzureSubnet{
				ResourceName:  vm.AzureSubnet.ResourceName,
				SubnetName:    vm.AzureSubnet.SubnetName,
				SubnetAddress: vm.AzureSubnet.SubnetAddress,
			},
			AzurePublicIp: models.AzurePublicIp{
				ResourceName:             vm.AzurePublicIp.ResourceName,
				PublicIpName:             vm.AzurePublicIp.PublicIpName,
				PublicIpAllocationMethod: vm.AzurePublicIp.PublicIpAllocationMethod,
			},
			AzureNetworkSecurityGroup: models.AzureNetworkSecurityGroup{
				ResourceName:             vm.AzureNetworkSecurityGroup.ResourceName,
				NetworkSecurityGroupName: vm.AzureNetworkSecurityGroup.NetworkSecurityGroupName,
			},
			AzureNetworkInterface: models.AzureNetworkInterface{
				ResourceName:              vm.AzureNetworkInterface.ResourceName,
				NetworkInterfaceName:      vm.AzureNetworkInterface.NetworkInterfaceName,
				SecurityGroupResourceName: vm.AzureNetworkInterface.SecurityGroupResourceName,
			},
			AzureLinuxVirtualMachine: models.AzureLinuxVirtualMachine{
				ResourceName:  vm.AzureLinuxVirtualMachine.ResourceName,
				VmName:        vm.AzureLinuxVirtualMachine.VmName,
				VmPublisher:   vm.AzureLinuxVirtualMachine.VmPublisher,
				VmOffer:       vm.AzureLinuxVirtualMachine.VmOffer,
				VmSku:         vm.AzureLinuxVirtualMachine.VmSku,
				VmVersion:     vm.AzureLinuxVirtualMachine.VmVersion,
				VmSize:        vm.AzureLinuxVirtualMachine.VmSize,
				Hostname:      vm.AzureLinuxVirtualMachine.Hostname,
				AdminUser:     vm.AzureLinuxVirtualMachine.AdminUser,
				AdminPassword: vm.AzureLinuxVirtualMachine.AdminPassword,
			},
			AzureNetworkSecurity: security,
		}
		fmt.Println("Checking files: ", vmValues)
		err = os.MkdirAll("./azure/"+vm.ID+"/vm/ltf", 0755)
		if err != nil {
			fmt.Println("Cannot create directory", err)
		}
		createFile, err := os.Create("./azure/" + vm.ID + "/vm/ltf/azure-linux-vm.tf")
		if err != nil {
			log.Println("create file: ", err)
			return
		}
		err = temp.Execute(createFile, vmValues)
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, nil)
		}

		installer := &releases.ExactVersion{
			Product: product.Terraform,
			Version: version.Must(version.NewVersion("1.5.5")),
		}

		execPath, err := installer.Install(context.Background())
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, err)
		}

		workingDir := "./azure/" + vm.ID + "/vm/ltf"
		tf, err := tfexec.NewTerraform(workingDir, execPath)
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, err)
		}

		err = tf.Init(context.TODO(), tfexec.Upgrade(true))
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, err)
		}

		err = tf.Apply(context.TODO())
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, err)
		} else {
			output, err := tf.Output(context.TODO())
			if err != nil {
				logrus.Error("ResponseCode:", 2018, "Results:", err)
				response.RespondJSON(w, 2018, err)
			} else {
				logrus.Error("ResponseCode:", 1000, "Results:", nil)
				response.RespondJSON(w, 1000, output)
			}
		}

	} else if vm.VMType == "Windows" || vm.VMType == "windows" || vm.VMType == "WINDOWS" {
		tf1 := path.Join("templates", "azure-windows-vm-template")
		fmt.Println("template file path: ", tf1)
		temp := template.Must(template.ParseFiles(tf1))

		for i := 0; i < len(vm.AzureNetworkSecurity); i++ {
			securityValue := models.AzureNetworkSecurity{
				SecurityName:      vm.AzureNetworkSecurity[i].SecurityName,
				SecurityDirection: vm.AzureNetworkSecurity[i].SecurityDirection,
				SecurityAccess:    vm.AzureNetworkSecurity[i].SecurityAccess,
				SecurityPriority:  vm.AzureNetworkSecurity[i].SecurityPriority,
				Protocol:          vm.AzureNetworkSecurity[i].Protocol,
				DestinationPort:   vm.AzureNetworkSecurity[i].DestinationPort,
			}
			security = append(security, securityValue)
		}

		vmValues := models.AzureVM{
			SubscriptionId: data.SubscriptionId,
			ClientId:       data.ClientId,
			ClientSecret:   data.ClientSecret,
			TenantId:       data.TenantId,
			AzureResourceGroup: models.AzureResourceGroup{
				ResourceName:  vm.AzureResourceGroup.ResourceName,
				ResourceGroup: vm.AzureResourceGroup.ResourceGroup,
				Location:      vm.AzureResourceGroup.Location,
			},
			AzureVirtualNetwork: models.AzureVirtualNetwork{
				ResourceName:               vm.AzureVirtualNetwork.ResourceName,
				VirtualNetworkName:         vm.AzureVirtualNetwork.VirtualNetworkName,
				VirtualNetworkAddressSpace: vm.AzureVirtualNetwork.VirtualNetworkAddressSpace,
			},
			AzureSubnet: models.AzureSubnet{
				ResourceName:  vm.AzureSubnet.ResourceName,
				SubnetName:    vm.AzureSubnet.SubnetName,
				SubnetAddress: vm.AzureSubnet.SubnetAddress,
			},
			AzurePublicIp: models.AzurePublicIp{
				ResourceName:             vm.AzurePublicIp.ResourceName,
				PublicIpName:             vm.AzurePublicIp.PublicIpName,
				PublicIpAllocationMethod: vm.AzurePublicIp.PublicIpAllocationMethod,
			},
			AzureNetworkSecurityGroup: models.AzureNetworkSecurityGroup{
				ResourceName:             vm.AzureNetworkSecurityGroup.ResourceName,
				NetworkSecurityGroupName: vm.AzureNetworkSecurityGroup.NetworkSecurityGroupName,
			},
			AzureNetworkInterface: models.AzureNetworkInterface{
				ResourceName:              vm.AzureNetworkInterface.ResourceName,
				NetworkInterfaceName:      vm.AzureNetworkInterface.NetworkInterfaceName,
				SecurityGroupResourceName: vm.AzureNetworkInterface.SecurityGroupResourceName,
			},
			AzureLinuxVirtualMachine: models.AzureLinuxVirtualMachine{
				ResourceName:  vm.AzureLinuxVirtualMachine.ResourceName,
				VmName:        vm.AzureLinuxVirtualMachine.VmName,
				VmPublisher:   vm.AzureLinuxVirtualMachine.VmPublisher,
				VmOffer:       vm.AzureLinuxVirtualMachine.VmOffer,
				VmSku:         vm.AzureLinuxVirtualMachine.VmSku,
				VmVersion:     vm.AzureLinuxVirtualMachine.VmVersion,
				VmSize:        vm.AzureLinuxVirtualMachine.VmSize,
				Hostname:      vm.AzureLinuxVirtualMachine.Hostname,
				AdminUser:     vm.AzureLinuxVirtualMachine.AdminUser,
				AdminPassword: vm.AzureLinuxVirtualMachine.AdminPassword,
			},
			AzureNetworkSecurity: security,
		}

		fmt.Println("Checking files: ", vmValues)
		err = os.MkdirAll("./azure/"+vm.ID+"/vm/wtf", 0755)
		if err != nil {
			fmt.Println("Cannot create directory", err)
		}
		createFile, err := os.Create("./azure/" + vm.ID + "/vm/wtf/azure-windows-vm.tf")
		if err != nil {
			log.Println("create file: ", err)
			return
		}
		err = temp.Execute(createFile, vmValues)
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, nil)
		}

		installer := &releases.ExactVersion{
			Product: product.Terraform,
			Version: version.Must(version.NewVersion("1.5.5")),
		}

		execPath, err := installer.Install(context.Background())
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, err)
		}

		workingDir := "./azure/" + vm.ID + "/vm/wtf"
		tf, err := tfexec.NewTerraform(workingDir, execPath)
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, err)
		}

		err = tf.Init(context.TODO(), tfexec.Upgrade(true))
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, err)
		}

		err = tf.Apply(context.TODO())
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, err)
		} else {
			output, err := tf.Output(context.TODO())
			if err != nil {
				logrus.Error("ResponseCode:", 2018, "Results:", err)
				response.RespondJSON(w, 2018, err)
			} else {
				logrus.Error("ResponseCode:", 1000, "Results:", nil)
				response.RespondJSON(w, 1000, output)
			}
		}
	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)
}

func CreateAzureVmWithDocker(vm models.AzureVM, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var err error
	var data models.CloudInfo
	var security []models.AzureNetworkSecurity

	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)

	jsonBytes := []byte(r.FormValue("data"))

	if err = json.Unmarshal(jsonBytes, &vm); err != nil {
		logrus.Error("ResponseCode:", 2000, "Results:", err)
		response.RespondJSON(w, 2000, nil)
	}

	Database, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		logrus.Error("ResponseCode:", 2091, "Results:", err)
		response.RespondJSON(w, 2091, nil)
	}
	err = Database.Where("id = ?", vm.ID).Find(&data).Error
	if err != nil {
		logrus.Error("ResponseCode:", 2020, "Results:", err)
		response.RespondJSON(w, 2020, vm.ID)
	}

	err = os.MkdirAll("./azure/"+vm.ID+"/vm/dockerTf/dockerData", 0755)
	if err != nil {
		fmt.Println("Cannot create directory", err)
	}

	tf1 := path.Join("templates", "azure-vm-script-template")
	fmt.Println("template file path: ", tf1)
	temp := template.Must(template.ParseFiles(tf1))

	for i := 0; i < len(vm.AzureNetworkSecurity); i++ {
		securityValue := models.AzureNetworkSecurity{
			SecurityName:      vm.AzureNetworkSecurity[i].SecurityName,
			SecurityDirection: vm.AzureNetworkSecurity[i].SecurityDirection,
			SecurityAccess:    vm.AzureNetworkSecurity[i].SecurityAccess,
			SecurityPriority:  vm.AzureNetworkSecurity[i].SecurityPriority,
			Protocol:          vm.AzureNetworkSecurity[i].Protocol,
			DestinationPort:   vm.AzureNetworkSecurity[i].DestinationPort,
		}
		security = append(security, securityValue)
	}

	vmValues := models.AzureVM{
		SubscriptionId:             data.SubscriptionId,
		ClientId:                   data.ClientId,
		ClientSecret:               data.ClientSecret,
		TenantId:                   data.TenantId,
		Location:                   data.Region,
		ResourceGroup:              vm.ResourceGroup,
		VirtualNetworkName:         vm.VirtualNetworkName,
		VirtualNetworkAddressSpace: vm.VirtualNetworkAddressSpace,
		SubnetName:                 vm.SubnetName,
		SubnetAddress:              vm.SubnetAddress,
		PublicIpName:               vm.PublicIpName,
		NetworkSecurityGroupName:   vm.NetworkSecurityGroupName,
		AzureNetworkSecurity:       security,
		NetworkInterfaceName:       vm.NetworkInterfaceName,
		VMName:                     vm.VMName,
		VmPublisher:                vm.VmPublisher,
		VmOffer:                    vm.VmOffer,
		VmSku:                      vm.VmSku,
		VmVersion:                  vm.VmVersion,
		Hostname:                   vm.Hostname,
		AdminUser:                  vm.AdminUser,
		AdminPassword:              vm.AdminPassword,
	}
	fmt.Println("Checking files: ", vmValues)

	createFile, err := os.Create("./azure/" + vm.ID + "/vm/dockerTf/azure-docker-vm.tf")
	if err != nil {
		log.Println("create file: ", err)
		return
	}

	err = temp.Execute(createFile, vmValues)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, nil)
	}

	tf2 := path.Join("AzureVm", "install")
	fmt.Println("script file path: ", tf2)
	temp1 := template.Must(template.ParseFiles(tf2))

	scriptV := models.AzureVM{
		AdminUser: vm.AdminUser,
	}
	createScriptFile, err := os.Create("./azure/" + vm.ID + "/vm/dockerTf/install.sh")
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = temp1.Execute(createScriptFile, scriptV)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, nil)
	}

	if vm.UploadFile == "true" {

		err = r.ParseMultipartForm(20000) // 10 MB limit
		if err != nil {
			fmt.Println("Error", err)
		}

		// Get files from request body
		files := r.MultipartForm.File["files"]
		if len(files) == 0 {
			fmt.Println("Error", err)
		}

		// Parse each file and return filename and contents
		for _, test := range files {
			// Open file from request body
			f, err := test.Open()
			if err != nil {
				fmt.Println("Error", err)

			}
			// Read file contents
			contents, err := ioutil.ReadAll(f)
			if err != nil {
				fmt.Println("Error", err)
			}

			createF, err := os.Create("./AzureVm/" + test.Filename)
			if err != nil {
				fmt.Println("Error", err)
			}

			// Write the contents of the request body to the file
			_, err = createF.Write(contents)
			if err != nil {
				fmt.Println("Error", err)
			}

			tf := path.Join("AzureVm", test.Filename)
			fmt.Println("script file path: ", tf)
			temp3 := template.Must(template.ParseFiles(tf))
			createDynamicFile, err := os.Create("./azure/" + vm.ID + "/vm/dockerTf/dockerData/" + test.Filename)
			if err != nil {
				log.Println("create file: ", err)
			}
			err = temp3.Execute(createDynamicFile, nil)
			if err != nil {
				logrus.Error("ResponseCode:", 2018, "Results:", err)
				response.RespondJSON(w, 2018, nil)
			}

		}
	}

	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.5.5")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}

	workingDir := "./azure/" + vm.ID + "/vm/dockerTf"
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}

	err = tf.Init(context.TODO(), tfexec.Upgrade(true))
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}

	err = tf.Apply(context.TODO())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	} else {
		output, err := tf.Output(context.TODO())
		if err != nil {
			logrus.Error("ResponseCode:", 2018, "Results:", err)
			response.RespondJSON(w, 2018, err)
		} else {
			logrus.Error("ResponseCode:", 1000, "Results:", nil)
			response.RespondJSON(w, 1000, output)
		}
	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)
}

//Destroy Azure VM

func DestroyAzureVM(id string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)
	var err error
	workingDir := "./azure/" + id + "/vm/tf"
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.5.5")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}
	err = tf.Destroy(context.TODO())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	} else {
		logrus.Error("ResponseCode:", 1000, "Results:", nil)
		response.RespondJSON(w, 1000, nil)
	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)
}

func DestroyAzureDockerVM(id string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)
	var err error
	workingDir := "./azure/" + id + "/vm/dockerTf"
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion("1.5.5")),
	}

	execPath, err := installer.Install(context.Background())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	}
	err = tf.Destroy(context.TODO())
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, err)
	} else {
		logrus.Error("ResponseCode:", 1000, "Results:", nil)
		response.RespondJSON(w, 1000, nil)
	}
	err = os.RemoveAll(workingDir)
	if err != nil {
		fmt.Println("Error while deleting files: "+workingDir, err)
	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)
}

// Listing Created VM's inside Azure

func ListAllAzureVM(id string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)
	var err error
	var data models.CloudInfo
	Database, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		logrus.Error("ResponseCode:", 2091, "Results:", err)
		response.RespondJSON(w, 2091, nil)
	}
	err = Database.Where("id = ?", id).Find(&data).Error
	if err != nil {
		logrus.Error("ResponseCode:", 2020, "Results:", err)
		response.RespondJSON(w, 2020, id)
	}

	cred, err := azidentity.NewClientSecretCredential(data.TenantId, data.ClientId, data.ClientSecret, nil)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, nil)
	}
	clientFactory, err := armcompute.NewClientFactory(data.SubscriptionId, cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	pager := clientFactory.NewVirtualMachinesClient().NewListAllPager(&armcompute.VirtualMachinesClientListAllOptions{StatusOnly: nil})
	for pager.More() {
		page, err := pager.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}
		for _, v := range page.Value {
			//page = v
			response.RespondJSON(w, 1000, v)
		}

	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)
}

func GetAzureVMDetails(id string, resourcegroup string, vmname string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	startTime := time.Now()
	fmt.Println("#########*********Api Calling time*********#########", startTime)
	var err error
	var data models.CloudInfo
	Database, err := gorm.Open(mysql.Open(connectionString), &gorm.Config{})
	if err != nil {
		logrus.Error("ResponseCode:", 2091, "Results:", err)
		response.RespondJSON(w, 2091, nil)
	}
	err = Database.Where("id = ?", id).Find(&data).Error
	if err != nil {
		logrus.Error("ResponseCode:", 2020, "Results:", err)
		response.RespondJSON(w, 2020, id)
	}

	cred, err := azidentity.NewClientSecretCredential(data.TenantId, data.ClientId, data.ClientSecret, nil)
	if err != nil {
		logrus.Error("ResponseCode:", 2018, "Results:", err)
		response.RespondJSON(w, 2018, nil)
	}
	clientFactory, err := armcompute.NewClientFactory(data.SubscriptionId, cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	res, err := clientFactory.NewVirtualMachinesClient().Get(context.TODO(), resourcegroup, vmname, &armcompute.VirtualMachinesClientGetOptions{Expand: nil})
	if err != nil {
		log.Fatalf("failed to finish the request: %v", err)
	} else {
		response.RespondJSON(w, 1000, res)
	}
	endTime := time.Now().Sub(startTime)
	fmt.Println("#############*************Time taken to complete api request**********#############", endTime)
}

// Creating GCP vm
