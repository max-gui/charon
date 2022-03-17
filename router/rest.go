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
	r.Any("/call/:from/:env/:serviceid/*path", sidecall)
	r.Any("/eurekaagent/:appname/:env/*path", regagent.Eurekaagent)
	r.Any("/consulagent/:appname/:env/*path", regagent.Consulagent)
	// r.Any("/call/:serviceid/*path", consulagent)
	return r
}

// func LoggerWithConfig() gin.HandlerFunc {

// 	return func(c *gin.Context) {
// 		// Start timer
// 		start := time.Now()
// 		path := c.Request.URL.Path
// 		raw := c.Request.URL.RawQuery

// 		// Process request
// 		c.Next()
// 		if raw != "" {
// 			path = path + "?" + raw
// 		}
// 		timestamp := time.Since(start)
// 		var infomsg string
// 		if c.Errors.String() == "" {
// 			infomsg = fmt.Sprintf("%s %s %s from %s cost %s;bodysize is %s;",
// 				strconv.Itoa(c.Writer.Status()), c.Request.Method, path, c.ClientIP(), timestamp, strconv.Itoa(c.Writer.Size()))
// 		} else {
// 			infomsg = fmt.Sprintf("%s %s %s from %s cost %s;bodysize is %s;errormsg: %s",
// 				strconv.Itoa(c.Writer.Status()), c.Request.Method, path, c.ClientIP(), timestamp, strconv.Itoa(c.Writer.Size()), c.Errors.String())
// 		}
// 		logagent.Inst(c).
// 			WithField("timestamp", timestamp).
// 			WithField("clientip", c.ClientIP()).
// 			WithField("method", c.Request.Method).
// 			WithField("statuscode", c.Writer.Status()).
// 			WithField("error", c.Errors.String()).
// 			WithField("bodysize", c.Writer.Size()).
// 			WithField("path", path).Infof(infomsg)
// 		// Log only when path is not being skipped

// 	}
// }

// func ginHeaderMiddle() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// // "trace": "%X{X-B3-TraceId:-}",？
// 		// // "span": "%X{X-B3-SpanId:-}",？
// 		// // "parent": "%X{X-B3-ParentSpanId:-}",？
// 		// // "x-baggage-AF-env": "%X{x-baggage-AF-env:-}",？
// 		// // "x-baggage-AF-region": "%X{x-baggage-AF-region:-}",？
// 		trace := c.Request.Header.Get("X-B3-TraceId")
// 		if trace == "" {
// 			trace = strings.ReplaceAll(uuid.NewString(), "-", "")
// 		}
// 		span := strings.ReplaceAll(uuid.NewString(), "-", "")[0:16]
// 		// c.Set("trace", c.Request.Header.Get("X-B3-TraceId"))
// 		// c.Set("span", c.Request.Header.Get("X-B3-SpanId"))
// 		c.Set("trace", trace)
// 		c.Set("span", span)
// 		// c.Set("parentspanid", c.Request.Header.Get("X-B3-ParentSpanId"))
// 		c.Set("env", c.Request.Header.Get("x-baggage-AF-env"))
// 		c.Set("region", c.Request.Header.Get("x-baggage-AF-region"))
// 		// logger := logagent.Inst(c)
// 		// logger.Print(c.Request.Header)
// 		// c.Set("log", logagent.Inst(c))
// 		// log.Print(c.Get("trace"))
// 		// log.Print(c.Get("span"))
// 		// log.Print(c.Get("X-B3-ParentSpanId"))
// 		// log.Print(c.Value("env"))
// 		// log.Print(c.Get("env"))
// 		// log.Print(c.Get("region"))
// 		c.Next()
// 		// host := c.Request.Host
// 		// fmt.Printf("Before: %s\n", host)
// 		// c.Next()
// 		// fmt.Println("Next: ...")
// 	}
// }

// func ginErrorMiddle() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		defer func() {
// 			if e := recover(); e != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{
// 					"msg": fmt.Sprint(e),
// 				})
// 				logger := logagent.Inst(c)
// 				logger.Panic(e)
// 			}
// 		}()

// 		c.Next()
// 		// host := c.Request.Host
// 		// fmt.Printf("Before: %s\n", host)
// 		// c.Next()
// 		// fmt.Println("Next: ...")
// 	}
// }

// type EurekaApplications struct {
// 	Applications struct {
// 		Versions__delta string          `json:"versions__delta"`
// 		Apps__hashcode  string          `json:"apps__hashcode"`
// 		Application     []Eurekaappinfo `json:"application"`
// 	} `json:"applications"`
// }

