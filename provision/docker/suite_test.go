package docker

import (
	"testing"

	dtesting "github.com/fsouza/go-dockerclient/testing"
	"github.com/virtengine/libgo/safe"
	"github.com/virtengine/vertice/provision"
	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	collName       string
	imageCollName  string
	repoNamespace  string
	deployCmd      string
	runBin         string
	runArgs        string
	port           string
	sshUser        string
	server         *dtesting.DockerServer
	extraServer    *dtesting.DockerServer
	targetRecover  []string
	oldProvisioner provision.Provisioner
	p              *dockerProvisioner
	logBuf         *safe.Buffer
}

var _ = check.Suite(&S{})

/*
func (s *S) SetUpSuite(c *check.C) {
	s.collName = "docker_unit"
	s.imageCollName = "docker_image"
	s.repoNamespace = "github.com/virtengine/vertice"
	s.sshUser = "root"
	s.port = "8888"
	config.Set("database:url", "127.0.0.1:27017?maxPoolSize=100")
	config.Set("database:name", "docker_provision_tests_s")
	config.Set("docker:repository-namespace", s.repoNamespace)
	config.Set("docker:router", "fake")
	config.Set("docker:collection", s.collName)
	config.Set("docker:deploy-cmd", "/var/lib/"github.com/virtengine/vertice/deploy")
	config.Set("docker:run-cmd:bin", "/usr/local/bin/circusd /etc/circus/circus.ini")
	config.Set("docker:run-cmd:port", s.port)
	config.Set("docker:user", s.sshUser)
	config.Set("docker:cluster:mongo-url", "127.0.0.1:27017")
	config.Set("docker:cluster:mongo-database", "docker_provision_tests_cluster_stor")
	config.Set("queue:mongo-url", "127.0.0.1:27017")
	config.Set("queue:mongo-database", "queue_provision_docker_tests")
	config.Set("queue:mongo-polling-interval", 0.01)
	config.Set("routers:fake:type", "fake")
	config.Set("repo-manager", "fake")
	config.Set("admin-team", "admin")
	config.Set("docker:registry-max-try", 1)
	config.Set("auth:hash-cost", bcrypt.MinCost)
	s.deployCmd = "/var/lib/"github.com/virtengine/vertice/deploy"
	s.runBin = "/usr/local/bin/circusd"
	s.runArgs = "/etc/circus/circus.ini"
	os.Setenv("github.com/virtengine/vertice_TARGET", "http://localhost")
	s.oldProvisioner = app.Provisioner
	var err error
	s.storage, err = db.Conn()
	c.Assert(err, check.IsNil)
	clusterDbUrl, _ := config.GetString("docker:cluster:mongo-url")
	s.clusterSess, err = mgo.Dial(clusterDbUrl)
	c.Assert(err, check.IsNil)
	err = dbtest.ClearAllCollections(s.storage.Apps().Database)
	c.Assert(err, check.IsNil)
	repositorytest.Reset()
	s.user = &auth.User{Email: "myadmin@arrakis.com", Password: "123456", Quota: quota.Unlimited}
	nativeScheme := auth.ManagedScheme(native.NativeScheme{})
	app.AuthScheme = nativeScheme
	_, err = nativeScheme.Create(s.user)
	c.Assert(err, check.IsNil)
	s.team = &auth.Team{Name: "admin", Users: []string{s.user.Email}}
	c.Assert(err, check.IsNil)
	err = s.storage.Teams().Insert(s.team)
	c.Assert(err, check.IsNil)
	s.token, err = nativeScheme.Login(map[string]string{"email": s.user.Email, "password": "123456"})
	c.Assert(err, check.IsNil)
}

func (s *S) SetUpTest(c *check.C) {
	config.Set("docker:api-timeout", 2)
	iaas.ResetAll()
	repositorytest.Reset()
	queue.ResetQueue()
	s.p = &dockerProvisioner{storage: &cluster.MapStorage{}}
	err := s.p.Initialize()
	c.Assert(err, check.IsNil)
	queue.ResetQueue()
	app.Provisioner = s.p
	s.server, err = dtesting.NewServer("127.0.0.1:0", nil, nil)
	c.Assert(err, check.IsNil)
	s.p.cluster, err = cluster.New(nil, s.p.storage,
		cluster.Node{Address: s.server.URL(), Metadata: map[string]string{"pool": "test-default"}},
	)
	c.Assert(err, check.IsNil)
	mainDockerProvisioner = s.p
	err = dbtest.ClearAllCollectionsExcept(s.storage.Apps().Database, []string{"users", "tokens", "teams"})
	c.Assert(err, check.IsNil)
	err = clearClusterStorage(s.clusterSess)
	c.Assert(err, check.IsNil)
	routertest.FakeRouter.Reset()
	opts := provision.AddPoolOptions{Name: "test-default", Default: true}
	err = provision.AddPool(opts)
	c.Assert(err, check.IsNil)
	s.storage.Tokens().Remove(bson.M{"appname": bson.M{"$ne": ""}})
	s.logBuf = safe.NewBuffer(nil)
	log.SetLogger(log.NewWriterLogger(s.logBuf, true))
}

func (s *S) TearDownTest(c *check.C) {
	log.SetLogger(nil)
	s.server.Stop()
	if s.extraServer != nil {
		s.extraServer.Stop()
		s.extraServer = nil
	}
}

func (s *S) TearDownSuite(c *check.C) {
	s.clusterSess.Close()
	s.storage.Close()
	os.Unsetenv("github.com/virtengine/vertice_TARGET")
	app.Provisioner = s.oldProvisioner
}

func clearClusterStorage(sess *mgo.Session) error {
	clusterDbName, _ := config.GetString("docker:cluster:mongo-database")
	return dbtest.ClearAllCollections(sess.DB(clusterDbName))
}

func (s *S) startMultipleServersCluster() (*dockerProvisioner, error) {
	var err error
	s.extraServer, err = dtesting.NewServer("localhost:0", nil, nil)
	if err != nil {
		return nil, err
	}
	otherUrl := strings.Replace(s.extraServer.URL(), "127.0.0.1", "localhost", 1)
	var p dockerProvisioner
	err = p.Initialize()
	if err != nil {
		return nil, err
	}
	p.storage = &cluster.MapStorage{}
	p.cluster, err = cluster.New(nil, p.storage,
		cluster.Node{Address: s.server.URL(), Metadata: map[string]string{"pool": "test-default"}},
		cluster.Node{Address: otherUrl, Metadata: map[string]string{"pool": "test-default"}},
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *S) startMultipleServersClusterSeggregated() (*dockerProvisioner, error) {
	var err error
	s.extraServer, err = dtesting.NewServer("localhost:0", nil, nil)
	if err != nil {
		return nil, err
	}
	otherUrl := strings.Replace(s.extraServer.URL(), "127.0.0.1", "localhost", 1)
	var p dockerProvisioner
	err = p.Initialize()
	if err != nil {
		return nil, err
	}
	p.storage = &cluster.MapStorage{}
	sched := segregatedScheduler{provisioner: &p}
	p.cluster, err = cluster.New(&sched, p.storage,
		cluster.Node{Address: s.server.URL(), Metadata: map[string]string{"pool": "pool1"}},
		cluster.Node{Address: otherUrl, Metadata: map[string]string{"pool": "pool2"}},
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *S) addServiceInstance(c *check.C, appName string, units []string, fn http.HandlerFunc) func() {
	ts := httptest.NewServer(fn)
	ret := func() {
		ts.Close()
		s.storage.Services().Remove(bson.M{"_id": "mysql"})
		s.storage.ServiceInstances().Remove(bson.M{"_id": "my-mysql"})
	}
	srvc := service.Service{Name: "mysql", Endpoint: map[string]string{"production": ts.URL}}
	err := srvc.Create()
	c.Assert(err, check.IsNil)
	instance := service.ServiceInstance{Name: "my-mysql", ServiceName: "mysql", Teams: []string{}, Units: units, Apps: []string{appName}}
	err = instance.Create()
	c.Assert(err, check.IsNil)
	return ret
}

func startTestRepositoryServer() func() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	repoUrl := strings.Replace(server.URL, "http://", "", 1)
	config.Set("docker:registry", repoUrl)
	return func() {
		config.Unset("docker:registry")
		server.Close()
	}
}

type unitSlice []provision.Unit

func (s unitSlice) Len() int {
	return len(s)
}

func (s unitSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s unitSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortUnits(units []provision.Unit) {
	sort.Sort(unitSlice(units))
}
*/
