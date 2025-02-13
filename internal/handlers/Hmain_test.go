package handlers

import (
	"log"
	"os/exec"
	"testing"
	"time"

	"github.com/Repinoid/kurs/internal/models"
	"github.com/Repinoid/kurs/internal/rual"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TstHandlers struct {
	suite.Suite
	cmnd *exec.Cmd
	t    time.Time
}

func (suite *TstHandlers) SetupSuite() {
	//var err error
	suite.t = time.Now()
	suite.cmnd = exec.Command("/acc.exe", "-d=postgres://postgres:passwordas@localhost:5432/forgo")
	err := suite.cmnd.Start()
	suite.Require().NoErrorf(err, "err %v", err)
	time.Sleep(time.Second)

	// ctx := context.Background()
	// dataBase, err := securitate.ConnectToDB(ctx) // local DB
	// suite.Require().NoErrorf(err, "err %v", err)
	// defer dataBase.DB.Close(ctx)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()
	models.Sugar = *logger.Sugar()

	log.Println("SetupTest() ---------------------")
	err = rual.InitAccrualForTests()
	suite.Require().NoErrorf(err, "err %v", err)
}

func (suite *TstHandlers) TearDownSuite() {
	err := suite.cmnd.Process.Kill()
	suite.Assert().NoErrorf(err, "err %v", err)
	log.Printf("Spent %v\n", time.Since(suite.t))
}

//	func (suite *TSuite) BeforeTest(suiteName, testName string) {
//		log.Println("BeforeTest()", suiteName, testName)
//	}
//
//	func (suite *TSuite) AfterTest(suiteName, testName string) {
//		log.Println("AfterTest()", suiteName, testName)
//	}
func TestHandlersSuite(t *testing.T) {
	log.Println("before run")
	suite.Run(t, new(TstHandlers))
}
