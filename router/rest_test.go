package router

import (
	"context"
	"log"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/max-gui/consulagent/pkg/consulhelp"
	"github.com/max-gui/regagent/pkg/ragcli"
	"github.com/stretchr/testify/assert"
)

func init() {
	testing.Init()
	// constset.StartupInit()
	// flag.Parse()
}

// 30.147.124.182 consul-szf-prod.kube.com
// 245d0a09-7139-config-prod-ff170a0562b1

func Test_sidecall(t *testing.T) {

	var c = context.Background()
	var service = "fls-aflm-nas-client"
	var flag = false
	consulapps := ragcli.GetConsulapps("fls-ls-commons", "test", service, true, c)
	for _, app := range consulapps {
		if app.Service.Service == service {
			flag = true
			break
		}
	}

	if !flag {
		log.Print("not found")

	}
}

func Test_tconsuldc(t *testing.T) {

	c := gin.Context{}
	c.Request = &http.Request{}
	var caller = "fls-ls-common"
	var env = "test"
	var dc = "LFB"
	var region = "default"
	var service = "fls-aflm-nas-client"
	trimstr := "/call/" + caller + "/" + env + "/" + dc + "/" + service
	proxy2callee(service, env, dc, region, "", trimstr, false, &c)
}

func Test_getservices(t *testing.T) {

	c := context.Background()
	dc := "LFB"
	env := "prod"
	region := "default"
	conservices := []*api.ServiceEntry{{Service: &api.AgentService{

		Meta:              map[string]string{"x-baggage-AF-env": env, "x-baggage-AF-region": region},
		Port:              0,
		Address:           "11111",
		SocketPath:        "",
		TaggedAddresses:   map[string]api.ServiceAddress{},
		Weights:           api.AgentWeights{},
		EnableTagOverride: false,
		CreateIndex:       0,
		ModifyIndex:       0,
		ContentHash:       "",
		Proxy:             &api.AgentServiceConnectProxyConfig{},
		Connect:           &api.AgentServiceConnect{},
		Namespace:         "",
		Partition:         "",
		Datacenter:        dc,
	}}}
	eurekaservices := ragcli.EurekaApplication{Application: ragcli.Eurekaappinfo{Name: "a", Instance: []ragcli.EurekaInstance{
		{HomePageUrl: "bb"},
		{HomePageUrl: "cc"},
	}}}
	res := getservices(dc, env, region, conservices, eurekaservices, c)
	log.Print(res)
}

func Test_acheronfull(t *testing.T) {
	c := context.Background()
	services := map[string]map[string][]Serverlist{}
	services["LFB"] = map[string][]Serverlist{}
	services["others"] = map[string][]Serverlist{}
	services["AAA"] = map[string][]Serverlist{}
	services["LFB"]["a"] = []Serverlist{{Url: "LFBa1"}, {Url: "LFBa2"}}
	services["LFB"]["default"] = []Serverlist{{Url: "LFBdefault1"}, {Url: "LFBdefault2"}}
	services["AAA"]["ab"] = []Serverlist{{Url: "LFBa1"}, {Url: "LFBa2"}}
	// services["AAA"]["default"] = []Serverlist{{Url: "LFBdefault1"}, {Url: "LFBdefault2"}}
	services["others"]["a"] = []Serverlist{{Url: "othersa1"}, {Url: "othersa2"}}
	services["others"]["default"] = []Serverlist{{Url: "othersdefault1"}, {Url: "othersdefault2"}}

	reslist := acheronfull("LFB", "test", "default", services, c)
	assert.Equal(t, services["LFB"]["default"], reslist, "LFB, test, default failed")

	reslist = acheronfull("LFE", "test", "default", services, c)
	assert.Equal(t, services["others"]["default"], reslist, "LFE, test, default failed")

	reslist = acheronfull("LFB", "test", "", services, c)
	assert.Equal(t, services["LFB"]["default"], reslist, "LFB, test, EMPTY failed")

	reslist = acheronfull("LFE", "test", "", services, c)
	assert.Equal(t, services["others"]["default"], reslist, "LFE, test, EMPTY failed")

	reslist = acheronfull("LFB", "test", "a", services, c)
	assert.Equal(t, services["LFB"]["a"], reslist, "LFB, test, a failed")

	reslist = acheronfull("LFE", "test", "a", services, c)
	assert.Equal(t, services["others"]["a"], reslist, "LFE, test, a failed")

	reslist = acheronfull("AAA", "test", "a", services, c)
	assert.Equal(t, services["others"]["a"], reslist, "AAA, test, a failed")

	reslist = acheronfull("AAA", "test", "b", services, c)
	assert.Equal(t, services["others"]["default"], reslist, "AAA, test, b failed")

}

