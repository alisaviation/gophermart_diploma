package rual

// Basic imports
import (
	"log"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TSuite struct {
	suite.Suite
	cmnd *exec.Cmd
}

func (suite *TSuite) SetupSuite() {
	suite.cmnd = exec.Command("/acc.exe", "-d=postgres://postgres:passwordas@localhost:5432/forgo")
	err := suite.cmnd.Start()
	require.NoErrorf(suite.T(), err, "err %v", err)
	time.Sleep(time.Second)

	err = InitAccrualForTests()
	suite.Require().NoErrorf(err, "err %v", err)
	log.Println("SetupTest() ---------------------")
}

func (suite *TSuite) TearDownSuite() {
	err := suite.cmnd.Process.Kill()
	assert.NoErrorf(suite.T(), err, "err %v", err)
}

//	func (suite *TSuite) BeforeTest(suiteName, testName string) {
//		log.Println("BeforeTest()", suiteName, testName)
//	}
//
//	func (suite *TSuite) AfterTest(suiteName, testName string) {
//		log.Println("AfterTest()", suiteName, testName)
//	}
func TestAccrualSuite(t *testing.T) {
	log.Println("before run")
	suite.Run(t, new(TSuite))
}
