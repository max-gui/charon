package router

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/max-gui/charon/internal/pkg/constset"
	"github.com/max-gui/consulagent/pkg/consulhelp"
	"github.com/max-gui/logagent/pkg/logagent"
	"github.com/max-gui/logagent/pkg/logsets"
	"github.com/max-gui/logagent/pkg/routerutil"
	"gopkg.in/yaml.v3"

	regagent "github.com/max-gui/regagent/pkg/agent"
	"github.com/max-gui/regagent/pkg/ragcli"
	"github.com/max-gui/regagent/pkg/regagentsets"
)

func SetupRouter() *gin.Engine {
	// http.DefaultTransport.(*http.Transport).MaxIdleConns = 500
	// http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 500

	if *logsets.Appenv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()                      //.Default()
	r.Use(routerutil.GinHeaderMiddle()) // ginHeaderMiddle())
	r.Use(routerutil.GinLogger())       //LoggerWithConfig())
	r.Use(routerutil.GinErrorMiddle())  //ginErrorMiddle())

	r.GET("/actuator/health", health)
	r.Any("/proxy/:serviceid/:env/*path", call)
	r.Any("/call/:from/:env/:dc/:serviceid/*path", sidecall)
	r.Any("/hostcall/:from/:env/:dc/*path", externalcall) //sidecall)
	r.Any("/agentcall/:from/:env/:dc/:serviceid/*path", sidecall)
	r.Any("/external/:extype/:from/:env/:dc/*path", externalcall)
	r.Any("/eurekaagent/:appname/:env/*path", regagent.Eureka8500)
	r.Any("/consulagent/:appname/:env/*path", regagent.Consulagent8500)
	r.Any("/eureka/:appname/:env/*path", regagent.Eureka80)
	r.Any("/consul/:appname/:env/*path", regagent.Consulagent80)
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
	// log.Print(*constset.Ingressgate)
	uri := c.Request.RequestURI
	logger := logagent.InstArch(c)
	logger.Info(uri)

	region := c.Value("region").(string)

	// if strings.HasPrefix(uri, "/actuator/health") {
	// 	c.String(http.StatusOK, "online")
	// } else {
	service := c.Param("serviceid")
	env := c.Param("env")
	dc := *logsets.Appdc

	trimstr := "/proxy/" + service + "/" + env + "/:" + *logsets.Port
	// serpath := strings.ReplaceAll(uri, "/proxy/"+service+"/"+env+"/:"+*logsets.Port, "")
	proxy2callee(service, env, dc, region, uri, trimstr, false, c)
	// }
}

func isApproved(extype, service string, srvmap map[string]map[string]interface{}, isAllpath bool, c context.Context) (bool, string, map[string]string, string) {
	flag := true
	msg := ""
	notfound := "service isnt in arch service list"
	headers := make(map[string]string)
	qappend := ""
	switch extype {
	case "hostsrv":
		if _, isApproved := srvmap["hostsrvs"][strings.ToLower(service)]; !isApproved {
			flag = false
			msg = notfound
		}
	case "extsrv":

		if _, isApproved := srvmap["extsrvs"][strings.ToLower(service)]; !isApproved {
			flag = false
			msg = notfound
		}
		headers["Af-Ext-Appid"] = service
		headers["Af-Ext-Token"] = "fls-aflm"
		headers["Clientid"] = "fls-aflm"
		headers["Serviceid"] = service

		// .queryParam("appId", "FLS-AFLM")
		// .queryParam("serviceId", "esg-inner-service-qirong")
		saasid := strings.TrimSuffix(service, ".saas")
		qappend = fmt.Sprintf("appId=FLS-AFLM&serviceId=%s", saasid)

		// c.Header("Af-Ext-Appid", service)
		// c.Header("Af-Ext-Token", "fls-aflm")
		// c.Header("Clientid", "fls-aflm")
		// c.Header("Serviceid", service)
	case "gwcode":

		if _, isApproved := srvmap["extcodes"][strings.ToLower(service)]; !isApproved {
			flag = false
			msg = notfound
		}
		headers["Clientid"] = "fls-aflm"
		headers["Serviceid"] = service
		// c.Header("Clientid", "fls-aflm")
		// c.Header("Serviceid", service)
	case "":

		if _, isApproved := srvmap["services"][strings.ToLower(service)]; !isAllpath && !isApproved {
			flag = false
			msg = notfound
		}

	}

	return flag, msg, headers, qappend
}

