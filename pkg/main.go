package main


import (
    "fmt"
    "os"
    machineapi "github.com/openshift/cluster-api/pkg/apis/machine/v1beta1"
    awsprovider "sigs.k8s.io/cluster-api-provider-aws/pkg/apis/awsproviderconfig/v1beta1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    //"k8s.io/utils/pointer"
    //corev1 "k8s.io/api/core/v1"
    //"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/kubernetes"
    "flag"
    mapiclient "github.com/openshift/cluster-api/pkg/client/clientset_generated/clientset"
	//machinev1beta1client "github.com/openshift/cluster-api/pkg/client/clientset_generated/clientset/typed/machine/v1beta1"
)

func main() {
    var (
		kubeconfig string
	)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()
    // creates the connection
    // creates the connection
	kconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println(err)
        os.Exit(1)
	}

    // creates the clientset
	client, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		fmt.Println(err)
        os.Exit(1)
	}
    fmt.Printf("client: %v \n", client)

    machineClient, err := mapiclient.NewForConfig(kconfig)
	if err != nil {
        fmt.Println(err)
        os.Exit(1)
	}

    machineClientset := machineClient.MachineV1beta1().MachineSets("openshift-machine-api")
    machinesets1, err := machineClientset.List(metav1.ListOptions{})
	if err != nil {
		fmt.Println("failed to list machines: %v", err)
        os.Exit(1)
	}
    fmt.Printf("machinesets: %v \n", machinesets1)

    /*
    instanceType := "testing"
    volumeType := "type"
    volumeSize := 10
    iops := 2500
	azs := []string{"a", "b", "c"}
    amiID := "test"
    clusterID := "test"
    role := "role"
    region := "region1"
    userDataSecret := "worker-remote"
    poolName := "x"
    userTags := map[string]string{"a": "1"}
    tags, _ := tagsFromUserTags(clusterID, userTags)
    */
	var newMachineSets []*machineapi.MachineSet
	for _, ms := range machinesets1.Items {
        // skip machinesets labeled as remote
        _, exists := ms.ObjectMeta.Labels["remote"]
        if exists {
            continue
        }
        newMS := ms.DeepCopy()
		//replicas := int32(1)
        //newMS.Spec.Template.Spec.ProviderSpec.Value = ms.Spec.Template.Spec.ProviderSpec
/*
		provider := &awsprovider.AWSMachineProviderConfig{
    		TypeMeta: metav1.TypeMeta{
    			APIVersion: "awsproviderconfig.openshift.io/v1beta1",
    			Kind:       "AWSMachineProviderConfig",
    		},
    		InstanceType: instanceType,
    		BlockDevices: []awsprovider.BlockDeviceMappingSpec{
    			{
    				EBS: &awsprovider.EBSBlockDeviceSpec{
    					VolumeType: pointer.StringPtr(volumeType),
    					VolumeSize: pointer.Int64Ptr(int64(volumeSize)),
    					Iops:       pointer.Int64Ptr(int64(iops)),
    				},
    			},
    		},
    		AMI:                awsprovider.AWSResourceReference{ID: &amiID},
    		Tags:               tags,
    		IAMInstanceProfile: &awsprovider.AWSResourceReference{ID: pointer.StringPtr(fmt.Sprintf("%s-%s-profile", clusterID, role))},
    		UserDataSecret:     &corev1.LocalObjectReference{Name: userDataSecret},
    		CredentialsSecret:  &corev1.LocalObjectReference{Name: "aws-cloud-credentials"},
    		Subnet: awsprovider.AWSResourceReference{
    			Filters: []awsprovider.Filter{{
    				Name:   "tag:Name",
    				Values: []string{fmt.Sprintf("%s-private-%s", clusterID, az)},
    			}},
    		},
    		Placement: awsprovider.Placement{Region: region, AvailabilityZone: az},
    		SecurityGroups: []awsprovider.AWSResourceReference{{
    			Filters: []awsprovider.Filter{{
    				Name:   "tag:Name",
    				Values: []string{fmt.Sprintf("%s-%s-sg", clusterID, role)},
    			}},
    		}},
    	}
		name := fmt.Sprintf("%s-%s-%s", clusterID, poolName, az)
		mset := &machineapi.MachineSet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "machine.openshift.io/v1beta1",
				Kind:       "MachineSet",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "openshift-machine-api",
				Name:      name,
				Labels: map[string]string{
					"machine.openshift.io/cluster-api-cluster": clusterID,
				},
			},
			Spec: machineapi.MachineSetSpec{
				Replicas: &replicas,
				Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"machine.openshift.io/cluster-api-machineset": name,
						"machine.openshift.io/cluster-api-cluster":    clusterID,
					},
				},
				Template: machineapi.MachineTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"machine.openshift.io/cluster-api-machineset":   name,
							"machine.openshift.io/cluster-api-cluster":      clusterID,
							"machine.openshift.io/cluster-api-machine-role": role,
							"machine.openshift.io/cluster-api-machine-type": role,
						},
					},
					Spec: machineapi.MachineSpec{
						ProviderSpec: machineapi.ProviderSpec{
							Value: &runtime.RawExtension{Object: provider},
						},
						// we don't need to set Versions, because we control those via cluster operators.
					},
				},
			},
		}
    */
		newMachineSets = append(newMachineSets, newMS)
	}
    fmt.Println("machinesets: %v", newMachineSets[0])
}

func tagsFromUserTags(clusterID string, usertags map[string]string) ([]awsprovider.TagSpecification, error) {
	tags := []awsprovider.TagSpecification{
		{Name: fmt.Sprintf("kubernetes.io/cluster/%s", clusterID), Value: "owned"},
	}
	forbiddenTags := sets.NewString()
	for idx := range tags {
		forbiddenTags.Insert(tags[idx].Name)
	}
	for k, v := range usertags {
		if forbiddenTags.Has(k) {
			return nil, fmt.Errorf("user tags may not clobber %s", k)
		}
		tags = append(tags, awsprovider.TagSpecification{Name: k, Value: v})
	}
	return tags, nil
}
