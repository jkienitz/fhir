package models2

import (
	"time"
	"fmt"
	"math"
	"strings"

	"github.com/eug48/fhir/utils"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"
)

type refsMap map[string]string

const Gofhir__strNum = "__strNum"
const Gofhir__strDate = "__strDate"
const Gofhir__num = "__num"
const Gofhir__from = "__from"
const Gofhir__to = "__to"


// Converts a FHIR JSON Resource into BSON for storage in MongoDB
// Does several transformations:
//   - re-writes references (for transactions)
//   - converts id to _id and puts first (_id converted to __id)
//   - converts extensions from { url, value } to { url: { value } } to enable better MongoDB queries
//   - converts decimal numbers to { __from, __to, __num, __strNum } for FHIR conformance
//   - converts dates to { __from, __to, __strDate } for FHIR conformance
//   - optionally encrypts certain fields
func ConvertJsonToGoFhirBSON(jsonBytes []byte, whatToEncrypt WhatToEncrypt, transformReferencesMap map[string]string) (out bson.D, err error) {

	debug("=== ConvertJsonToGoFhirBSON ===")

	bsonRoot := make([]bson.DocElem, 0, 8)
	refsMap := refsMap(transformReferencesMap)
	resourceType, err := jsonparser.GetString(jsonBytes, "resourceType")
	if err != nil {
		err = errors.Wrapf(err, "ConvertJsonToGoFhirBSON: failed to get resourceType")
	}

	if err == nil {
		pos := positionInfo{pathHere: resourceType, element: resourceType}
		err = jsonparser.ObjectEach(jsonBytes, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			err4 := addToBSONdoc(&bsonRoot, pos, key, value, dataType, offset, refsMap)
			return err4
		})
	}

	if err == nil {
		err = encryptBSON(&bsonRoot, resourceType, whatToEncrypt)
		if err != nil {
			err = errors.Wrapf(err, "encryptBSON failed")
		}
	}

	if err == nil {
		return bsonRoot, nil
	} else {
		return nil, err
	}
}

func addToBSONdoc(output *[]bson.DocElem, pos positionInfo, key []byte, value []byte, dataType jsonparser.ValueType, offset int, refsMap refsMap) error {
	strKey := string(key)
	nextPos := pos.downTo(strKey, value)

	valueBson, err := convertValue(nextPos, value, dataType, refsMap)
	if err != nil {
		return errors.Wrapf(err, "object convertValue failed at %s", nextPos.pathHere)
	}

	bsonKey := strKey
	putFirst := false
	if strKey == "id" {
		bsonKey = "_id"
		putFirst = true
	} else if strKey == "_id" {
		bsonKey = "__id"
	}
	elem := bson.DocElem{Name: bsonKey, Value: valueBson}
	if putFirst {
		*output = append([]bson.DocElem{elem}, (*output)...)
	} else {
		*output = append(*output, elem)
	}

	if pos.atReference() {

		// transform reference during transactions
		reference := string(value)
		transformedReferenced, found := refsMap[reference]
		if found {
			reference = transformedReferenced
			(*output)[len(*output)-1].Value = transformedReferenced
		}

		// add reference__id, reference__type and reference__external fields
		splitURL := strings.Split(reference, "/")
		if len(splitURL) >= 2 {
			// TODO: validate?
			referenceID := splitURL[len(splitURL)-1]
			typeStr := splitURL[len(splitURL)-2]
			*output = append(*output, bson.DocElem{Name: "reference__id", Value: referenceID})
			*output = append(*output, bson.DocElem{Name: "reference__type", Value: typeStr})
		}
		external := strings.HasPrefix(reference, "http")
		*output = append(*output, bson.DocElem{Name: "reference__external", Value: external})
	}

	return nil
}

func addToBSONarray(output *[]interface{}, pos positionInfo, value []byte, dataType jsonparser.ValueType, offset int, refsMap refsMap) error {

	valueBson, err := convertValue(pos.intoArray(value), value, dataType, refsMap)
	if err != nil {
		return errors.Wrapf(err, "array convertValue failed at %s", pos.pathHere)
	}
	*output = append(*output, valueBson)
	return nil
}

func convertValue(pos positionInfo, value []byte, dataType jsonparser.ValueType, refsMap refsMap) (out interface{}, err error) {

	switch dataType {
	case jsonparser.Object:
		subDoc := make([]bson.DocElem, 0, 4)

		err = jsonparser.ObjectEach(value, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			err2 := addToBSONdoc(&subDoc, pos, key, value, dataType, offset, refsMap)
			// fmt.Printf("Key: '%s'\n Value: '%s'\n Type: %s\n", string(key), string(value), dataType)
			return err2
		})
		if err != nil {
			return nil, errors.Wrapf(err, "ObjectEach failed at %s", pos.pathHere)
		}

		return subDoc, nil

	case jsonparser.Array:
		array := make([]interface{}, 0, 4)

		if pos.atExtension() {
			err = convertExtensionArray(&array, value, pos, refsMap)
			if err != nil {
				err = errors.Wrap(err, "convertExtensionArray failed")
			}
			return array, err
		}

		var err5 error
		_, err := jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err3 error) {
			if err3 == nil && err5 == nil {
				err5 = addToBSONarray(&array, pos, value, dataType, offset, refsMap)
			}
		})
		if err != nil {
			return nil, errors.Wrapf(err, "ArrayEach failed at %s", pos.pathHere)
		}
		if err5 != nil {
			return nil, errors.Wrapf(err5, "ArrayEach.addToBSONarray failed at %s", pos.pathHere)
		}

		return array, nil

	case jsonparser.String:

		if pos.atDate() {
			out, err = convertDateValue(value, pos)
			if err != nil {
				err = errors.Wrap(err, "convertDateValue failed")
			}
			return
		} else if pos.atInstant() {
			out, err = convertInstant(value, pos)
			if err != nil {
				err = errors.Wrap(err, "convertInstant failed")
			}
			return
		} else {
			unescaped, err := jsonparser.Unescape(value, nil)
			if err != nil {
				return nil, errors.Wrap(err, "jsonparser.Unescape failed")
			}
			return string(unescaped), nil
		}

	case jsonparser.Null:
		return nil, nil

	case jsonparser.Boolean:
		boo, err := jsonparser.GetBoolean(value)
		if err != nil {
			return nil, errors.Wrap(err, "GetBoolean failed")
		}
		return boo, nil

	case jsonparser.Number:

		elem, err := convertNumberValue(value, pos)
		if err != nil {
			return nil, err
		}

		return elem, nil

	default:
		panic(fmt.Errorf("unhandled json datatype: %d", dataType))
	}

}