// type EurekaInstance struct {
// 	InstanceId       string `json:"instanceId"`
// 	HostName         string `json:"hostName"`
// 	App              string `json:"app"`
// 	Status           string `json:"status"`
// 	Overriddenstatus string `json:"overriddenstatus"`
// 	IpAddr           string `json:"ipAddr"`
// 	Port             struct {
// 		Realport int    `json:"$"`
// 		Enabled  string `json:"@enabled"`
// 	} `json:"port"`
// 	SecurePort struct {
// 		Realport int    `json:"$"`
// 		Enabled  string `json:"@enabled"`
// 	} `json:"securePort"`
// 	CountryId      int `json:"countryId"`
// 	DataCenterInfo struct {
// 		Class string `json:"@class"`
// 		Name  string `json:"name"`
// 	} `json:"dataCenterInfo"`
// 	LeaseInfo struct {
// 		RenewalIntervalInSecs int   `json:"renewalIntervalInSecs"`
// 		DurationInSecs        int   `json:"durationInSecs"`
// 		RegistrationTimestamp int64 `json:"registrationTimestamp"`
// 		LastRenewalTimestamp  int64 `json:"lastRenewalTimestamp"`
// 		EvictionTimestamp     int64 `json:"evictionTimestamp"`
// 		ServiceUpTimestamp    int64 `json:"serviceUpTimestamp"`
// 	} `json:"leaseInfo"`
// 	Metadata                      map[string]string `json:"metadata"`
// 	HomePageUrl                   string            `json:"homePageUrl"`
// 	StatusPageUrl                 string            `json:"statusPageUrl"`
// 	HealthCheckUrl                string            `json:"healthCheckUrl"`
// 	VipAddress                    string            `json:"vipAddress"`
// 	SecureVipAddress              string            `json:"secureVipAddress"`
// 	IsCoordinatingDiscoveryServer string            `json:"isCoordinatingDiscoveryServer"`
// 	LastUpdatedTimestamp          string            `json:"lastUpdatedTimestamp"`
// 	LastDirtyTimestamp            string            `json:"lastDirtyTimestamp"`
// 	ActionType                    string            `json:"actionType"`
// }

// type EurekaApplication struct {
// 	Application Eurekaappinfo `json:"application"`
// }

// type Eurekaappinfo struct {
// 	Name     string           `json:"name"`
// 	Instance []EurekaInstance `json:"instance"`
// }

// func tocall(method, url string, heads map[string]string, c context.Context) *http.Response {
// 	logger := logagent.Inst(c)
// 	var netTransport = &http.Transport{
// 		Dial: (&net.Dialer{
// 			Timeout: 5 * time.Second,
// 		}).Dial,
// 		TLSHandshakeTimeout: 5 * time.Second,
// 	}
// 	var netClient = &http.Client{
// 		Timeout:   time.Second * 10,
// 		Transport: netTransport,
// 	}
// 	req, _ := http.NewRequest(method, url, nil)
// 	for k, v := range heads {
// 		req.Header.Add(k, v)
// 	}
// 	response, err := netClient.Do(req)
// 	if err != nil {
// 		logger.Panic(err)
// 	}
// 	return response
// }

type Serverlist struct {
	Url string
}

func health(c *gin.Context) {
	c.String(http.StatusOK, "online")
}
func call(c *gin.Context) {
	log.Print(*constset.Ingressgate)
	uri := c.Request.RequestURI
	logger := logagent.Inst(c)
	logger.Info(uri)

	if strings.HasPrefix(uri, "/actuator/health") {
		c.String(http.StatusOK, "online")
	} else {
		service := c.Param("serviceid")
		env := c.Param("env")
		// body, _ := ioutil.ReadAll(c.Request.Body)
		// fmt.Println("---body/--- \r\n " + string(body))
		// env := c.Value("env").(string)
		// logger.Info(strings.Title("x-baggage-AF-region"))
		// if val, ok := c.Request.Header[strings.Title(strings.ToLower("x-baggage-AF-region"))]; ok {
		// 	region = val[0]
		// }
		// if val, ok := c.Request.Header[strings.Title(strings.ToLower("x-baggage-AF-env"))]; ok {
		// 	env = val[0]
		// }
		//should be open when closed sync
		//"fls-usedcar-trans-ms")
		// serip := "127.0.0.1"
		// serport := "8080"
		// serlocation := "http://" + serip + ":" + serport
		// serpath = "service/" + strings.ToLower(service) + serpath
		// c.Request.RemoteAddr = "user:eureka@eureka.kube.com"
		// c.Request.RequestURI = "/eureka" + strings.ReplaceAll(uri, "/eurekaagent", "")
		// c.Request.URL.User = remote.User
		//strings.Split(strings.ReplaceAll(uri, fixstr, ""), "?")[0]
		// serpath //strings.Split(strings.ReplaceAll(uri, fixstr, ""), "?")[0]
		// c.Request.URL.RawQuery = strings.ReplaceAll(c.Request.URL.RawQuery, "token=", "nekot=")
		// auth := "user:eureka"
		// basicAuth := "Bearer " + base64.StdEncoding.EncodeToString([]byte(*constset.Acltoken))
		// c.Request.Header.Add("Authorization", basicAuth)
		// if c.Request.Method == http.MethodPut {
		// }
		proxy2callee(service, env, uri, c)

		// c.Redirect(http.StatusMovedPermanently, serlocation+serpath)
	}
}

