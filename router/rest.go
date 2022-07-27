package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/max-gui/charon/internal/pkg/constset"
	"github.com/max-gui/consulagent/pkg/consulhelp"
	"github.com/max-gui/logagent/pkg/logagent"
	"github.com/max-gui/logagent/pkg/logsets"
	"github.com/max-gui/logagent/pkg/routerutil"

	// "github.com/max-gui/charon/internal/pkg/logagent"
	regagent "github.com/max-gui/regagent/pkg/agent"
	"github.com/max-gui/regagent/pkg/ragcli"
	// nethttp "net/http"
)

func SetupRouter() *gin.Engine {
	// gin.New()

	r := gin.New()                      //.Default()
	r.Use(routerutil.GinHeaderMiddle()) // ginHeaderMiddle())
	r.Use(routerutil.GinLogger())       //LoggerWithConfig())
	r.Use(routerutil.GinErrorMiddle())  //ginErrorMiddle())

	// r.Use(ginErrorMiddle())

	// r.Any("/eurekaagent/apps/DEMO/192.168.226.203:demo:9898?status=UP&lastDirtyTimestamp=1631682844761
	r.GET("/actuator/health", health)
	r.Any("/proxy/:serviceid/:env/*path", call)
	r.Any("/call/:from/:env/:dc/:serviceid/*path", sidecall)
	r.Any("/eurekaagent/:appname/:env/*path", regagent.Eurekaagent)
	r.Any("/consulagent/:appname/:env/*path", regagent.Consulagent)
	// r.Any("/call/:serviceid/*path", consulagent)
	return r
}

type Serverlist struct {
	Url    string
	Dc     string
	Env    string
	Region string
	Exturl string
}

func health(c *gin.Context) {
	c.String(http.StatusOK, "online")
}
func call(c *gin.Context) {
	log.Print(*constset.Ingressgate)
	uri := c.Request.RequestURI
	logger := logagent.Inst(c)
	logger.Info(uri)

	region := c.Value("region").(string)

	if strings.HasPrefix(uri, "/actuator/health") {
		c.String(http.StatusOK, "online")
	} else {
		service := c.Param("serviceid")
		env := c.Param("env")
		dc := *logsets.Appdc

		trimstr := "/proxy/" + service + "/" + env + "/:" + *logsets.Port
		// serpath := strings.ReplaceAll(uri, "/proxy/"+service+"/"+env+"/:"+*logsets.Port, "")
		proxy2callee(service, env, dc, region, uri, trimstr, c)
	}
}

func sidecall(c *gin.Context) {
	log.Print(*constset.Ingressgate)
	uri := c.Request.RequestURI
	logger := logagent.Inst(c)
	logger.Info(uri)

	service := c.Param("serviceid")
	env := c.Param("env")
	dc := c.Param("dc")
	caller := c.Param("from")

	flag := false
	consulapps := ragcli.GetConsulapps(caller, env, service, c)
	for _, app := range consulapps {
		if app.Service.Service == service {
			flag = true
			break
		}
	}

	if !flag {
		logger.Info("not found")
		c.String(http.StatusNotFound, "not found")
		return
	}

	region := c.Value("region").(string)

	trimstr := "/call/" + caller + "/" + env + "/" + dc + "/" + service //:from/:env/:dc/:serviceid
	// serpath := strings.ReplaceAll(uri, "/proxy/"+service+"/"+env+"/:"+*logsets.Port, "")

	proxy2callee(service, env, dc, region, uri, trimstr, c)

}

