package main

import (
	"time"
	"log"
	"encoding/json"
	"os"
)



type Geo struct {
	Type string `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type Prop struct {
	Name userIDType `json:"name"`
	Duration int64  `json:"duration_sec"`
	Date time.Time `json:"date_UTC"`
}

type Feature struct {
	Type string `json:"type"`
	Geometry Geo `json:"geometry"`
	Properties Prop `json:"properties"`
}

type FeatureList struct {
	Type string `json:"type"`
	Features []Feature `json:"features"`
}

// slices are pointers anyway so ok to pass it here
func WriteJsonReport(filepath string, features []Feature) {
	list := FeatureList{Type: "FeatureCollection", Features: features}

	fp, err := os.Create(filepath)
	if err != nil {
		log.Fatalln("error create json file")
		return
	}
	defer fp.Close()

	enc := json.NewEncoder(fp)

	err = enc.Encode(&list)
	if err != nil {
		log.Fatalln("error encoding json")
		return
	}
}

func AddFeature(list *[]Feature, pos Position, name userIDType, duration int64, timestamp int64) {
	p := Prop{Name: name, Duration: duration, Date: time.Unix(timestamp,0).UTC()}
	g := Geo{Type: "Point", Coordinates: []float64{pos.Long, pos.Lat}}
	f := Feature{Type: "Feature", Geometry: g, Properties: p}

	*list = append(*list,f)
}
