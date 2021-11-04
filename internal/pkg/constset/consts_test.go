package constset

import (
	"os"
	"testing"
)

// var args []string

func setup() {
	// plaintext = "123"
	// cryptedHexText = "1bda1896724a4521cfb7f38646824197929cd1"
}

func teardown() {

}

// func Test_Cases(t *testing.T) {
// 	// <setup code>
// 	setup()

// 	t.Run("StartupInit=StartupInit", Test_StartupInit)
// 	// t.Run("Decrypt=hex2str", Test_Decryptbyhex2str)
// 	// t.Run("Decrypt=hex2byte", Test_Decryptbyhex)
// 	// t.Run("Write=ExistedFile", Test_Write)
// 	// t.Run("Write=WhiteToPath", Test_WhiteToPath)
// 	// <tear-down code>
// 	teardown()
// }

// func Test_StartupInit(t *testing.T) {

// 	//test for default
// 	testport := "8080"
// 	testconsul_host := "http://dev-consul:8500"
// 	testacltoken := ""
// 	testenvset := []string{"sit", "uat", "prod"}
// 	args := []string{}
// 	StartupInit(args)

// 	var f0 = func() {

// 		assert.Equal(t, testport, Port)
// 		assert.Equal(t, testconsul_host, Consul_host)
// 		assert.Equal(t, testacltoken, Acltoken)

// 		// if strings.Compare(testport, Port) != 0 {
// 		// 	t.Fatalf("Test_EncryptStr2hex failed! port should be:%s, get:%s", testport, Port)
// 		// }
// 		// if strings.Compare(testconsul_host, Consul_host) != 0 {
// 		// 	t.Fatalf("Test_EncryptStr2hex failed! consul_host should be:%s, get:%s", testconsul_host, Consul_host)
// 		// }
// 		// if strings.Compare(testacltoken, Acltoken) != 0 {
// 		// 	t.Fatalf("Test_EncryptStr2hex failed! acltoken should be:%s, get:%s", testacltoken, Acltoken)
// 		// }

// 		// var flag = true
// 		// for _, env := range testenvset {
// 		// 	for _, Env := range EnvSet {
// 		// 		if strings.Compare(env, Env) == 0 {
// 		// 			flag = true
// 		// 			break
// 		// 		}
// 		// 		flag = false
// 		// 	}
// 		// }
// 		// if !flag {
// 		// 	t.Fatalf("Test_EncryptStr2hex failed! envset should be:%s, get:%s", testenvset, EnvSet)
// 		// }
// 		assert.Equal(t, testenvset, EnvSet)
// 		log.Printf("Test_StartupInit result is:\nport:%s; consul_host:%s; acltoken:%s; envset:%s", Port, Consul_host, Acltoken, EnvSet)
// 	}

// 	f0()

// 	//test for args
// 	testport = "1234"
// 	testconsul_host = "http://123:8500"
// 	testacltoken = "ddd"
// 	testenvset = []string{"test", "uat"}
// 	args = []string{port + "=" + testport, consul_host + "=" + testconsul_host, acltoken + "=" + testacltoken, envset + "=" + "test,uat"}
// 	StartupInit(args)
// 	f0()
// }

func TestMain(m *testing.M) {
	setup()
	// constset.StartupInit()
	// sendconfig2consul()
	// configgen.Getconfig = getTestConfig

	exitCode := m.Run()
	teardown()
	// // 退出
	os.Exit(exitCode)
}
