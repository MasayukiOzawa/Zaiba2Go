package main

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func getMeasurement(sql *string, fields interface{}) {

	// クエリ実行
	rows, err := db.Queryx(*sql)
	if err != nil {
		fmt.Println(fmt.Sprintf("[%v] : SQL Execution Error : %s\n", time.Now().Format(timeFormat), err.Error()))
		wg.Done()
		return
	}
	buf := make([]byte, 0)

	for rows.Next() {
		var tagValue []byte
		var fieldValue []byte
		var measurement string

		rows.StructScan(fields)

		rt, rv := reflect.TypeOf(fields).Elem(), reflect.ValueOf(fields).Elem()

		for i := 0; i < rt.NumField(); i++ {
			fi := rt.Field(i)
			switch fi.Tag.Get("type") {
			case "measurement":
				measurement = strings.Replace(rv.Field(i).Interface().(string), " ", "\\ ", -1)
			case "tag":
				if tagValue == nil {
					tagValue = append(tagValue, (fi.Tag.Get("db") + "=" + strings.Replace(rv.Field(i).Interface().(string), " ", "\\ ", -1))...)
				} else {
					tagValue = append(tagValue, ("," + fi.Tag.Get("db") + "=" + strings.Replace(rv.Field(i).Interface().(string), " ", "\\ ", -1))...)
				}
			case "field":
				if fieldValue == nil {
					fieldValue = append(fieldValue, (strings.Replace(fi.Tag.Get("db"), " ", "\\ ", -1) + "=" + strconv.FormatFloat(rv.Field(i).Interface().(float64), 'f', 2, 64))...)
				} else {
					fieldValue = append(fieldValue, ("," + strings.Replace(fi.Tag.Get("db"), " ", "\\ ", -1) + "=" + strconv.FormatFloat(rv.Field(i).Interface().(float64), 'f', 2, 64))...)
				}
			}
		}
		buf = append(buf, fmt.Sprintf("%s,%s,%s %s %d\n",
			measurement,
			applicationIntent,
			string(tagValue),
			string(fieldValue),
			time.Now().UnixNano(),
		)...)
	}

	req, err := http.NewRequest("POST", influxdbURI, bytes.NewBuffer(buf))
	// fmt.Println(string(buf))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("[%v] : Response Status [%s]\n", time.Now().Format(timeFormat), resp.Status)
	}
	wg.Done()
}
