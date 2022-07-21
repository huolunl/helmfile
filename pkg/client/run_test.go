package client

import (
	"log"
	"testing"
)

func TestExec(t *testing.T) {
	out, err := Exec([]string{"--kube-token", "eyJhbGciOiJSUzI1NiIsImtpZCI6InM3a3NnUVdudUFJTVNNTmtaWlA2MmdyS0VGUnkteVhyRDBVbWpNeXJ6VkUifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJuaWthIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6Im5pa2EtdG9rZW4tc2hrcmwiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoibmlrYSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjI0MjJiMTg4LWRmNjAtNDMyMi1iNmFiLWExNzEwZTlhZDZlZSIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpuaWthOm5pa2EifQ.TqoCWT2euXp1BTbapWh6MaJE9Tqa21sci4LbRMG2Y9Vwg-C_hCL9vDXzVOlfHUFk8-VNc-qZE4Ga2lqsi89z3BO-dddoTKEeznSFetNYksVWPa8rbPI7yOX4_ZhfPbOUYvwltCC-KnLSW1uf5Mu1MchWrSMO0zXey4GbL4T7nb0VNBDebRm6wQLwwukWNSBTT_vv4GK-b6mv1QNjn7hiIz1LuvPuuIqKpBZkAw7tA3dlnsk4MbNJGNvhcxlfEqIn0xjA_8sYCiCZbhaBGqvPAIsUY-Sj4yPmDTeWJvWv66dl_CCTfL0Alzn2zzZrETZ-TVAHdwOiceV5gOYApb4IYQ",
		"--kube-apiserver", "https://172.29.2.206:6443", "--kube-ca-file", "/Users/huolun/.kube/ca.crt"}, "dddddqwdqwdqwd", "helmfile", "-f", "/Users/huolun/project/huolun/helmfile/examples/test/helmfile.yaml", "delete")
	log.Println(string(out))
	log.Println(err)
}
