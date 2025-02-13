package rual

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/theplant/luhn"
)

type Tovar struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}
type Buyback struct {
	Match       string `json:"match"`
	Reward      int    `json:"reward"`
	Reward_type string `json:"reward_type"`
}
type orda struct {
	Order string  `json:"order"`
	Goods []Tovar `json:"goods"`
}
type OrderStatus struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

var Accrualhost = "localhost:8080"
var Time429 time.Time

// func main() {

// 	cmnd := exec.Command("./acc.exe", "-d=postgres://postgres:passwordas@localhost:5432/forgo")
// 	cmnd.Start()

// 	time.Sleep(time.Second)

// 	if err := run(); err != nil {
// 		panic(err)
// 	}

// }
var marks = []Buyback{
	{Match: "Acer", Reward: 20, Reward_type: "pt"},
	{Match: "Bork", Reward: 10, Reward_type: "%"},
	{Match: "Asus", Reward: 20, Reward_type: "pt"},
	{Match: "Samsung", Reward: 25, Reward_type: "%"},
	{Match: "Apple", Reward: 35, Reward_type: "%"},
}

func LoadGood(num int, goodIdx int, price float64) error {
	ord := orda{Order: strconv.Itoa(Luhner(num)), Goods: []Tovar{
		{Description: "Smth " + marks[goodIdx].Match + " " + strconv.Itoa(num), Price: price}}}
	buyM, _ := json.Marshal(ord)
	err := poster("/api/orders", buyM)
	return err
}

func InitAccrualForTests() error {
	for _, r := range marks { // load to accrual good's type and buybacks
		buyM, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		err = poster("/api/goods", buyM)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}
	for idx := range 999 {
		err := LoadGood(idx+1, int(rand.Int63n(5)), 1000)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}
	return nil
}

func poster(postCMD string, wts []byte) error {
	httpc := resty.New() //
	httpc.SetBaseURL("http://" + Accrualhost)
	req := httpc.R().
		SetHeader("Content-Type", "application/json").
		SetBody(wts)
	_, err := req.
		SetDoNotParseResponse(false).
		Post(postCMD) //
		//	log.Printf("%s responce from server %+v  body is %s\n", postCMD, resp.StatusCode(), resp.Body())
	return err
}

func Luhner(numb int) int {
	// if luhn.Valid(numb) {
	// 	return numb
	// }
	return 10*numb + luhn.CalculateLuhn(numb)
}

// OrderStatus - {номер заказа; статус расчёта начисления; рассчитанные баллы к начислению}
func GetFromAccrual(number string) (orderStat OrderStatus, StatusCode int) {

	wait429 := time.Until(Time429) // время до разморозки
	time.Sleep(wait429)

	httpc := resty.New() //
	httpc.SetBaseURL("http://" + Accrualhost)
	getReq := httpc.R()

	resp, err := getReq.
		SetResult(&orderStat).
		SetDoNotParseResponse(false).
		SetHeader("Content-Type", "application/json").
		Get("/api/orders/" + number)
	if err != nil {
		return orderStat, http.StatusInternalServerError // 500
	}

	contentType := resp.Header().Get("Content-Type")

	if resp.StatusCode() == http.StatusTooManyRequests && contentType == "text/plain" { // http.StatusTooManyRequests 429
		delayTime := resp.Header().Get("Retry-After")
		dTime, err := strconv.Atoi(delayTime)
		if err == nil {
			var mutter sync.Mutex // установка wait429 - everybody sleeps until this
			mutter.Lock()
			Time429 = time.Now().Add(time.Duration(dTime) * time.Second)
			mutter.Unlock()
			time.Sleep(time.Duration(dTime) * time.Second)
		}
	}
	status := resp.StatusCode()
	if status == http.StatusTooManyRequests {
		status = http.StatusOK
	}
	return orderStat, status
}

/*
t:= time.Now().Add(2*time.Second)

time.Sleep(9*time.Second)
u := time.Until(t)
fmt.Println(u)

time.Sleep(u)

fmt.Println(t, "\n", time.Now())

*/
