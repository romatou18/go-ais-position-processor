
# Go json AIS zero speed position parser
Takes an AIS vessel position feed, and tries detecting suspicious or illegal fishing or other unauthorised activities.

## Input output
This golang executable takes an AIS line delimited JSON file, containing AIS data messages. Some of them message types are positional messages containing a GPS position for a given vessel.
The message types are  :1, 2, 3, 18, 19, 27.

The program takes this AIS message list as an input, extracts the GPS position, calculates displacement using the Haversine formula, then computes the vessel speed.
For each vessel keeps a state object up to date with the [i-1]-th and [i]-th position with i in [1-N] and the position set from [0-N].
It processes each message iteratively and updates the vessel state on every new position found.
When calculating speed and time and meeting the following conditions: **speed < 1 knot for 60 mins or more** then a zero speed stopped geojson feature is created as an output.
Whenever the vessel moves again the speed status is reset in the relevant vessel state object, until finding a new stopped postion with the above mentioned conditions, or until running out of input data, reaching the end of the data input file.

## AIS JSON input example:
`
{"Message":{"MessageID":3,"RepeatIndicator":0,"UserID":12345,"Valid":true,"NavigationalStatus":3,"RateOfTurn":0,"Sog":1.7,"PositionAccuracy":false,"Longitude":166.90754833333332,"Latitude":-21.46249,"Cog":290.3,"TrueHeading":34,"Timestamp":57,"SpecialManoeuvreIndicator":0,"Spare":0,"Raim":false,"CommunicationState":4336},"UTCTimeStamp":1588636800}
{"Message":{"MessageID":5,"RepeatIndicator":0,"UserID":56789,"Valid":true,"AisVersion":0,"ImoNumber":9344382,"CallSign":"V2CZ5  ","Name":"GRETA               ","Type":79,"Dimension":{"A":86,"B":15,"C":11,"D":4},"FixType":1,"Eta":{"Month":4,"Day":18,"Hour":14,"Minute":0},"MaximumStaticDraught":3.9,"Destination":"CAPE TOWN           ","Dte":false,"Spare":false},"UTCTimeStamp":1588636800}
{"Message":{"MessageID":18,"RepeatIndicator":0,"UserID":89087,"Valid":true,"Spare1":0,"Sog":0,"PositionAccuracy":true,"Longitude":174.76563,"Latitude":-36.821688333333334,"Cog":360,"TrueHeading":511,"Timestamp":1,"Spare2":0,"ClassBUnit":true,"ClassBDisplay":false,"ClassBDsc":true,"ClassBBand":true,"ClassBMsg22":true,"AssignedMode":false,"Raim":true,"CommunicationStateIsItdma":true,"CommunicationState":393222},"UTCTimeStamp":1588636801}
{"Message":{"MessageID":19,"RepeatIndicator":0,"UserID":654321,"Valid":true,"Spare1":0,"Sog":0,"PositionAccuracy":false,"Longitude":170.18316666666666,"Latitude":-15.192758333333334,"Cog":0,"TrueHeading":511,"Timestamp":60,"Spare2":0,"Name":"AABB            11V3","Type":0,"Dimension":{"A":3,"B":5,"C":1,"D":2},"FixType":1,"Raim":false,"Dte":true,"AssignedMode":false,"Spare3":0},"UTCTimeStamp":1588636801}
{"Message":{"MessageID":27,"RepeatIndicator":3,"UserID":12345,"Valid":true,"PositionAccuracy":false,"Raim":false,"NavigationalStatus":0,"Longitude":10.520000000000001,"Latitude":-16.695,"Sog":9,"Cog":168,"PositionLatency":false,"Spare":false},"UTCTimeStamp":1588636801}
`

## GeoJson ouput file:

`{"type":"FeatureCollection","features":[<array of features...>]}`

A feature is of the following form:
{"type":"Feature","geometry":{"type":"Point","coordinates":[<longitude_degrees>, <latitude_degrees>]},"properties":{"name":<VesselID>,"duration_sec":<seconds>,"date_UTC":"<UTC time>"}}

For instance:
{"type":"Feature","geometry":{"type":"Point","coordinates":[173.54578, -41.67789]},"properties":{"name":12345,"duration_sec":3602,"date_UTC":"2021-06-05T01:02:21Z"}}

## Approach and performance considerations
Fast iterative approach using the Golang Reader/ Json decoder interface + state objects that lazy load the required data. This allows fast processing and low RAM needs.

Steps:.
- Reading json input line by line using the golang Json decoder buffered interface.
- Creating a state object per found AIS vessel, updating the state at each relevant record found.
- Updating an ouput map of selected position for the final position report.

On a modest 1500 NZD 2020 laptop i5-10210U with 16GB of RAM, the program processed a **3GB** input file counting **8.955.836** AIS messages (json records) in **1min20secs** on my laptop and used max 700MB of RAM.

## Sample input and output
File sample available in the **sample_data** folder.

## Pre-compiled executable
A linux x86_64 executable is provided **govessel_x86_64_linux**

## Build and run
### Pre requisite
- install golang version 1.15.8 or superior as per the go.mod file

### Build and Run the program
- go build .

`./govessel -in sample_data/part0000.json -out geojson_output_file.json`


## Mapping and representation on the world map
The generated geojson report can be uploaded on http://geojson.io in order to be visualised on the world map.
