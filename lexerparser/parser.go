package parser

import (
	"bytes"
)

func parse(words []lexeme) (*Message, error) {
	m := &Message{}

	nWords := len(words)
	var prevCommand lexeme
	for i := 0; i < nWords; {
		if !words[i].isCommand() {
			//return nil, fmt.Errorf("Expected command, got %q", string(words[i]))
		}
		command := words[i]
		switch {

		case bytes.Equal(command, cmdTitle):
			i += 1 + consumeTitleValue(words[i+1:], m)
		case bytes.Equal(command, cmdAdep):
			i += 1 + consumeAdepValue(words[i+1:], m)
		case bytes.Equal(command, cmdAltnz):
			i += 1 + consumeAlternateValue(words[i+1:], m)
		case bytes.Equal(command, cmdAdes):
			i += 1 + consumeAdesValue(words[i+1:], m)
		case bytes.Equal(command, cmdArcid):
			i += 1 + consumeArcidValue(words[i+1:], m)
		case bytes.Equal(command, cmdArctyp):
			i += 1 + consumeArcTypeValue(words[i+1:], m)
		case bytes.Equal(command, cmdCeqpt):
			i += 1 + consumeCeqptValue(words[i+1:], m)
		case bytes.Equal(command, cmdMsgtxt):
			i += 1 + consumeMessageTextValue(words[i+1:], m)
		case bytes.Equal(command, cmdComment):
			i += 1 + consumeCommentValue(words[i+1:], m)

		case bytes.Equal(command, cmdEetfir):
			i += 1 + consumeEetfirValue(words[i+1:], m)
		case bytes.Equal(command, cmdSpeed):
			i += 1 + consumeSpeedValue(words[i+1:], m)

		case bytes.Equal(command, cmdEstdata):
			i += 1 + consumeEstdataValue(words[i+1:], m)
		case bytes.Equal(command, cmdGeo):
			i += 1 + consumeGeoValue(words[i+1:], m)
		case bytes.Equal(command, cmdRtepts):
			i += 1 + consumeRteptsValue(words[i+1:], m)

		case bytes.Equal(command, wordBegin):
			// Assume always "-BEGIN RTEPTS"
			if !bytes.Equal(words[i+1], []byte("RTEPTS")) {
				panic(words[i+1])
			}
			i += 3 // Skip "RTEPTS", skip newline

			for {
				k := consumeRteptsValue(words[i:], m)
				i += k
				if k == 3 {
					// Read "-END", "RTEPTS", newline
					break
				}
			}

		case bytes.Equal(prevCommand, cmdComment):
			i += consumeCommentValue(words[i:], m)
			continue

		default:
			// ??
			panic(string(command))
		}
		prevCommand = command
	}

	return m, nil
}

func consumeLine(words []lexeme) (value string, consumed int) {
	b := make([]byte, 0, 100)
	for i, word := range words {
		if word.isNewline() {
			return string(b), 1 + i
		}
		if i > 0 {
			b = append(b, ' ')
		}
		b = append(b, word...)
	}
	return string(b), len(words)
}

func consumeTitleValue(words []lexeme, m *Message) (consumed int) {
	m.Title, consumed = consumeLine(words)
	return consumed
}

func consumeAdepValue(words []lexeme, m *Message) (consumed int) {
	m.Adep, consumed = consumeLine(words)
	return consumed
}

func consumeAlternateValue(words []lexeme, m *Message) (consumed int) {
	m.Alternate, consumed = consumeLine(words)
	return consumed
}

func consumeAdesValue(words []lexeme, m *Message) (consumed int) {
	m.Ades, consumed = consumeLine(words)
	return consumed
}

func consumeArcidValue(words []lexeme, m *Message) (consumed int) {
	m.Arcid, consumed = consumeLine(words)
	return consumed
}

func consumeArcTypeValue(words []lexeme, m *Message) (consumed int) {
	m.ArcType, consumed = consumeLine(words)
	return consumed
}

func consumeCeqptValue(words []lexeme, m *Message) (consumed int) {
	m.Ceqpt, consumed = consumeLine(words)
	return consumed
}

func consumeMessageTextValue(words []lexeme, m *Message) (consumed int) {
	m.MessageText, consumed = consumeLine(words)
	return consumed
}

