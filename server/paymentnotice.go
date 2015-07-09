package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/intervention-engine/fhir/models"
	"gopkg.in/mgo.v2/bson"
)

func PaymentNoticeIndexHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var result []models.PaymentNotice
	c := Database.C("paymentnotices")
	iter := c.Find(nil).Limit(100).Iter()
	err := iter.All(&result)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	var paymentnoticeEntryList []models.PaymentNoticeBundleEntry
	for _, paymentnotice := range result {
		var entry models.PaymentNoticeBundleEntry
		entry.Id = paymentnotice.Id
		entry.Resource = paymentnotice
		paymentnoticeEntryList = append(paymentnoticeEntryList, entry)
	}

	var bundle models.PaymentNoticeBundle
	bundle.Id = bson.NewObjectId().Hex()
	bundle.Type = "searchset"
	bundle.Total = len(result)
	bundle.Entry = paymentnoticeEntryList

	log.Println("Setting paymentnotice search context")
	context.Set(r, "PaymentNotice", result)
	context.Set(r, "Resource", "PaymentNotice")
	context.Set(r, "Action", "search")

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(rw).Encode(bundle)
}

func LoadPaymentNotice(r *http.Request) (*models.PaymentNotice, error) {
	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	} else {
		return nil, errors.New("Invalid id")
	}

	c := Database.C("paymentnotices")
	result := models.PaymentNotice{}
	err := c.Find(bson.M{"_id": id.Hex()}).One(&result)
	if err != nil {
		return nil, err
	}

	log.Println("Setting paymentnotice read context")
	context.Set(r, "PaymentNotice", result)
	context.Set(r, "Resource", "PaymentNotice")
	return &result, nil
}

func PaymentNoticeShowHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	context.Set(r, "Action", "read")
	_, err := LoadPaymentNotice(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(rw).Encode(context.Get(r, "PaymentNotice"))
}

func PaymentNoticeCreateHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	decoder := json.NewDecoder(r.Body)
	paymentnotice := &models.PaymentNotice{}
	err := decoder.Decode(paymentnotice)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	c := Database.C("paymentnotices")
	i := bson.NewObjectId()
	paymentnotice.Id = i.Hex()
	err = c.Insert(paymentnotice)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	log.Println("Setting paymentnotice create context")
	context.Set(r, "PaymentNotice", paymentnotice)
	context.Set(r, "Resource", "PaymentNotice")
	context.Set(r, "Action", "create")

	host, err := os.Hostname()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	rw.Header().Add("Location", "http://"+host+":3001/PaymentNotice/"+i.Hex())
}

func PaymentNoticeUpdateHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	} else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	decoder := json.NewDecoder(r.Body)
	paymentnotice := &models.PaymentNotice{}
	err := decoder.Decode(paymentnotice)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	c := Database.C("paymentnotices")
	paymentnotice.Id = id.Hex()
	err = c.Update(bson.M{"_id": id.Hex()}, paymentnotice)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	log.Println("Setting paymentnotice update context")
	context.Set(r, "PaymentNotice", paymentnotice)
	context.Set(r, "Resource", "PaymentNotice")
	context.Set(r, "Action", "update")
}

func PaymentNoticeDeleteHandler(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	var id bson.ObjectId

	idString := mux.Vars(r)["id"]
	if bson.IsObjectIdHex(idString) {
		id = bson.ObjectIdHex(idString)
	} else {
		http.Error(rw, "Invalid id", http.StatusBadRequest)
	}

	c := Database.C("paymentnotices")

	err := c.Remove(bson.M{"_id": id.Hex()})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Setting paymentnotice delete context")
	context.Set(r, "PaymentNotice", id.Hex())
	context.Set(r, "Resource", "PaymentNotice")
	context.Set(r, "Action", "delete")
}
