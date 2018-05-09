package good

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"regexp"
)

const (
	// Pattern to parse subfields in a complex token
	patternSubfield = "-.[^-]*"

	// List of ADEXP tokens
	tokenTitle   = "TITLE"
	tokenAdep    = "ADEP"
	tokenAltnz   = "ALTNZ"
	tokenAdes    = "ADES"
	tokenArcid   = "ARCID"
	tokenArctyp  = "ARCTYP"
	tokenCeqpt   = "CEQPT"
	tokenMsgtxt  = "MSGTXT"
	tokenComment = "COMMENT"
	tokenEetfir  = "EETFIR"
	tokenSpeed   = "SPEED"
	tokenEstdata = "ESTDATA"
	tokenGeo     = "GEO"
	tokenRtepts  = "RTEPTS"

	// List of ADEXP subtokens
	subtokenPtid   = "PTID"
	subtokenEto    = "ETO"
	subtokenFl     = "FL"
	subtokenGeoid  = "GEOID"
	subtokenLattd  = "LATTD"
	subtokenLongtd = "LONGTD"
)

var (
	bytesNewline     = []byte("\n")
	bytesNewlineDash = []byte("\n-")
	bytesBegin       = []byte("-BEGIN")
	bytesEnd         = []byte("-END")
	bytesEmpty       = []byte(" ")
	bytesComment     = []byte("//")
	bytesDash        = []byte("-")

	regexpSubfield = regexp.MustCompile(patternSubfield)

	// Map containing the mapping function given a specific token name
	factory = map[string]func(string, []byte) interface{}{
		tokenTitle:   parseSimpleToken,
		tokenAdep:    parseSimpleToken,
		tokenAltnz:   parseSimpleToken,
		tokenAdes:    parseSimpleToken,
		tokenArcid:   parseSimpleToken,
		tokenArctyp:  parseSimpleToken,
		tokenCeqpt:   parseSimpleToken,
		tokenMsgtxt:  parseSimpleToken,
		tokenComment: parseSimpleToken,

		tokenEetfir: parseSimpleToken,
		tokenSpeed:  parseSimpleToken,

		tokenEstdata: parseComplexToken,
		tokenGeo:     parseComplexToken,
		tokenRtepts:  parseComplexToken,
	}
)

// Parse an ADEXP message using a byte list as an input. This function returns a Message and an eventual error in case of a parsing error.
func ParseAdexpMessage(bytes []byte) (Message, error) {
	log.Debugf("parsing: %v", string(bytes))

	// Preprocessing
	preprocessed, err := preprocess(bytes)
	if err != nil {
		return Message{}, err
	}

	// Actual processing
	message, err := process(preprocessed)
	if err != nil {
		return Message{}, err
	}

	log.Debugf("returning message: %v", message)

	return message, nil
}

// Preprocessing of an ADEXP message (cleaning white spaces, rearranging multi-lined tokens etc.). Returns a byte list cleansed and an eventual error if the input is invalid.
func preprocess(in []byte) ([][]byte, error) {
	if len(in) == 0 {
		log.Errorf("input is empty")
		return nil, errors.New("input is empty")
	}

	lines := bytes.Split(in, bytesNewline)
	var result [][]byte
	var currentLine []byte

	for _, line := range lines {
		if bytes.HasPrefix(line, bytesEnd) {
			// Nothing
		} else if bytes.HasPrefix(line, bytesBegin) {
			result = append(result, currentLine)

			trimed := trim(line)
			currentLine = append(bytesDash, trimed[len(bytesBegin)+1:]...)
		} else if bytes.HasPrefix(line, bytesDash) {
			result = append(result, currentLine)

			currentLine = trim(line)
		} else if bytes.HasPrefix(line, bytesEmpty) {
			currentLine = append(append(currentLine, bytesEmpty...), trim(line)...)
		} else {
			currentLine = append(append(currentLine, bytesEmpty...), trim(line)...)
		}
	}

	if len(currentLine) > 0 {
		result = append(result, currentLine)
	}

	return result, nil
}