func routingPrepare(extype, service, env, dc, pathUri string, headers map[string]string, c context.Context) (bool, string, string) {
	skipRouting := false
	proxyUri := pathUri
	proxysrv := service
	logger := logagent.InstArch(c)
	if extype != "" {
		confbytes := consulhelp.Getconfibytes(*regagentsets.ConfResPrefix, extype, service, env+dc, c)

		maptmp := make(map[string]string)
		err := yaml.Unmarshal(confbytes, &maptmp)
		if err != nil {
			logger.Panic(err)
		}

		srvhost := maptmp["host"]
		uripara := maptmp["uri"]
		if !strings.Contains(uripara, ".afproxy") {
			skipRouting = true
			proxyUri = fmt.Sprintf("%s%s", uripara, pathUri)
		} else {
			uritmp := strings.TrimPrefix(uripara, "http://")
			uritmp = strings.TrimPrefix(uritmp, "https://")
			uriarr := strings.Split(uripara, ".afproxy")
			proxysrv = strings.Split(uritmp, ".afproxy")[0]
			srvpath := ""
			if len(uriarr) > 1 {

				srvpath = strings.TrimSuffix(strings.Split(uripara, ".afproxy")[1], "/")
				srvpath = strings.TrimPrefix(srvpath, "/")
			}
			proxyUri = fmt.Sprintf("/%s%s", srvpath, proxyUri)
		}

		logger.Info(srvhost)
		logger.Info(proxyUri)
		headers["srvhost"] = srvhost
		// c.Header("srvhost", srvhost)
	}

	return skipRouting, proxyUri, proxysrv
}

func externalcall(c *gin.Context) {

	// log.Print(*constset.Ingressgate)
	uri := c.Request.RequestURI
	logger := logagent.InstArch(c)
	logger.Info(uri)

	extype := c.Param("extype")
	env := c.Param("env")
	dc := c.Param("dc")
	caller := c.Param("from")
	logger.Info("hostname:" + c.Request.Host)
	var service = strings.TrimSuffix(c.Request.Host, ".afproxy")

	isAllpath, srvmap := ragcli.GetAproveServicesSingle(caller, c)

	approved, msg, headers, qappend := isApproved(extype, service, srvmap, isAllpath, c)
	if !approved {
		logger.Info("not found")
		c.String(http.StatusNotAcceptable, msg)
		return
	}

	region := c.Value("region").(string)
	proxysrv := service
	redipath := c.Param("path")
	rawq, err := url.QueryUnescape(c.Request.URL.RawQuery)
	if err != nil {
		logger.Panic(err)
	}

	// skipRouting := false
	// logger.Info(proxyuri)
	// if extype != "" {
	// 	confbytes := consulhelp.Getconfibytes(*regagentsets.ConfResPrefix, extype, service, env+dc, c)

	// 	maptmp := make(map[string]string)
	// 	err := yaml.Unmarshal(confbytes, &maptmp)
	// 	if err != nil {
	// 		logger.Panic(err)
	// 	}

	// 	srvhost := maptmp["host"]
	// 	uripara := maptmp["uri"]
	// 	if !strings.Contains(uripara, ".afproxy") {
	// 		skipRouting = true
	// 		proxyuri = fmt.Sprintf("%s%s", uripara, proxyuri)
	// 	} else {
	// 		uritmp := strings.TrimPrefix(uripara, "http://")
	// 		uritmp = strings.TrimPrefix(uritmp, "https://")
	// 		uriarr := strings.Split(uripara, ".afproxy")
	// 		proxysrv = strings.Split(uritmp, ".afproxy")[0]
	// 		srvpath := ""
	// 		if len(uriarr) > 1 {

	// 			srvpath = strings.TrimSuffix(strings.Split(uripara, ".afproxy")[1], "/")
	// 			srvpath = strings.TrimPrefix(srvpath, "/")
	// 		}
	// 		proxyuri = fmt.Sprintf("/%s%s", srvpath, proxyuri)
	// 	}

	// 	logger.Info(srvhost)
	// 	logger.Info(proxyuri)
	// 	c.Header("srvhost", srvhost)
	// }
	skipRouting, proxyuri, proxysrv := routingPrepare(extype, service, env, dc, "", headers, c)

	if rawq != "" {
		rawq = fmt.Sprintf("?%s", rawq)
		if qappend != "" && skipRouting {
			rawq = fmt.Sprintf("%s&%s", rawq, qappend)
		}
	} else if qappend != "" && skipRouting {
		rawq = fmt.Sprintf("?%s", qappend)
	}

	logger.Info(c.Request.URL.RawQuery)

	proxyurifull := fmt.Sprintf("%s%s%s", proxyuri, redipath, rawq)

	logger.Info(proxyurifull)
	for k, v := range headers {
		c.Header(k, v)
	}
	if skipRouting {
		c.Redirect(301, proxyurifull)
	} else {
		proxy2calleeuri(proxysrv, env, dc, region, proxyurifull, true, c)
	}

}

