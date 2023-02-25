module github.com/michaelgugino/machineset-copier

go 1.12

require (
	// kube 1.16
	github.com/openshift/cluster-api v0.0.0-20191003080455-24cfb34ea1f9
	k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apimachinery v0.15.7
	k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	sigs.k8s.io/cluster-api-provider-aws v0.0.0-00010101000000-000000000000

)

replace sigs.k8s.io/cluster-api-provider-aws => github.com/openshift/cluster-api-provider-aws v0.2.1-0.20191004064500-548de9226608

replace sigs.k8s.io/controller-runtime => github.com/enxebre/controller-runtime v0.2.0-beta.1.0.20190930160522-58015f7fc885
