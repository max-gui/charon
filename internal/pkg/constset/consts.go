package constset

import (
	"flag"
	"os"

	"github.com/max-gui/logagent/pkg/logsets"
	"github.com/max-gui/regagent/pkg/regagentsets"
)

const PthSep = string(os.PathSeparator)

var Key = []byte{74, 103, 115, 173, 168, 227, 72, 68, 25, 245, 63, 49, 136, 236, 197, 236}
var Nonce = []byte{9, 65, 48, 149, 170, 165, 84, 222, 74, 84, 4, 106}

var (
	// ConfWatchPrefix *string
	// Cacheminutes *int
	Ingressgate *bool
	IngressHost *string
)

func StartupInit(bytes []byte) {

	regagentsets.StartupInit(bytes)
	regagentsets.AgentPort = logsets.Port
	// bytes, err := os.ReadFile(*Apppath + string(os.PathSeparator) + "application-" + *logsets.Appenv + ".yml")
	// if err != nil {
	// 	log.Panic(err)
	// }
	// confmap := map[string]interface{}{}
	// yaml.Unmarshal(bytes, confmap)
	// *consulsets.Acltoken = confmap["af-arch"].(map[interface{}]interface{})["resource"].(map[interface{}]interface{})["private"].(map[interface{}]interface{})["acl-token"].(string)
	// eu_hosts := confmap["af-arch"].(map[interface{}]interface{})["resource"].(map[interface{}]interface{})["agent"].(map[interface{}]interface{})["client"].(map[interface{}]interface{})["serviceUrl"].(map[interface{}]interface{})["defaultZone"].(string)
	// *regagentsets.Eu_host = strings.Split(eu_hosts, "/,")[0]
}

func init() {
	// logsets.Appname  Appname = "charon"
	*logsets.Appname = "charon"
	Ingressgate = flag.Bool("ingressgate", false, "ingressgate or not")
	// Cacheminutes = flag.Int("cacheminutes", 5, "service cache for minutes")
	IngressHost = flag.String("ingresshost", "", "host pash for ingress gate")
	// ConfWatchPrefix = flag.String("ConfWatchPrefix", "ops/", "watch prefix for consul")

	// testing.Init()
	// flag.Parse()

}

// var Reppath = func() string {
// 	return *Apppath + PthSep + *Repopathname + PthSep
// }
