package parser

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test the parsing of a simple ADEXP message
func TestParse(t *testing.T) {
	b, err := ioutil.ReadFile("../resources/tests/adexp.txt")
	if err != nil {
		t.Fatal(err)
	}

	lexemes := lex(b)
	m, err := parse(lexemes)
	if err != nil {
		t.Fatal(err)
	}

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

	assert.Equal(t, "(ACH-BEL20B-LIML1050-EBBR-DOF/150521-14/HOC/1120F320 -18/PBN/B1 DOF/150521 REG/OODWK RVR/150 OPR/BEL ORGN/LSAZZQZG SRC/AFP RMK/AGCS EQUIPPED)", m.MessageText)
	assert.Equal(t, "???FPD.F15: N0410F300 ARLES UL153 PUNSA/N0410F300 UL153 VADEM/N0400F320 UN853 PENDU/N0400F330 UN853 IXILU/N0400F340 UN853 DIK/N0400F320 UY37 BATTY", m.Comment)

}

func BenchmarkParser(b *testing.B) {
	raw, err := ioutil.ReadFile("../resources/tests/adexp.txt")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexemes := lex(raw)
		_, err = parse(lexemes)
		if err != nil {
			b.Fatal(err)
		}
	}
}
