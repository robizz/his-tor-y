package arghandler

// output - Custom type to hold value for the output style.
type Output int

// Declare related constants for each Output
const (
	Text Output = iota
	Json
)

// String - Creating common behavior - give the type a String function
func (o Output) String() string {
	return [...]string{"text", "json"}[o]
}