func Das_Rheingold(service, env, dc, region string, eu_service ragcli.EurekaApplication, getConsulServices func(servicename string, c context.Context) []*api.ServiceEntry, c *gin.Context) ([]Serverlist, ragcli.EurekaApplication) {
	// m := func() []*api.ServiceEntry {
	con_service := getConsulServices(service, c)

	if len(con_service) == 0 && len(eu_service.Application.Instance) < 1 {
		eu_service = ragcli.Eurekapp(service, c)
	}

	// 	return con_service
	// }

	// services := m()

	Serverlist := getservices(dc, env, region, con_service, eu_service, c)
	serviceinstance := acheronfull(dc, env, region, Serverlist, c)

	return serviceinstance, eu_service
}
func proxy2callee(service, env, dc, region, uri, trmstr string, c *gin.Context) {

	logger := logagent.Inst(c)

	logger.Info(service)

	fmt.Println("---header/--- ")
	headers := c.Request.Header
	for k, v := range headers {
		fmt.Println(k, v)
	}

	// if region == "" {
	// 	region = "default"
	// }

	logger.Infof("region:%s", region)
	logger.Infof("env:%s", env)
	logger.Infof("dc:%s", dc)

	// euservices := ragcli.EurekaApplication{}

	// services := []*api.ServiceEntry{}
	// m := func(getConsulServices func(servicename string, c context.Context) []*api.ServiceEntry, eu_service ragcli.EurekaApplication) ([]*api.ServiceEntry, ragcli.EurekaApplication) {
	// 	con_service := getConsulServices(service, c)

	// 	if len(con_service) == 0 && len(eu_service.Application.Instance) < 1 {
	// 		eu_service = ragcli.Eurekapp(service, c)
	// 	}

	// 	return con_service, eu_service
	// }

	// services, euservices := m(consulhelp.GetHealthService, ragcli.EurekaApplication{})

	// Serverlist := getservices(dc, env, region, services, euservices, c)
	// serviceinstance := acheronfull(dc, env, region, Serverlist, c)

	serviceinstance, euservices := Das_Rheingold(service, env, dc, region, ragcli.EurekaApplication{}, consulhelp.GetHealthService, c)
	if len(serviceinstance) <= 0 {
		logger.Printf("not found in %s", dc)

		serviceinstance, _ = Das_Rheingold(service, env, dc, region, euservices, consulhelp.GetHealthServiceDc, c)

		// services, euservices = m(consulhelp.GetHealthServiceDc, ragcli.EurekaApplication{})
		// // services = consulhelp.GetHealthServiceDc(service, c)
		// Serverlist = getservices(dc, env, region, services, euservices, c)
		// serviceinstance = acheronfull(dc, env, region, Serverlist, c)
	}

	if len(serviceinstance) <= 0 {
		logger.Info("not found in all dcs")
		c.String(http.StatusNotFound, "not found")
		return
	}
	var index int
	servicemap.help(func(kvs map[string]int) (bool, interface{}) {
		if val, ok := kvs[service]; ok {
			index = val + 1
			kvs[service] = index
		} else {
			kvs[service] = 0
		}

		return true, 0
	})
	instanceindex := index % len(serviceinstance)
	logger.Printf("index:%d", index)
	logger.Printf("instanceindex:%d", instanceindex)
	instance := serviceinstance[instanceindex]

	logger.Printf("instance:%s", instance)

	var serlocation string

	// serpath := strings.ReplaceAll(uri, trmstr "/proxy/"+service+"/"+env+"/:"+*logsets.Port, "")
	serpath := strings.ReplaceAll(uri, trmstr, "")
	if *constset.Ingressgate {
		serlocation = *constset.IngressHost + strings.ToLower(service)

		serpath = strings.Split("service/"+strings.ToLower(service)+serpath, "?")[0]
	} else {
		serlocation = strings.TrimSuffix(instance.Url, "/")
		serpath = strings.Split(serpath, "?")[0]
	}

	logger.Info(serlocation + serpath)

	remote, err := url.Parse(serlocation)
	if err != nil {
		logger.Panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	c.Request.Host = remote.Host

	serpathdecoded, err := url.QueryUnescape(serpath)
	if err != nil {
		logger.Panic(err)
	}
	c.Request.URL.Path = serpathdecoded

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Panic(err)
	}
	logger.Print(c.Request)

	proxy.ServeHTTP(c.Writer, c.Request)
}

type mutexKV struct {
	sync.RWMutex
	kvs map[string]int
}

var servicemap = mutexKV{kvs: make(map[string]int)}

func (v *mutexKV) help(tricky func(map[string]int) (bool, interface{})) (bool, interface{}) {
	v.Lock()
	ok, res := tricky(v.kvs)
	v.Unlock()
	return ok, res
}

