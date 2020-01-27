package mystruct

import "errors"

//Howout enum holds proper value for how batsman got out
type Howout string

//Bowled Catch
const (
	Bowled    Howout = "b"
	Catch            = "c"
	Runout           = "ro"
	Notout           = "no"
	Didnotbat        = "dnb"
	Stumped          = "st"
)

//Valid function returns nil if the passed var is valid
func (h Howout) Valid() error {
	switch h {
	case Bowled, Catch, Runout, Notout, Didnotbat, Stumped:
		return nil
	}

	return errors.New("Invalid howout, it has to be one of [b,c,ro,no,dnb,st] ")
}
