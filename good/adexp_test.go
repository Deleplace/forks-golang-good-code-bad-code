package good

import (
	"fmt"
	"io/ioutil"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Test the parsing of a simple ADEXP message
func TestParseAdexpMessage(t *testing.T) {
	bytes, _ := ioutil.ReadFile("../resources/tests/adexp.txt")

	m, err := ParseAdexpMessage(bytes)
	assert.Nil(t, err)

	// Test upper level
	assert.Equal(t, true, m.IsUpperLevel())

	// Simple
	assert.Equal(t, "IFPL", m.Title)
	assert.Equal(t, "CYYZ", m.Adep)
	assert.Equal(t, "EASTERN :CREEK'()+,./", m.Alternate)
	assert.Equal(t, "AFIL", m.Ades)
	assert.Equal(t, "ACA878", m.Arcid)
	assert.Equal(t, "A333", m.ArcType)
	assert.Equal(t, "SDE3FGHIJ3J5LM1ORVWXY", m.Ceqpt)

	// Repeating
	assert.Equal(t, 13, len(m.Eetfir))
	assert.Equal(t, 2, len(m.Speed))

	// Complex
	assert.Equal(t, 2, len(m.Estdata))
	assert.Equal(t, 3, len(m.Geo))
	assert.Equal(t, 5, len(m.RoutePoints))

	// Route points
	assert.Equal(t, "CYYZ", m.RoutePoints[0].Ptid)
	assert.Equal(t, 0, m.RoutePoints[0].FlightLevel)
	assert.Equal(t, "170301220429", m.RoutePoints[0].Eto)
	assert.Equal(t, "JOOPY", m.RoutePoints[1].Ptid)
	assert.Equal(t, 390, m.RoutePoints[1].FlightLevel)
	assert.Equal(t, "170302002327", m.RoutePoints[1].Eto)
	assert.Equal(t, "GEO01", m.RoutePoints[2].Ptid)
	assert.Equal(t, 390, m.RoutePoints[2].FlightLevel)
	assert.Equal(t, "170302003347", m.RoutePoints[2].Eto)
	assert.Equal(t, "BLM", m.RoutePoints[3].Ptid)
	assert.Equal(t, 171, m.RoutePoints[3].FlightLevel)
	assert.Equal(t, "170302051642", m.RoutePoints[3].Eto)
	assert.Equal(t, "LSZH", m.RoutePoints[4].Ptid)
	assert.Equal(t, 14, m.RoutePoints[4].FlightLevel)
	assert.Equal(t, "170302052710", m.RoutePoints[4].Eto)
}

func TestFindSubfields(t *testing.T) {
	line := []byte("-ESTDATA -PTID XETBO -ETO 170302032300 -FL F390")
	subfields := findSubfields(line)
	assert.Equal(t, len(subfields), 4)
	assert.Equal(t, string(subfields[0]), "ESTDATA ")
	assert.Equal(t, string(subfields[1]), "PTID XETBO ")
	assert.Equal(t, string(subfields[2]), "ETO 170302032300 ")
	assert.Equal(t, string(subfields[3]), "FL F390")
}

func TestParseLine(t *testing.T) {
	line := []byte("-ESTDATA -PTID XETBO -ETO 170302032300 -FL F390")
	subfields := findSubfields(line)

	h, l := parseLine(subfields[0])
	assert.Equal(t, string(h), "ESTDATA")
	assert.Equal(t, string(l), "")

	h, l = parseLine(subfields[1])
	assert.Equal(t, string(h), "PTID")
	assert.Equal(t, string(l), "XETBO ")

	h, l = parseLine(subfields[2])
	assert.Equal(t, string(h), "ETO")
	assert.Equal(t, string(l), "170302032300 ")

	h, l = parseLine(subfields[3])
	assert.Equal(t, string(h), "FL")
	assert.Equal(t, string(l), "F390")
}

// Test ...
func TestParseSimpleToken(t *testing.T) {
	// Do something
}

// Test ...
func TestParseComplexToken(t *testing.T) {
	// Do something
}

// Performance test of an ADEXP message parsing
func BenchmarkParseAdexpMessage(t *testing.B) {
	log.SetLevel(log.FatalLevel)
	bytes, _ := ioutil.ReadFile("../resources/tests/adexp.txt")

	for i := 0; i < t.N; i++ {
		ParseAdexpMessage(bytes)
	}
}

// Performance test of an ADEXP message parsing
func BenchmarkParseBatch(b *testing.B) {
	log.SetLevel(log.FatalLevel)
	bytes, _ := ioutil.ReadFile("../resources/tests/adexp.txt")

	const C = 5000
	inputs := make([][]byte, C)
	for i := range inputs {
		inputs[i] = make([]byte, len(bytes))
		copy(inputs[i], bytes)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseMessages(inputs)
	}
}

func parseMessages(inputs [][]byte) []Message {
	C := len(inputs)
	messages := make([]Message, C)

	// Sequential
	//
	// for j := range inputs {
	// 	messages[j], _ = ParseAdexpMessage(inputs[j])
	// }

	// Concurrent
	//
	// var wg sync.WaitGroup
	// wg.Add(C)
	// for j := range inputs {
	// 	j := j
	// 	input := inputs[j]
	// 	go func() {
	// 		messages[j], _ = ParseAdexpMessage(input)
	// 		wg.Done()
	// 	}()
	// }
	// wg.Wait()

	// Concurrent batches: W workers of M messages
	const W = 100
	if C%W != 0 {
		panic(fmt.Sprintf("Can't uniformely dispatch %d to %d workers", C, W))
	}
	M := C / W
	var wg sync.WaitGroup
	wg.Add(W)
	for w := 0; w < W; w++ {
		hi, lo := w*M, (w+1)*M
		subinputs := inputs[hi:lo]
		submsg := messages[hi:lo]
		go func() {
			for j, input := range subinputs {
				submsg[j], _ = ParseAdexpMessage(input)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return messages
}
