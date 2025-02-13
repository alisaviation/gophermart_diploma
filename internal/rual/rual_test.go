package rual

// Basic imports
import (
	"log"
	"net/http"
	"strconv"

	"github.com/stretchr/testify/assert"
)

func (suite *TSuite) Test01Setup() {


	log.Println("testexample5")

}
func (suite *TSuite) Test02GetFromAccrual() {

	for idx := range marks {
		Order := strconv.Itoa(Luhner(idx))
		orderStat, StatusCode := GetFromAccrual(Order)
	//	assert.NoErrorf(suite.T(), err, "err %v", err)
		assert.Equal(suite.T(), http.StatusOK, StatusCode)
		log.Println(orderStat)

	}
	log.Println("TestGetFromAccrual")

}
