package main

import (
	"time"
	"math"
	"fmt"
	"encoding/json"
	"io"
	"log"
	"os"
	"flag"
)

// {"Message":{"MessageID":18,"RepeatIndicator":0,"UserID":416004341,"Valid":true,"Spare1":0,"Sog":6.8,"PositionAccuracy":false,
// "Longitude":171.32811666666666,"Latitude":-7.578556666666667,
// "Cog":272,"TrueHeading":511,"Timestamp":58,"Spare2":0,"ClassBUnit":true,"ClassBDisplay":false,"ClassBDsc":true,"ClassBBand":false,
// "ClassBMsg22":false,"AssignedMode":false,"Raim":false,"CommunicationStateIsItdma":true,"CommunicationState":393222},
// "UTCTimeStamp":1588636800}

type userIDType int64

type Msg struct {
	MessageID int
	UserID userIDType
	Longitude float64
	Latitude float64
}

type Row struct {
	Message Msg
	UTCTimeStamp int64
}


type VesselState struct {
	UserID userIDType
	Pos [2]Position // Position N-1 and N
	Timestamp [2]int64 // Timestamp N-1,N
	DTime int64 //secs
	DDistance float32 // meters
	Speed float64 // in knots
	IsStopped bool // true if less than 1 kn of speed
	ZeroSpeedTimeSecs int64 // total cumulated continuous stopped time less than 1kn time
	RowCount int // records per vessel
}

type VesselMap map[userIDType]*VesselState

var (
	inFile = flag.String("in", "", "-in inputfile.json")
	outFile = flag.String("out", "", "-out outpufile.json")
	g_msg_types = [] int {1, 2, 3, 18, 19, 27}
	vesselMap = VesselMap{}
	features = []Feature{}
	vesselCount = 0
	stopCount = 0

)

func (r Row) is_relevant() bool { // filter message types
	for _, t := range g_msg_types {
        if t == r.Message.MessageID {
			return true
		}
    }
	return false
}

func (s VesselState) getDeltaDist() float32{
	ddist := DeltaDistKm(s.Pos[0], s.Pos[1])
	// fmt.Println("delta Dist m %f", ddist)
	return float32(ddist)
}

func (s VesselState) getDeltaTime() int64{
	t1 := time.Unix(s.Timestamp[1], 0)
	t0 := time.Unix(s.Timestamp[0], 0)
	diff := t1.Sub(t0)
	dtime := diff.Seconds()
	// fmt.Println("delta time secs %f", dtime)
	return int64(dtime)
}

func getSpeedKnots(dist float32, time int64) float64{
	if time == 0.0 {
		return 0.0
	}
	speed := math.Abs( (float64(dist) / float64(time)) ) * 1.944 // 1m/s == 1.944 knots
	return speed
}

func (s VesselState) printPosition() { // called when finding a stopped position
	fmt.Printf("Found stop #%d : ID %d", stopCount, s.UserID)
	fmt.Printf(" Date %s", time.Unix(s.Timestamp[0],0).UTC())
	fmt.Printf(" row-count %d", s.RowCount)
	fmt.Printf(" last delta-time %ds, total zero speed time = %ds\n", s.DTime, s.ZeroSpeedTimeSecs)

	// fmt.Printf(" timestamp [0] %d [1] %d ", s.Timestamp[0], s.Timestamp[1])
	// fmt.Printf(" dist %f m", s.DDistance)
	// fmt.Printf(" Speed %f kn\n\n", s.Speed)
}

// init new vessel state object
func NewVessel(r Row) *VesselState {
	posZero := Position{Lat: r.Message.Latitude, Long: r.Message.Longitude}
	v := &VesselState{ 
		UserID: r.Message.UserID, 
		Pos: [2]Position{posZero, {0.0, 0.0}},
		DDistance: 0.0,
		DTime: 0,
		Timestamp: [2]int64{r.UTCTimeStamp, r.UTCTimeStamp},
		Speed: 0.0,
		IsStopped: false,
		ZeroSpeedTimeSecs: 0,
		RowCount: 0,
	}

	return v
}


func (s VesselState) reportStoppedPosition() {
	// fmt.Printf(" time = %d, pos = %v \n", s.ZeroSpeedTimeSecs, s.Pos[0])

	s.printPosition()
	stopCount++
	AddFeature(&features, s.Pos[0], s.UserID, s.ZeroSpeedTimeSecs, s.Timestamp[0])
}

// accumulates stopped time and gets the last record/position of the found stopped 
// continuous sequence of row for which the speed is less than 1 kn
func  updateVesselState(r Row, s *VesselState) {
	// fmt.Println(s)

	s.Pos[1] = Position{Lat: r.Message.Latitude, Long: r.Message.Longitude} // position [1] is position N,  [0] is N -1
	s.Timestamp[1] = r.UTCTimeStamp
	// calculate delta time, delta distance and then current speed
	s.DDistance = s.getDeltaDist() * 1000 // meters
	s.DTime = s.getDeltaTime() //secs
	s.Speed = getSpeedKnots(s.DDistance, s.DTime) // kn


	if s.Speed < 1 { //while stopped already or stopping accumulating time stopped
		
		s.ZeroSpeedTimeSecs += s.DTime // time accumulator
		s.IsStopped = true // flag considered stopped on cuurent record, useful to close up the report.
	} else {
		
		if s.ZeroSpeedTimeSecs >= 3600 {
			// end of stop but did stop for more than 1hr, so produce a stop position report
			s.reportStoppedPosition()
		}

		// end of stop moving again, reset state
		s.ZeroSpeedTimeSecs = 0
		s.IsStopped = false
	}

	s.Pos[0] = s.Pos[1]
	s.Timestamp[0] = s.Timestamp[1]
}

func (r Row) processRow() { // 1 record at at time
	vessel, ok := vesselMap[r.Message.UserID]
	if !ok { // new vessel found
		new := NewVessel(r)
		vesselMap[new.UserID] = new
		vesselCount++
		fmt.Printf("New vessel found %d, current vessel count = %d...\n", new.UserID, vesselCount)

	} else {
		//update existing vessel
		updateVesselState(r, vessel)

		// iter next
		vessel.RowCount++
	}
}

// The mysterious inteverview question : this is the answer.
// when finished processing the input file, reporting all vessels of the map blocked in the stopped state.
func closeUpReporting() {
	for _, s := range vesselMap {
		if s.IsStopped && s.ZeroSpeedTimeSecs >= 3600 {
			s.reportStoppedPosition()
		}
	}

	// produce the geojson report
	WriteJsonReport(*outFile, features)
}


func main() {

	flag.Parse()

	if *inFile == "" ||  *outFile == "" {
		fmt.Println("usage : -in inpufile.json -out outputfile.json")
		os.Exit(1)
	}
	fmt.Printf("Input file = %s\n", *inFile)
	fmt.Printf("Output file = %s\n", *outFile)

	
	f, err := os.Open(*inFile)
	if err != nil {
	   // handle error, later, or get me the job and ill handle it ;-)
	}
	dec := json.NewDecoder(f)

	for {
		var r Row
		if err := dec.Decode(&r); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("%v\n", r)

		if r.is_relevant() {
			r.processRow() 
		}

	}

	
	closeUpReporting()
	fmt.Printf("\n\nFinal vessel count = %d, %d stopped position found.\n", vesselCount, stopCount)
	
}