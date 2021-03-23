package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dave/lakeland_pilgrimage/gpx"
)

type AllData struct {
	Legs      []LegData
	Waypoints []WaypointData
	Mapping   []MappingData
}

type MappingData struct {
	Gpx, Data string
}

type WaypointData struct {
	Waypoint, Description string
	Terrain, Summary      string
}

type LegData struct {
	Leg                     int
	From, To                string
	Highlights, Description string
}

func main() {

	out := &gpx.Root{}

	dataBytes, err := os.ReadFile("./data.json")
	if err != nil {
		panic(err)
	}
	var data AllData
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		panic(err)
	}

	mappings := map[string]string{}
	gpxNulls := map[string]bool{}
	dataNulls := map[string]bool{}
	for _, m := range data.Mapping {
		if m.Data == "" {
			gpxNulls[m.Gpx] = true
		} else if m.Gpx == "" {
			dataNulls[m.Data] = true
		} else {
			mappings[m.Gpx] = m.Data
		}
	}

	waypointsData := map[string]WaypointData{}
	for _, waypoint := range data.Waypoints {
		waypointsData[waypoint.Waypoint] = waypoint
	}

	legsData := map[int]LegData{}
	for _, leg := range data.Legs {
		legsData[leg.Leg] = leg
	}

	tracksGpx, err := gpx.Load("./routes.gpx")
	if err != nil {
		panic(err)
	}
	for _, track := range tracksGpx.Tracks {
		leg := legsData[track.Leg]
		route := gpx.Route{
			Name:   fmt.Sprintf("Leg %d / 12: %s to %s", leg.Leg, leg.From, leg.To),
			Desc:   fmt.Sprintf("%s\n\n%s", strings.Replace(leg.Highlights, "•", "★", -1), leg.Description),
			Points: gpx.LinePoints(track.Segments[0].Line()),
		}
		out.Routes = append(out.Routes, route)
	}

	waypointsGpx, err := gpx.Load("./waypoints.gpx")
	if err != nil {
		panic(err)
	}
	done := map[string]bool{}
	for _, waypoint := range waypointsGpx.Waypoints {
		name := waypoint.Name
		if mappings[name] != "" {
			name = mappings[name]
		}
		w := waypointsData[name]
		if w == (WaypointData{}) {
			if !gpxNulls[name] {
				panic(name)
			}
		} else {
			done[name] = true
		}
		waypoint.Name = name
		if w.Summary != "" {
			waypoint.Name += " (" + w.Summary + ")"
		}
		if w != (WaypointData{}) {
			waypoint.Desc = "★ " + w.Description
			if w.Terrain != "" {
				waypoint.Desc += "\n\nTerrain: " + w.Terrain
			}
		}
		out.Waypoints = append(out.Waypoints, waypoint)
	}
	for _, w := range waypointsData {
		if !done[w.Waypoint] {
			if !dataNulls[w.Waypoint] {
				panic(w.Waypoint)
			}
		}
	}

	err = out.Save("./lakeland-pilgrimage-12-legs.gpx")
	if err != nil {
		panic(err)
	}
}
