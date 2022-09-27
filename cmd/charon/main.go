package main

import (
	"flag"

	"github.com/gin-contrib/pprof"
	"github.com/max-gui/charon/internal/pkg/constset"
	"github.com/max-gui/charon/router"
	"github.com/max-gui/consulagent/pkg/consulhelp"
	"github.com/max-gui/logagent/pkg/confload"
	"github.com/max-gui/logagent/pkg/logagent"
	"github.com/max-gui/logagent/pkg/logsets"
	"github.com/max-gui/regagent/pkg/regagentsets"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func main() {

	flag.Parse()
	// router.Test(os.Args[1])
	// var Argsetmap = make(map[string]interface{})

	c := logagent.GetRootContextWithTrace()
	bytes := confload.Load(c)
	constset.StartupInit(bytes)

	go consulhelp.StartWatch(*regagentsets.ConfWatchPrefix, true, c)
	// config := consulhelp.Getconfaml(*constset.ConfResPrefix, "redis", "redis-sentinel-proxy", *constset.Appenv)
	// redisops.Url = config["url"].(string)
	// redisops.Pwd = config["password"].(string)
	// go consulhelp.StartWatch(*constset.ConfWatchPrefix, true)

	// if len(os.Args) > 2 {
	// 	port = os.Args[1]
	// }
	// if len(os.Args) > 2 {
	// 	consulhelp.Consulurl = os.Args[2]
	// }
	// if len(os.Args) > 3 {
	// 	consulhelp.AclToken = os.Args[3]
	// }

	// router.Envs
	//port := "4000"

	// githelp.UpdateAll()
	r := router.SetupRouter()
	p := ginprometheus.NewPrometheus("gin")
	pprof.Register(r)
	p.Use(r)

	r.Run(":" + *logsets.Port)
}
