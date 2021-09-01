package consumption

import (
	"encoding/binary"
	"io"
	"strconv"
	"strings"

	"github.com/jacobsa/go-serial/serial"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/cfgwarn"
	"github.com/elastic/beats/v7/metricbeat/mb"
)

// init registers the MetricSet with the central registry as soon as the program
// starts. The New function will be called later to instantiate an instance of
// the MetricSet for each host defined in the module's configuration. After the
// MetricSet has been created then Fetch will begin to be called periodically.
func init() {
	mb.Registry.MustAddMetricSet("power", "consumption", New)
}

// MetricSet holds any configuration or state information. It must implement
// the mb.MetricSet interface. And this is best achieved by embedding
// mb.BaseMetricSet because it implements all of the required mb.MetricSet
// interface methods except for Fetch.
type MetricSet struct {
	mb.BaseMetricSet
	node   uint16
	power1 uint16
	power2 uint16
	power3 uint16
	power4 uint16
	vrms   float32
}

// New creates a new instance of the MetricSet. New is responsible for unpacking
// any MetricSet specific configuration options if there are any.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	cfgwarn.Beta("The power consumption metricset is beta.")

	config := struct{}{}
	if err := base.Module().UnpackConfig(&config); err != nil {
		return nil, err
	}

	return &MetricSet{
		BaseMetricSet: base,
		node:          0,
		power1:        0,
		power2:        0,
		power3:        0,
		power4:        0,
		vrms:          0,
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right
// format. It publishes the event which is then forwarded to the output. In case
// of an error set the Error field of mb.Event or simply call report.Error().
func (m *MetricSet) Fetch(report mb.ReporterV2) error {

	//x1 := binary.LittleEndian.Uint16(b[0:sudo ])

	// Set up options.
	options := serial.OpenOptions{
		PortName:        "/dev/ttyAMA0",
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		return err
	}

	// Make sure to close it later.
	defer port.Close()

	input := ""
	lastTwoChars := ""

	for lastTwoChars != string([]byte{13, 10}) {
		buf := make([]byte, 128)
		n, err := port.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
		} else {
			buf = buf[:n]
			lastTwoChars = string(buf[len(buf)-2 : len(buf)])
			input = input + string(buf)
		}
	}

	s := strings.Split(strings.TrimSpace(input), " ")
	vals := []byte{}

	for _, i := range s {
		j, err := strconv.Atoi(i)
		if err != nil {
			return err
		}
		vals = append(vals, byte(j))
	}

	if len(vals) < 11 {
		return nil
	}

	node := uint16(vals[0])

	if node == 10 {
		m.node = node
		m.power1 = binary.LittleEndian.Uint16(vals[1:3])
		m.power2 = binary.LittleEndian.Uint16(vals[3:5])
		m.power3 = binary.LittleEndian.Uint16(vals[5:7])
		m.power4 = binary.LittleEndian.Uint16(vals[7:9])
		m.vrms = float32(binary.LittleEndian.Uint16(vals[9:11])) / 100

		report.Event(mb.Event{
			MetricSetFields: common.MapStr{
				"node":   m.node,
				"power1": m.power1,
				"power2": m.power2,
				"power3": m.power3,
				"power4": m.power4,
				"vrms":   m.vrms,
			},
		})
	}

	return nil
}

//kWh for 30 seconds

//kWh = (1,500 × 0.0083333) ÷ 1,000
//kWh = 3,750 ÷ 1,000
//kWh = 3.75
