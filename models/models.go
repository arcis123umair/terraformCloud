package models

type CloudInfo struct {
	ID             string `gorm:"primary_key;not null;unique"`
	UserID         string
	AccessKey      string
	SecretKey      string
	Region         string
	CloudType      string
	TenantId       string
	SubscriptionId string
	ClientId       string
	ClientSecret   string
}

type Ec2Instance struct {
	ID                           string
	AccessKey                    string
	SecretKey                    string
	Region                       string
	AwsInstance                  AwsInstance
	AwsVpc                       AwsVpc
	AwsInternetGatewayName       string
	AwsSubnet                    AwsSubnet
	AwsKeyId                     string
	AwsKeyName                   string
	AwsSecurityGroupResourceName string
	AwsSecurityGroupName         string
	Hostname                     string
	UploadFile                   bool
	AwsIncomingNetworkSecurity   []AwsIncomingNetworkSecurity
	AwsOutgoingNetworkSecurity   []AwsOutgoingNetworkSecurity
}

type AwsInstance struct {
	ResourceName      string
	InstanceType      string
	InstanceImage     string
	InstanceFirstTag  string
	InstanceSecondTag string
}

type AwsVpc struct {
	VpcResourceName string
	VpcCidrBlock    string
	VpcFirstTag     string
	VpcSecondTag    string
}

type AwsSubnet struct {
	SubnetResourceName     string
	SubnetCidrBlock        string
	SubnetAvailabilityZone string
	SubnetFirstTag         string
	SubnetSecondTag        string
}

type AwsIncomingNetworkSecurity struct {
	FromPort  int64
	ToPort    int64
	Protocol  string
	CidrBlock string
}

type AwsOutgoingNetworkSecurity struct {
	FromPort  int64
	ToPort    int64
	CidrBlock string
}

//////////////////// Azure /////////////////////////////////

type AzureResourceGroup struct {
	ResourceName  string
	ResourceGroup string
	Location      string
}

type AzureVirtualNetwork struct {
	ResourceName               string
	VirtualNetworkName         string
	VirtualNetworkAddressSpace string
}

type AzureSubnet struct {
	ResourceName  string
	SubnetName    string
	SubnetAddress string
}

type AzurePublicIp struct {
	ResourceName             string
	PublicIpName             string
	PublicIpAllocationMethod string
}

type AzureNetworkSecurityGroup struct {
	ResourceName             string
	NetworkSecurityGroupName string
}

type AzureNetworkInterface struct {
	ResourceName              string
	NetworkInterfaceName      string
	SecurityGroupResourceName string
}

type AzureLinuxVirtualMachine struct {
	ResourceName  string
	VmName        string
	VmSize        string
	VmPublisher   string
	VmOffer       string
	VmSku         string
	VmVersion     string
	Hostname      string
	AdminUser     string
	AdminPassword string
}

type AzureVM struct {
	ID                        string
	VMType                    string
	ClientId                  string
	ClientSecret              string
	TenantId                  string
	SubscriptionId            string
	AzureResourceGroup        AzureResourceGroup
	AzureVirtualNetwork       AzureVirtualNetwork
	AzureSubnet               AzureSubnet
	AzurePublicIp             AzurePublicIp
	AzureNetworkSecurityGroup AzureNetworkSecurityGroup
	AzureNetworkSecurity      []AzureNetworkSecurity
	AzureNetworkInterface     AzureNetworkInterface
	AzureLinuxVirtualMachine  AzureLinuxVirtualMachine
	UploadFile                string
}

type AzureNetworkSecurity struct {
	SecurityName      string
	SecurityPriority  int32
	SecurityDirection string
	SecurityAccess    string
	Protocol          string
	DestinationPort   int64
}
