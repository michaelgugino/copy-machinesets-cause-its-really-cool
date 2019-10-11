package main

import (
	"fmt"
	machineapi "github.com/openshift/cluster-api/pkg/apis/machine/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	awsprovider "sigs.k8s.io/cluster-api-provider-aws/pkg/apis/awsproviderconfig/v1beta1"
	//"k8s.io/utils/pointer"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/kubernetes"
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
	kconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// creates the clientset
	/*
		client, err := kubernetes.NewForConfig(kconfig)
		if err != nil {
			fmt.Println(err)
	        os.Exit(1)
		}
	    //fmt.Printf("client: %v \n", client)
	*/

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
	//fmt.Printf("machinesets: %v \n", machinesets1)

	codec, err := awsprovider.NewCodec()
	if err != nil {
		fmt.Println("failed to build codec: %v", err)
		os.Exit(1)
	}

	var newMachineSets []*machineapi.MachineSet
	msMap := make(map[string]*machineapi.MachineSet)
	var originalMachineSetNames []string
	for i, ms := range machinesets1.Items {
		msMap[ms.Name] = &machinesets1.Items[i]
		// skip machinesets labeled as remote
		_, exists := ms.ObjectMeta.Labels["remote"]
		if exists {
			continue
		}
		originalMachineSetNames = append(originalMachineSetNames, ms.Name)
	}
	//fmt.Println("map: ", msMap)
	for _, msName := range originalMachineSetNames {
		if _, ok := msMap[msName+"-remote"]; ok {
			// already copied to remote.
			continue
		}
		ms := *msMap[msName]
		newMS := ms.DeepCopy()
		replicasDesired := int32(0)
		newMS.Spec.Replicas = &replicasDesired
		newObjectMeta := metav1.ObjectMeta{
			Namespace: "openshift-machine-api",
			Name:      msName + "-remote",
			Labels: map[string]string{
				"remote": "remote",
			},
		}
		newMS.ObjectMeta = newObjectMeta
		newMS.Status = machineapi.MachineSetStatus{}
		newMS.Spec.Template.ObjectMeta.Labels["machine.openshift.io/cluster-api-machineset"] = msName + "-remote"
		newMS.Spec.Selector.MatchLabels["machine.openshift.io/cluster-api-machineset"] = msName + "-remote"

		conf, err2 := providerConfigFromMachine(newMS.Spec.Template, codec)
		if err2 != nil {
			fmt.Println("failed to get provider spec: ", err2)
			os.Exit(1)
		}
		conf.UserDataSecret = &corev1.LocalObjectReference{Name: conf.UserDataSecret.Name + "-remote"}
		//fmt.Println("spec: %v", conf)
		newMS.Spec.Template.Spec.ProviderSpec.Value = &runtime.RawExtension{Object: conf}
		newMachineSets = append(newMachineSets, newMS)
	}
	fmt.Println("New machinesets to be created ", len(newMachineSets))

	for _, ms := range newMachineSets {
		if _, err := machineClientset.Create(ms); err != nil {
			fmt.Println("Unable to create machineset ", ms.Name, err)
			os.Exit(1)
		}
	}
	//fmt.Println("machinesets: %v", newMachineSets[0])
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

// providerConfigFromMachine gets the machine provider config MachineSetSpec from the
// specified cluster-api MachineSpec.
func providerConfigFromMachine(machineTemplate machineapi.MachineTemplateSpec, codec *awsprovider.AWSProviderConfigCodec) (*awsprovider.AWSMachineProviderConfig, error) {
	if machineTemplate.Spec.ProviderSpec.Value == nil {
		return nil, fmt.Errorf("unable to find machine provider config: Spec.ProviderSpec.Value is not set")
	}

	var config awsprovider.AWSMachineProviderConfig
	if err := codec.DecodeProviderSpec(&machineTemplate.Spec.ProviderSpec, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

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
