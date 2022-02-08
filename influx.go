package main

import (
	"fmt"
	"sort"
	"strings"
)

type InfluxData struct {
	Timestamp int64
	Name      string
	Tags      map[string]string
	Fields    map[string]string
}

// Create an InFluxData struct
func NewInfluxData() (m *InfluxData) {
	m = &InfluxData{}
	m.Tags = make(map[string]string)
	m.Fields = make(map[string]string)

	return
}

// Marshal InfluxData into into Influx wire protocol
func (m *InfluxData) Marshal() (line string) {

	tags := make([]string, 0, len(m.Tags))
	for tag := range m.Tags {
		tags = append(tags, tag+"="+m.Tags[tag])
	}
	sort.Strings(tags)

	fields := make([]string, 0, len(m.Fields))
	for field := range m.Fields {
		fields = append(fields, field+"="+m.Fields[field])
	}
	sort.Strings(fields)

	line = fmt.Sprintf("%s,%s %s %v\n",
		m.Name,
		strings.Join(tags, ","),
		strings.Join(fields, ","),
		m.Timestamp)

	return
}
