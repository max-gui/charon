package router

import (
	"context"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/max-gui/regagent/pkg/ragcli"
	"github.com/stretchr/testify/assert"
)

func init() {
	testing.Init()
	// constset.StartupInit()
	// flag.Parse()
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