func sidecall(c *gin.Context) {
	log.Print(*constset.Ingressgate)
	uri := c.Request.RequestURI
	logger := logagent.Inst(c)
	logger.Info(uri)

	service := c.Param("serviceid")
	env := c.Param("env")
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
		logger.Print("not found")
		c.String(http.StatusNotFound, "not found")
		return
	}

	proxy2callee(service, env, uri, c)

}

func proxy2callee(service string, env string, uri string, c *gin.Context) {

	logger := logagent.Inst(c)

	logger.Info(service)

	fmt.Println("---header/--- ")
	headers := c.Request.Header
	for k, v := range headers {
		fmt.Println(k, v)
	}
	region := c.Value("region").(string)

	logger.Infof("region:%s", region)
	logger.Infof("env:%s", env)

	euservices := ragcli.EurekaApplication{}

	services := consulhelp.GetHealthService(service, c)

	if len(services) == 0 {

		euservices = ragcli.Eurekapp(service, c)

	}

	serviceinstance := acheron(env, region, services, euservices, c)

	if len(serviceinstance) == 0 {
		serviceinstance = acheron(env, "default", services, euservices, c)
		if len(serviceinstance) == 0 {
			serviceinstance = acheron("", "", services, euservices, c)
			logger.WithField("fallback", "noregion").Info(service)
		} else {
			logger.WithField("fallback", "default").Info(service)
		}
	}

	if len(serviceinstance) <= 0 {
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

	serpath := strings.ReplaceAll(uri, "/proxy/"+service+"/"+env+"/:"+*logsets.Port, "")
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

	logger.Print(c.Request)

	proxy.ServeHTTP(c.Writer, c.Request)
}

// func sidecall(c *gin.Context) {
// 	log.Print(*constset.Ingressgate)
// 	uri := c.Request.RequestURI
// 	logger := logagent.Inst(c)
// 	logger.Info(uri)

// 	service := c.Param("serviceid")
// 	env := c.Param("env")
// 	caller := c.Param("from")

// 	flag := false
// 	consulapps := ragcli.GetConsulapps(caller, env, service, c)
// 	for _, app := range consulapps {
// 		if app.Service.Service == service {
// 			flag = true
// 			break
// 		}
// 	}

// 	if !flag {
// 		logger.Print("not found")
// 		c.String(http.StatusNotFound, "not found")
// 		return
// 	}

// 	proxy2callee(service, env, uri, c)
// 	logger.Info(service)
// 	// body, _ := ioutil.ReadAll(c.Request.Body)
// 	// fmt.Println("---body/--- \r\n " + string(body))

// 	fmt.Println("---header/--- ")
// 	headers := c.Request.Header
// 	for k, v := range headers {
// 		fmt.Println(k, v)
// 	}
// 	region := c.Value("region").(string)
// 	// env := c.Value("env").(string)
// 	// logger.Info(strings.Title("x-baggage-AF-region"))
// 	// if val, ok := c.Request.Header[strings.Title(strings.ToLower("x-baggage-AF-region"))]; ok {
// 	// 	region = val[0]
// 	// }
// 	// if val, ok := c.Request.Header[strings.Title(strings.ToLower("x-baggage-AF-env"))]; ok {
// 	// 	env = val[0]
// 	// }
// 	logger.Infof("region:%s", region)
// 	logger.Infof("env:%s", env)

// 	euservices := ragcli.EurekaApplication{}

// 	services := consulhelp.GetHealthService(service, c)

// 	//should be open when closed sync
// 	if len(services) == 0 {

// 		euservices = ragcli.Eurekapp(service, c) //"fls-usedcar-trans-ms")

// 	}

// 	serviceinstance := acheron(env, region, services, euservices, c)

// 	if len(serviceinstance) == 0 {
// 		serviceinstance = acheron(env, "default", services, euservices, c)
// 		if len(serviceinstance) == 0 {
// 			serviceinstance = acheron("", "", services, euservices, c)
// 			logger.WithField("fallback", "noregion").Info(service)
// 		} else {
// 			logger.WithField("fallback", "default").Info(service)
// 		}
// 	}

// 	if len(serviceinstance) <= 0 {
// 		c.String(http.StatusNotFound, "not found")
// 		return
// 	}
// 	var index int
// 	servicemap.help(func(kvs map[string]int) (bool, interface{}) {
// 		if val, ok := kvs[service]; ok {
// 			index = val + 1
// 			kvs[service] = index
// 		} else {
// 			kvs[service] = 0
// 		}

// 		return true, 0
// 	})
// 	instanceindex := index % len(serviceinstance)
// 	logger.Printf("index:%d", index)
// 	logger.Printf("instanceindex:%d", instanceindex)
// 	instance := serviceinstance[instanceindex]

// 	logger.Printf("instance:%s", instance)

// 	// serip := "127.0.0.1"
// 	// serport := "8080"

// 	// serlocation := "http://" + serip + ":" + serport
// 	var serlocation string

// 	serpath := strings.ReplaceAll(uri, "/proxy/"+service+"/"+env+"/:"+*logsets.Port, "")
// 	if *constset.Ingressgate {
// 		serlocation = *constset.IngressHost + strings.ToLower(service)
// 		// serpath = "service/" + strings.ToLower(service) + serpath
// 		serpath = strings.Split("service/"+strings.ToLower(service)+serpath, "?")[0]
// 	} else {
// 		serlocation = strings.TrimSuffix(instance.Url, "/")
// 		serpath = strings.Split(serpath, "?")[0]
// 	}

// 	logger.Info(serlocation + serpath)

// 	remote, err := url.Parse(serlocation)
// 	if err != nil {
// 		logger.Panic(err)
// 	}

// 	proxy := httputil.NewSingleHostReverseProxy(remote)
// 	c.Request.Host = remote.Host
// 	// c.Request.RemoteAddr = "user:eureka@eureka.kube.com"
// 	// c.Request.RequestURI = "/eureka" + strings.ReplaceAll(uri, "/eurekaagent", "")
// 	// c.Request.URL.User = remote.User
// 	serpathdecoded, err := url.QueryUnescape(serpath) //strings.Split(strings.ReplaceAll(uri, fixstr, ""), "?")[0]
// 	if err != nil {
// 		logger.Panic(err)
// 	}
// 	c.Request.URL.Path = serpathdecoded // serpath //strings.Split(strings.ReplaceAll(uri, fixstr, ""), "?")[0]
// 	// c.Request.URL.RawQuery = strings.ReplaceAll(c.Request.URL.RawQuery, "token=", "nekot=")
// 	// auth := "user:eureka"
// 	// basicAuth := "Bearer " + base64.StdEncoding.EncodeToString([]byte(*constset.Acltoken))
// 	// c.Request.Header.Add("Authorization", basicAuth)
// 	// if c.Request.Method == http.MethodPut {
// 	logger.Print(c.Request)
// 	// }
// 	proxy.ServeHTTP(c.Writer, c.Request)
// 	// c.Redirect(http.StatusMovedPermanently, serlocation+serpath)

// }

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
			// if env == "" && region == "" {
			// 	serviceinstance = append(serviceinstance, Serverlist{Url: instance.HomePageUrl})
			// } else

			if envok && regionok && entryenv == env && entryregion == region {
				serviceinstance = append(serviceinstance, Serverlist{Url: instance.HomePageUrl})
			} else if env == "" {
				serviceinstance = append(serviceinstance, Serverlist{Url: instance.HomePageUrl})
			}
		}
	}
	return serviceinstance
}

// func eurekaapp(servicename string, c context.Context) EurekaApplication {
// 	logger := logagent.Inst(c)
// 	resp := tocall("GET", *regagentsets.Eu_host+"/apps/"+servicename, map[string]string{"Accept": "application/json"}, c)
// 	// resp := tocall("GET", "http://user:eureka@eureka.kube.com/eureka/apps/"+servicename, map[string]string{"Accept": "application/json"})
// 	resbody, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	var resjson = EurekaApplication{}
// 	err = json.Unmarshal(resbody, &resjson)
// 	//logger.Info(resjson.Application.Instance[0].HomePageUrl)
// 	//logger.Info(resjson.Application.Instance[0].IpAddr)
// 	//logger.Info(resjson.Application.Instance[0].Metadata)
// 	//logger.Info(resjson.Application.Instance[0].Port.Realport)
// 	if err != nil {
// 		logger.Print(err)
// 	}
// 	return resjson
// }
