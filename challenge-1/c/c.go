package main

import (
	"encoding/json"

	"github.com/hashicorp/go-retryablehttp"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Coupon struct {
	Code string
}

type Coupons struct {
	Coupon []Coupon
}

func (c Coupons) Check(code string) string {
	for _, item := range c.Coupon {
		if code == item.Code {
			return "valid"
		}
	}
	return "invalid"
}

type Result struct {
	Status string
}

var coupons Coupons

func main() {
	coupon1 := Coupon{
		Code: "abc",
	}

	coupons.Coupon = append(coupons.Coupon, coupon1)

	coupon2 := Coupon{
		Code: "abcd",
	}

	coupons.Coupon = append(coupons.Coupon, coupon2)

	http.HandleFunc("/", home)
	http.ListenAndServe(":9092", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	coupon := r.PostFormValue("coupon")
	valid := coupons.Check(coupon)

	resultCoupon := makeHttpCall("http://localhost:9093", coupon)

	result := Result{Status: valid}

	if resultCoupon.Status == "REMOVED" {
		result.Status = "REMOVED"
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		log.Fatal("Error converting json")
	}

	fmt.Fprintf(w, string(jsonResult))

}

func makeHttpCall(urlMicroservice string, coupon string) Result {

	values := url.Values{}
	values.Add("coupon", coupon)

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5

	res, err := retryClient.PostForm(urlMicroservice, values)
	if err != nil {
		result := Result{Status: "Servidor fora do ar!"}
		return result
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal("Error processing result")
	}

	result := Result{}

	json.Unmarshal(data, &result)

	return result
}