func getservices(dc, env, region string, services []*api.ServiceEntry, euservices ragcli.EurekaApplication, c context.Context) map[string]map[string][]Serverlist {

	logger := logagent.Inst(c)
	// serviceinstance := []Serverlist{}
	serviceInfos := map[string]map[string][]Serverlist{}
	serviceInfos[dc] = map[string][]Serverlist{}
	serviceInfos[dc][region] = []Serverlist{}
	serviceInfos[dc]["default"] = []Serverlist{}
	serviceInfos["others"] = map[string][]Serverlist{}
	serviceInfos["others"][region] = []Serverlist{}
	serviceInfos["others"]["default"] = []Serverlist{}

	dctag := ""
	dcurl := ""
	serviceLen := 0
	for _, entry := range services {
		logger.Info(consulhelp.ServiceEntryPrint(entry))
		if entryenv, ok := entry.Service.Meta["x-baggage-AF-env"]; ok {
			if entryregion, ok := entry.Service.Meta["x-baggage-AF-region"]; ok {
				if entrydc, ok := entry.Service.Meta["dc"]; ok {
					// if entryextaddress, ok := entry.Service.Meta["extaddress"]; ok {
					// 	if entryextport, ok := entry.Service.Meta["extport"]; ok {
					entryextaddress, _ := entry.Service.Meta["extaddress"]
					entryextport, _ := entry.Service.Meta["extport"]
					if entrydc == "" {
						entrydc = dc
					}
					if entryregion == "" {
						entryregion = "default"
					}
					if entryenv == "" {
						entryenv = env
					}
					if (entryregion == region || entryregion == "default") && entryenv == env {

						if entrydc != dc {
							dctag = "others"
							dcurl = "http://" + entryextaddress + ":" + entryextport
						} else {
							dctag = dc
							dcurl = "http://" + entry.Service.Address + ":" + strconv.Itoa(entry.Service.Port)
						}

						serviceInfos[dctag][entryregion] = append(serviceInfos[dctag][entryregion],
							Serverlist{
								Url:    dcurl,
								Env:    entryenv,
								Dc:     entrydc,
								Region: entryregion,
								Exturl: "http://" + entry.Service.Address + ":" + strconv.Itoa(entry.Service.Port),
							})
						serviceLen++

					}

				}
			}
		}
	}

	if serviceLen <= 0 {
		for _, instance := range euservices.Application.Instance {
			// logger.Info(instance.Print())
			logger.Info(instance)
			entryenv, envok := instance.Metadata["x-baggage-AF-env"]
			entryregion, regionok := instance.Metadata["x-baggage-AF-region"]

			if entryregion == "" || !regionok {
				entryregion = "default"
			}
			if entryenv == "" || !envok {
				entryenv = env
			}
			if (entryregion == region || entryregion == "default") && entryenv == env {

				serviceInfos[dc][entryregion] = append(serviceInfos[dctag][entryregion],
					Serverlist{
						Url:    instance.HomePageUrl,
						Env:    entryenv,
						Dc:     dc,
						Region: entryregion,
						Exturl: instance.HomePageUrl,
					})
			}

			// if envok && regionok && dcok || env == "" {
			// 	serviceinstance = append(serviceinstance,
			// 		Serverlist{Url: instance.HomePageUrl,
			// 			Env:    entryenv,
			// 			Dc:     entrydc,
			// 			Region: entryregion,
			// 		})
			// }
		}
	}

	logger.Info(serviceInfos)
	return serviceInfos
}

func acheronfull(dc, env, region string, services map[string]map[string][]Serverlist, c context.Context) []Serverlist {

	logger := logagent.Inst(c)
	// var serviceinstance []Serverlist

	f := func(indc, inregion string) []Serverlist {
		if dcservices, ok := services[indc]; ok && len(dcservices) > 0 {
			logger.Info(indc)
			if regionserivces, ok := dcservices[inregion]; ok && len(regionserivces) > 0 {
				logger.Info(inregion)
				return regionserivces
			} else if len(dcservices["default"]) > 0 {
				logger.Info("default")
				return dcservices["default"]
			}
		}
		return nil
	}

	var serviceinstance = f(dc, region)
	if len(serviceinstance) < 1 {
		serviceinstance = f("others", region)
	}

	logger.Info(serviceinstance)
	return serviceinstance
}

func acheron(env string, region string, services []*api.ServiceEntry, euservices ragcli.EurekaApplication, c context.Context) []Serverlist {

	logger := logagent.Inst(c)
	serviceinstance := []Serverlist{}

	for _, entry := range services {
		logger.Info(entry.Service)
		if entryenv, ok := entry.Service.Meta["x-baggage-AF-env"]; ok {
			if entryregion, ok := entry.Service.Meta["x-baggage-AF-region"]; ok {
				if entryenv == env && entryregion == region {
					serviceinstance = append(serviceinstance, Serverlist{Url: "http://" + entry.Service.Address + ":" + strconv.Itoa(entry.Service.Port)})
				}
			}
		}
	}

	if len(serviceinstance) <= 0 {
		for _, instance := range euservices.Application.Instance {
			logger.Info(instance)
			entryenv, envok := instance.Metadata["x-baggage-AF-env"]
			entryregion, regionok := instance.Metadata["x-baggage-AF-region"]

			if envok && regionok && entryenv == env && entryregion == region {
				serviceinstance = append(serviceinstance, Serverlist{Url: instance.HomePageUrl})
			} else if env == "" {
				serviceinstance = append(serviceinstance, Serverlist{Url: instance.HomePageUrl})
			}
		}
	}
	return serviceinstance
}