// Processing of an ADEXP message. Returns a Message structure and an eventual error in case of a processing error.
func process(in [][]byte) (Message, error) {
	msg := Message{}

	// Gather the goroutine results
	for _, line := range in {
		data := mapLine(line)

		// A mapper function can return a nil value (a line is potentially invalid, a comment etc.). In that case we simply discard the line.
		if data == nil {
			continue
		}

		// Enrich the message depending on the data type sent by the goroutines
		switch data.(type) {
		case simpleToken:
			simpleToken := data.(simpleToken)
			value := simpleToken.value
			switch simpleToken.token {
			case tokenTitle:
				msg.Title = value
			case tokenAdep:
				msg.Adep = value
			case tokenAltnz:
				msg.Alternate = value
			case tokenAdes:
				msg.Ades = value
			case tokenArcid:
				msg.Arcid = value
			case tokenArctyp:
				msg.ArcType = value
			case tokenCeqpt:
				msg.Ceqpt = value
			case tokenMsgtxt:
				msg.MessageText = value
			case tokenComment:
				msg.Comment = value
			case tokenEetfir:
				msg.Eetfir = append(msg.Eetfir, value)
			case tokenSpeed:
				msg.Speed = append(msg.Speed, value)
			default:
				log.Errorf("unexpected token type %v", simpleToken.token)
				return Message{}, fmt.Errorf("unexpected token type %v", simpleToken.token)
			}
		case complexToken:
			complexToken := data.(complexToken)
			value := complexToken.value

			switch complexToken.token {
			case tokenEstdata:
				for _, v := range value {
					fl, err := extractFlightLevel(v[subtokenFl])
					if err != nil {
						return Message{}, fmt.Errorf("flight level %v cannot be parsed", fl)
					}
					msg.Estdata = append(msg.Estdata, estdata{v[subtokenPtid], v[subtokenEto], fl})
				}
			case tokenGeo:
				for _, v := range value {
					msg.Geo = append(msg.Geo, geo{v[subtokenGeoid], v[subtokenLattd], v[subtokenLongtd]})
				}
			case tokenRtepts:
				for _, v := range value {
					fl, err := extractFlightLevel(v[subtokenFl])
					if err != nil {
						return Message{}, fmt.Errorf("flight level %v cannot be parsed", fl)
					}
					msg.RoutePoints = append(msg.RoutePoints, rtepts{v[subtokenPtid], fl, v[subtokenEto]})
				}
			default:
				log.Errorf("unexpected token type %v", complexToken.token)
				return Message{}, fmt.Errorf("unexpected token type %v", complexToken.token)
			}
		}
	}

	msg.Type = AdexpType

	return msg, nil
}

// Process a line and returns a token
func mapLine(in []byte) interface{} {
	// Filter empty lines and comment lines
	if len(in) == 0 || bytes.HasPrefix(in, bytesComment) {
		return nil
	}
	if in[0] != '-' {
		log.Warnf("Line doesn't start with a proper token: %q", string(in))
		return nil
	}

	token, value := parseLine(in[1:])
	if token == nil {
		log.Warnf("Token name is empty on line %v", string(in))
		return nil
	}

	sToken := string(token)

	// Checks in the factory map if the token has been configured
	if f, contains := factory[sToken]; contains {
		return f(sToken, value)
	}

	log.Warnf("Token %v is not managed by the parser", string(in))
	return nil
}

// Parse a simple token and returns a simpleToken structure
func parseSimpleToken(token string, value []byte) interface{} {
	return simpleToken{token, string(value)}
}

// Parse a complex token and returns a commplexToken structure
func parseComplexToken(token string, value []byte) interface{} {
	if len(value) == 0 {
		log.Warnf("Empty value")
		return complexToken{token, nil}
	}

	var v []map[string]string
	currentMap := make(map[string]string)

	// Find all subfields
	matches := findSubfields(value)

	// Iterate over each subfields to enrich the returned data
	for _, sub := range matches {
		h, l := parseLine(sub)

		if _, contains := currentMap[string(h)]; contains {
			v = append(v, currentMap)
			currentMap = make(map[string]string)
		}

		currentMap[string(h)] = string(trim(l))
	}

	// Append the latest map
	v = append(v, currentMap)

	return complexToken{token, v}
}

// Extract subfields from a line (with dash removed).
// E.g.
//       "-ESTDATA -PTID XETBO -ETO 170302032300 -FL F390"
//    -> ["ESTDATA ", "PTID XETBO ", "ETO 170302032300 ", "FL F390"]
// This is efficient because each element is a subslice of the original line.
func findSubfields(value []byte) (subfields [][]byte) {
	subfields = bytes.Split(value, bytesDash)
	if len(subfields) > 0 && len(trim(subfields[0])) == 0 {
		subfields = subfields[1:]
	}
	return subfields
}

// This custom loop is faster than the generic-purpose bytes.Trim .
// It expects 1 char == 1 byte (no multi-byte UTF-8 runes)
func trim(s []byte) []byte {
	const space = ' '
	n := len(s)
	low, high := 0, n
	for low < n && s[low] == space {
		low++
	}
	for high > low && s[high-1] == space {
		high--
	}
	return s[low:high]
}