func sidecall(c *gin.Context) {
	// log.Print(*constset.Ingressgate)
	uri := c.Request.RequestURI
	logger := logagent.InstArch(c)
	logger.Info(uri)

	service := c.Param("serviceid")
	// service = strings.ToLower(service)
	env := c.Param("env")
	dc := c.Param("dc")
	caller := c.Param("from")
	logger.Info("hostname:" + c.Request.Host)
	hostmode := false
	if service == "" && strings.HasSuffix(c.Request.Host, ".afproxy") {
		// %s.afproxy"
		//agservice.Address = fmt.Sprintf("%s.%s.%s.%s.afproxy", appname, env, *logsets.Appdc, service)
		infos := strings.TrimSuffix(c.Request.Host, ".afproxy")
		// infos := strings.Split(c.Request.Host, ".")
		service = infos
		hostmode = true
		// if len(infos) > 3 {
		// 	// caller = infos[0]
		// 	env = infos[0]
		// 	dc = infos[1]
		// 	service = infos[2]
		// 	if caller == "" {
		// 		ip := strings.Split(c.RemoteIP(), ":")[0]
		// 		caller = consulhelp.GetAgentServices(ip, c)
		// 	}
		// } else {
		// 	logger.Panic("hostname format is wrong")
		// }
		// strings.Split(infos, ".")
	} else if service == "" {
		logger.Panic("no from and no *.afproxy host and no serviceid para")
	}

	isAllpath, consulapps, _, _, _ := ragcli.GetAproveServices(caller, c)
	// service_lowcase := strings.ToLower(service)
	if _, isApproved := consulapps[strings.ToLower(service)]; !isAllpath && !isApproved {
		logger.Info("not found")
		c.String(http.StatusNotAcceptable, "service isnt in arch service list")
		return
	}
	// flag := false

	// for _, app := range consulapps {
	// 	if app.Service.Service == service {
	// 		flag = true
	// 		break
	// 	}
	// }

	// if !flag {
	// 	logger.Info("not found")
	// 	c.String(http.StatusNotFound, "not found")
	// 	return
	// }

	region := c.Value("region").(string)
	var trimstr string
	if strings.HasPrefix(uri, "/agentcall") {
		trimstr = "/agentcall/" + caller + "/" + env + "/" + dc + "/" + service //:from/:env/:dc/:serviceid
		// serpath := strings.ReplaceAll(uri, "/proxy/"+service+"/"+env+"/:"+*logsets.Port, "")
	} else if strings.HasPrefix(uri, "/hostcall") {
		trimstr = "/hostcall/" + caller + "/" + env + "/" + dc
	} else {
		trimstr = "/call/" + caller + "/" + env + "/" + dc + "/" + service
	}
	proxy2callee(service, env, dc, region, uri, trimstr, hostmode, c)

}

func Das_Rheingold(service, env, dc, region string,
	eu_service ragcli.EurekaApplication, skipEureka bool,
	getConsulServices func(servicename string, c context.Context) []*api.ServiceEntry, c *gin.Context) ([]Serverlist, ragcli.EurekaApplication) {
	// m := func() []*api.ServiceEntry {
	con_service := getConsulServices(service, c)
	if !skipEureka && len(con_service) == 0 && len(eu_service.Application.Instance) < 1 {
		eu_service, _ = ragcli.Eurekapp(service, c)

	}

	// 	return con_service
	// }

	// services := m()

	Serverlist := getservices(dc, env, region, con_service, eu_service, c)
	serviceinstance := acheronfull(dc, env, region, Serverlist, c)

	return serviceinstance, eu_service
}