func Test_consuldc(t *testing.T) {
	c := context.Background()
	m := consulhelp.GetHealthServiceDc("af-front-platform-admin-external", c)
	n := consulhelp.GetHealthService("af-front-platform-admin-external", c)

	log.Printf("%+v", m[0])
	log.Printf("%+v", n[0])
}
func Test_ArchDef_commit_check(t *testing.T) {
	env := "test"
	region := "default"
	consulapps := []*api.ServiceEntry{
		{
			Node: &api.Node{},
			Service: &api.AgentService{
				Meta: map[string]string{
					"x-baggage-AF-env":    "test",
					"x-baggage-AF-region": "default",
				},
			},
		}}

	euservices := ragcli.EurekaApplication{
		Application: ragcli.Eurekaappinfo{
			Name: "",
			Instance: []ragcli.EurekaInstance{
				{
					Metadata: map[string]string{
						"x-baggage-AF-env":    "test",
						"x-baggage-AF-region": "default",
					},
				},
			},
		},
	}

	ff := func(env string, region string, services []*api.ServiceEntry, euservices ragcli.EurekaApplication) []Serverlist {
		c := context.Background()
		serviceinstance := acheron(env, region, services, euservices, c)

		if len(serviceinstance) == 0 {
			serviceinstance = acheron(env, "default", services, euservices, c)
			if len(serviceinstance) == 0 {
				serviceinstance = acheron("", "", services, euservices, c)

				t.Log("third")
			} else {

				t.Log("twice")
			}
		} else {

			t.Log("once")
		}
		t.Log(serviceinstance)
		return serviceinstance
		// serverlist := acheron(env, region, consulapps, euservices, context.Background())
	}
	assert.Equal(t, 2, len(ff(env, region, consulapps, euservices)))

	env = "test"
	region = ""
	assert.Equal(t, 2, len(ff(env, region, consulapps, euservices)))

	env = "test"
	region = ""
	consulapps = []*api.ServiceEntry{
		{
			Node: &api.Node{},
			Service: &api.AgentService{
				Meta: map[string]string{
					"x-baggage-AF-env":    "test",
					"x-baggage-AF-region": "default",
				},
			},
		}}

	euservices = ragcli.EurekaApplication{
		Application: ragcli.Eurekaappinfo{
			Name: "",
			Instance: []ragcli.EurekaInstance{
				{
					Metadata: map[string]string{
						"x-baggage-AF-env":    "test",
						"x-baggage-AF-region": "",
					},
				},
			},
		},
	}
	assert.Equal(t, 1, len(ff(env, region, consulapps, euservices)))

	env = "test"
	region = ""
	consulapps = []*api.ServiceEntry{
		{
			Node: &api.Node{},
			Service: &api.AgentService{
				Meta: map[string]string{
					"x-baggage-AF-env": "test",
					// "x-baggage-AF-region": "",
				},
			},
		}}

	euservices = ragcli.EurekaApplication{
		Application: ragcli.Eurekaappinfo{
			Name: "",
			Instance: []ragcli.EurekaInstance{
				{
					Metadata: map[string]string{
						"x-baggage-AF-env": "",
						// "x-baggage-AF-region": "",
					},
				},
			},
		},
	}
	assert.Equal(t, 1, len(ff(env, region, consulapps, euservices)))
	// t.Log(serverlist)
}

// func TestMain(m *testing.M) {
// 	setup()
// 	// constset.StartupInit()
// 	// sendconfig2consul()
// 	// configgen.Getconfig = getTestConfig

// 	exitCode := m.Run()
// 	teardown()
// 	// // 退出
// 	os.Exit(exitCode)
// }