func consumeCommentValue(words []lexeme, m *Message) (consumed int) {
	var value string
	value, consumed = consumeLine(words)
	if m.Comment == "" {
		m.Comment = value
	} else {
		m.Comment += " " + value
	}
	return consumed
}

func consumeEetfirValue(words []lexeme, m *Message) (consumed int) {
	var value string
	value, consumed = consumeLine(words)
	m.Eetfir = append(m.Eetfir, value)
	return consumed
}

func consumeSpeedValue(words []lexeme, m *Message) (consumed int) {
	var value string
	value, consumed = consumeLine(words)
	m.Speed = append(m.Speed, value)
	return consumed
}

func consumeEstdataValue(words []lexeme, m *Message) (consumed int) {
	// Expects exactly 6 words + 1 newline (or eof)
	var est estdata
	var err error
	for i := 0; i < 6; i += 2 {
		subcmd := words[i]
		value := string(words[i+1])
		switch {
		case bytes.Equal(subcmd, subCmdPtid):
			est.Ptid = value
		case bytes.Equal(subcmd, subCmdEto):
			est.Eto = value
		case bytes.Equal(subcmd, subCmdFl):
			est.FlightLevel, err = extractFlightLevel(value)
			checkerr(err)
		}
	}
	m.Estdata = append(m.Estdata, est)
	if len(words) == 6 {
		return 6
	}
	if !words[6].isNewline() {
		panic(string(words[6]))
	}
	return 7
}

func consumeGeoValue(words []lexeme, m *Message) (consumed int) {
	// Expects exactly 6 words + 1 newline
	var geo geo
	for i := 0; i < 6; i += 2 {
		subcmd := words[i]
		value := string(words[i+1])
		switch {
		case bytes.Equal(subcmd, subCmdGeoid):
			geo.Geoid = value
		case bytes.Equal(subcmd, subCmdLattd):
			geo.Latitude = value
		case bytes.Equal(subcmd, subCmdLongtd):
			geo.Longitude = value
		}
	}
	m.Geo = append(m.Geo, geo)
	if len(words) == 6 {
		return 6
	}
	if !words[6].isNewline() {
		panic(string(words[6]))
	}
	return 7
}

func consumeRteptsValue(words []lexeme, m *Message) (consumed int) {
	if len(words) >= 3 && bytes.Equal(words[0], wordEnd) {
		return 3
	}

	// Expects exactly 7 words + 1 newline
	var r rtepts
	var err error
	// Ignore "-PT"
	for i := 1; i < 7; i += 2 {
		subcmd := words[i]
		value := string(words[i+1])
		switch {
		case bytes.Equal(subcmd, subCmdPtid):
			r.Ptid = value
		case bytes.Equal(subcmd, subCmdEto):
			r.Eto = value
		case bytes.Equal(subcmd, subCmdFl):
			r.FlightLevel, err = extractFlightLevel(value)
			checkerr(err)
		}
	}
	m.RoutePoints = append(m.RoutePoints, r)
	if len(words) == 7 {
		return 7
	}
	if !words[7].isNewline() {
		panic(string(words[6]))
	}
	return 8
}

const (
	noCommand = iota
	simpleCommand
	complexCommand
)

var (
	// List of ADEXP tokens
	cmdTitle   = []byte("-TITLE")
	cmdAdep    = []byte("-ADEP")
	cmdAltnz   = []byte("-ALTNZ")
	cmdAdes    = []byte("-ADES")
	cmdArcid   = []byte("-ARCID")
	cmdArctyp  = []byte("-ARCTYP")
	cmdCeqpt   = []byte("-CEQPT")
	cmdMsgtxt  = []byte("-MSGTXT")
	cmdComment = []byte("-COMMENT")
	cmdEetfir  = []byte("-EETFIR")
	cmdSpeed   = []byte("-SPEED")
	cmdEstdata = []byte("-ESTDATA")
	cmdGeo     = []byte("-GEO")
	cmdRtepts  = []byte("-RTEPTS")

	// List of ADEXP subtokens
	subCmdPtid   = []byte("-PTID")
	subCmdEto    = []byte("-ETO")
	subCmdFl     = []byte("-FL")
	subCmdGeoid  = []byte("-GEOID")
	subCmdLattd  = []byte("-LATTD")
	subCmdLongtd = []byte("-LONGTD")

	// Control
	wordSpace = []byte(" ")
	wordBegin = []byte("-BEGIN")
	wordEnd   = []byte("-END")
)

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}