func proxy2callee(service, env, dc, region, uri, trmstr string, hostmode bool, c *gin.Context) {

	logger := logagent.InstArch(c)

	logger.Info(service)

	// fmt.Println("---header/--- ")
	// headers := c.Request.Header
	// for k, v := range headers {
	// 	fmt.Println(k, v)
	// }

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

	serviceinstance, euservices := Das_Rheingold(service, env, dc, region, ragcli.EurekaApplication{}, false, consulhelp.GetHealthService, c)
	if len(serviceinstance) <= 0 && *logsets.Appdc != "DR" {
		logger.Printf("not found in %s", dc)

		serviceinstance, _ = Das_Rheingold(service, env, dc, region, euservices, true, consulhelp.GetHealthServiceDc, c)

		// services, euservices = m(consulhelp.GetHealthServiceDc, ragcli.EurekaApplication{})
		// // services = consulhelp.GetHealthServiceDc(service, c)
		// Serverlist = getservices(dc, env, region, services, euservices, c)
		// serviceinstance = acheronfull(dc, env, region, Serverlist, c)
	}
	var instance Serverlist
	// if len(serviceinstance) > 0 {
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
	instance = serviceinstance[instanceindex]
	// } else {
	// 	instance = Serverlist{Url: "http://127.0.0.1:8080"}
	// }
	logger.Printf("instance:%s", instance)
	if hostmode {
		redirectUri := strings.ReplaceAll(c.Request.RequestURI, trmstr, strings.TrimSuffix(instance.Url, "/"))
		logger.Info(redirectUri)
		c.Redirect(301, redirectUri)
	} else {
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
		start := c.Value("starttime").(time.Time)
		timestamp := time.Since(start).Milliseconds()
		logger.WithField("LB_time_span", timestamp).Print(c.Request)

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func proxy2calleeuri(service, env, dc, region, uri string, hostmode bool, c *gin.Context) {

	logger := logagent.InstArch(c)

	logger.Info(service)

	// fmt.Println("---header/--- ")
	// headers := c.Request.Header
	// for k, v := range headers {
	// 	fmt.Println(k, v)
	// }

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

	serviceinstance, euservices := Das_Rheingold(service, env, dc, region, ragcli.EurekaApplication{}, false, consulhelp.GetHealthService, c)
	if len(serviceinstance) <= 0 && *logsets.Appdc != "DR" {
		logger.Printf("not found in %s", dc)

		serviceinstance, _ = Das_Rheingold(service, env, dc, region, euservices, true, consulhelp.GetHealthServiceDc, c)

		// services, euservices = m(consulhelp.GetHealthServiceDc, ragcli.EurekaApplication{})
		// // services = consulhelp.GetHealthServiceDc(service, c)
		// Serverlist = getservices(dc, env, region, services, euservices, c)
		// serviceinstance = acheronfull(dc, env, region, Serverlist, c)
	}
	var instance Serverlist
	// if len(serviceinstance) > 0 {
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
	instance = serviceinstance[instanceindex]
	// } else {
	// 	instance = Serverlist{Url: "http://127.0.0.1:8080"}
	// }
	logger.Printf("instance:%s", instance)
	if hostmode {
		// redirectUri := strings.ReplaceAll(c.Request.RequestURI, trmstr, strings.TrimSuffix(instance.Url, "/"))
		redirectUri := fmt.Sprintf("%s%s", instance.Url, uri)
		logger.Info(redirectUri)
		c.Redirect(301, redirectUri)
	} else {
		var serlocation string

		// serpath := strings.ReplaceAll(uri, trmstr "/proxy/"+service+"/"+env+"/:"+*logsets.Port, "")
		serpath := uri
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

		// serpathdecoded, err := url.QueryUnescape(serpath)
		if err != nil {
			logger.Panic(err)
		}
		c.Request.URL.Path = serpath //serpathdecoded

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Panic(err)
		}
		start := c.Value("starttime").(time.Time)
		timestamp := time.Since(start).Milliseconds()
		logger.WithField("LB_time_span", timestamp).Print(c.Request)

		proxy.ServeHTTP(c.Writer, c.Request)
	}
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

	logger := logagent.InstArch(c)
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
				entrydc, ok := entry.Service.Meta["dc"]
				// if entryextaddress, ok := entry.Service.Meta["extaddress"]; ok {
				// 	if entryextport, ok := entry.Service.Meta["extport"]; ok {
				entryextaddress := entry.Service.Meta["extaddress"]
				entryextport := entry.Service.Meta["extport"]
				if entrydc == "" || !ok {
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

				serviceInfos[dc][entryregion] = append(serviceInfos[dc][entryregion],
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

	logger := logagent.InstArch(c)
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

	logger := logagent.InstArch(c)
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