func convertExtensionArray(output *[]interface{}, jsonBytes []byte, pos positionInfo, refsMap refsMap) (err error) {
	debug("convertExtensionArray started")
	var funcErr error
	_, err = jsonparser.ArrayEach(jsonBytes, func(origExtensonBytes []byte, dataType jsonparser.ValueType, offset int, err3 error) {
		if err3 == nil && funcErr == nil {

			if dataType == jsonparser.Null {
				*output = append(*output, nil)
				debug("convertExtensionArray: added nil")
				return
			}

			if dataType != jsonparser.Object {
				funcErr = fmt.Errorf("getExtensionArray: element is not an object at %s (%d)", pos.pathHere, dataType)
				return
			}

			// promote url to a key to enable searching in Mongodb
			var url string
			url, funcErr = jsonparser.GetString(origExtensonBytes, "url")
			if funcErr != nil {
				funcErr = errors.Wrap(funcErr, "failed to get url")
				debug("convertExtensionArray: failed to get url: %v", funcErr)
				return
			}

			newChildExtensionObj := make([]bson.DocElem, 0, 4)
			funcErr = jsonparser.ObjectEach(origExtensonBytes, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
				strKey := string(key)
				if strKey == "url" {
					debug("convertExtensionArray: child object: %s (skipped)", strKey)
					return nil
				} else {
					debug("convertExtensionArray: child object: %s", strKey)
				}
				err4 := addToBSONdoc(&newChildExtensionObj, pos, key, value, dataType, offset, refsMap)
				return err4
			})
			if funcErr != nil {
				return
			}

			newParentExtensionObj := []bson.DocElem{
				bson.DocElem{Name: url, Value: newChildExtensionObj},
			}

			*output = append(*output, newParentExtensionObj)
		}
		// fmt.Printf("Key: '%s'\n Value: '%s'\n Type: %s\n", string(key), string(value), dataType)
	})

	debug("convertExtensionArray finished: errors %v %v", funcErr, err)

	if funcErr != nil {
		return funcErr
	}
	if err != nil {
		return err
	}
	return nil
}

func convertInstant(jsonBytes []byte, pos positionInfo) (elem interface{}, err error) {
	var t time.Time
	t, err = time.Parse(time.RFC3339, string(jsonBytes))
	if err == nil {
		elem = t
	}
	return
}

func convertDateValue(jsonBytes []byte, pos positionInfo) (elem interface{}, err error) {

	stringForm := string(jsonBytes)
	date, err := utils.ParseDate(stringForm)
	if err != nil {
		return nil, errors.Wrap(err, "ParseDate failed")
	}

	elem = []bson.DocElem{
		bson.DocElem{Name: Gofhir__from, Value: date.RangeLowIncl()},
		bson.DocElem{Name: Gofhir__to, Value: date.RangeHighExcl()},
		bson.DocElem{Name: Gofhir__strDate, Value: stringForm},
	}
	return
}

// FHIR requires a decimal's string representation to be preserved exactly
// so we store a string representation of decimals
func convertNumberValue(jsonBytes []byte, pos positionInfo) (elem interface{}, err error) {

	stringForm := string(jsonBytes)

	if pos.atDecimal() {

		var numValue interface{}
		// If number has a dot store as a float, otherwise int
		if strings.Contains(stringForm, ".") {
			numValue, err = jsonparser.GetFloat(jsonBytes)
		} else {
			numValue, err = jsonparser.GetInt(jsonBytes)
			// TODO: looks like jsonparser.parseInt overflows??
		}
		if err != nil {
			return nil, errors.Wrapf(err, "GetFloat or GetInt failed for %s at %s", stringForm, pos.pathHere)
		}

		num := utils.ParseNumber(stringForm)
		numFrom, _ := num.RangeLowIncl().Float64()
		numTo, _ := num.RangeHighExcl().Float64()

		elem = []bson.DocElem{
			// TODO: set ranges properly based on precision
			bson.DocElem{Name: Gofhir__from, Value: numFrom},
			bson.DocElem{Name: Gofhir__to, Value: numTo},
			bson.DocElem{Name: Gofhir__num, Value: numValue},
			bson.DocElem{Name: Gofhir__strNum, Value: stringForm},
		}
	} else {
		if strings.Contains(stringForm, ".") {
			return nil, errors.Wrapf(err, "non-decimal numer has a decimal point (%s)", stringForm, pos.pathHere)
		}

		elemInt64, err := jsonparser.GetInt(jsonBytes)
		if err != nil {
			return nil, errors.Wrapf(err, "GetInt failed for %s at %s", stringForm, pos.pathHere)
		}
		if elemInt64 >= math.MinInt32 && elemInt64 <= math.MaxInt32 {
			elemInt32 := int32(elemInt64)
			elem = &elemInt32
		} else {
			elem = &elemInt64
		}
	}

	return
}
