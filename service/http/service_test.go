package http

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/rpc"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

type testCfg struct {
	httpCfg string
	rpcCfg  string
	target  string
}

func (cfg *testCfg) Get(name string) service.Config {
	if name == ID {
		return &testCfg{target: cfg.httpCfg}
	}

	if name == rpc.ID {
		return &testCfg{target: cfg.rpcCfg}
	}
	return nil
}
func (cfg *testCfg) Unmarshal(out interface{}) error {
	return json.Unmarshal([]byte(cfg.target), out)
}

func Test_Service_NoConfig(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{}`}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusRegistered, st)
}

func Test_Service_Configure_Disable(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"enable": false,
			"address": ":8070",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusRegistered, st)
}

func Test_Service_Configure_Enable(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":8070",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusConfigured, st)
}

func Test_Service_Echo(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusConfigured, st)

	// should do nothing
	s.Stop()

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	req, err := http.NewRequest("GET", "http://localhost:6029?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))
}

func Test_Service_Middleware(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusConfigured, st)

	s.(*Service).AddMiddleware(func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path == "/halt" {
			w.WriteHeader(500)
			w.Write([]byte("halted"))
			return true
		}

		return false
	})

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)
	defer c.Stop()

	req, err := http.NewRequest("GET", "http://localhost:6029?hello=world", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	req, err = http.NewRequest("GET", "http://localhost:6029/halt", nil)
	assert.NoError(t, err)

	r, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err = ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)
	assert.Equal(t, "halted", string(b))
}

func Test_Service_Listener(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php echo pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	s, st := c.Get(ID)
	assert.NotNil(t, s)
	assert.Equal(t, service.StatusConfigured, st)

	stop := make(chan interface{})
	s.(*Service).AddListener(func(event int, ctx interface{}) {
		if event == roadrunner.EventServerStart {
			stop <- nil
		}
	})

	go func() { c.Serve() }()
	time.Sleep(time.Millisecond * 100)

	c.Stop()
	assert.True(t, true)
}

func Test_Service_Error(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php echo pipes",
				"relay": "---",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	assert.Error(t, c.Serve())
}

func Test_Service_Error2(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.NoError(t, c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php broken pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))

	assert.Error(t, c.Serve())
}

func Test_Service_Error3(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.Error(t, c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": ":6029",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers"
				"command": "php ../../php-src/tests/http/client.php broken pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))
}

func Test_Service_Error4(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)

	c := service.NewContainer(logger)
	c.Register(ID, &Service{})

	assert.Error(t, c.Init(&testCfg{httpCfg: `{
			"enable": true,
			"address": "----",
			"maxRequest": 1024,
			"uploads": {
				"dir": ` + tmpDir() + `,
				"forbid": []
			},
			"workers":{
				"command": "php ../../php-src/tests/http/client.php broken pipes",
				"relay": "pipes",
				"pool": {
					"numWorkers": 1, 
					"allocateTimeout": 10000000,
					"destroyTimeout": 10000000 
				}
			}
	}`}))
}

func tmpDir() string {
	p := os.TempDir()
	r, _ := json.Marshal(p)

	return string(r)
}
