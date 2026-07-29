package constants

import "os"

type Landscape string

const (
	LandscapeLocal Landscape = "local"
	LandscapeTest  Landscape = "test"
	LandscapeDev   Landscape = "dev"
	LandscapeProd  Landscape = "prod"
)

func (l Landscape) IsValid() bool {
	switch l {
	case LandscapeLocal, LandscapeTest, LandscapeDev, LandscapeProd:
		return true
	default:
		return false
	}
}

func GetLandscapeFromEnv() Landscape {
	if _, ok := os.LookupEnv("LANDSCAPE"); !ok {
		return LandscapeLocal
	}
	l := Landscape(os.Getenv("LANDSCAPE"))
	if !l.IsValid() {
		panic("invalid Landscape environment variable")
	}
	return l
}
