package main 

 
import (
    "math"
)

type Position struct {
	Long float64
	Lat float64
}

// taken from rosetta.org code
 
func haversine(θ float64) float64 {
    return .5 * (1 - math.Cos(θ))
}
 
type pos struct {
    φ float64 // latitude, radians
    ψ float64 // longitude, radians
}

const earthRadius = float64(6372.8)
 
func degPos(lat, lon float64) pos {
    return pos{lat * math.Pi / 180, lon * math.Pi / 180}
}
 
const rEarth = 6372.8 // km
 
func hsDist(p1, p2 pos) float64 {
    return 2 * rEarth * math.Asin(math.Sqrt(haversine(p2.φ-p1.φ)+
        math.Cos(p1.φ)*math.Cos(p2.φ)*haversine(p2.ψ-p1.ψ)))
}


func DeltaDistKm(p1, p2 Position) float64 {
	return hsDist(degPos(p1.Lat, p1.Long), degPos(p2.Lat, p2.Long))
}
