package constants

type NexHealthEHR string

const (
	NexHealthEHRAthena NexHealthEHR = "athena"
	NexHealthCurve2    NexHealthEHR = "curve2"
	NexHealthEHRCloud9 NexHealthEHR = "cloud9"

	NexHealthEHRDenticon      NexHealthEHR = "denticon"
	NexHealthEHRDentrix       NexHealthEHR = "dentrix"
	NexHealthEHRDentrixAscend NexHealthEHR = "dentrixascend"

	NexHealthDrChrono NexHealthEHR = "drchrono"

	NexHealthEHREaglesoft NexHealthEHR = "eaglesoft"
	NexHealthECW          NexHealthEHR = "ecw"

	NexHealthEHRModMed     NexHealthEHR = "modmed"
	NexHealthEHROpenDental NexHealthEHR = "opendental"

	NexHealthEHRNextGen NexHealthEHR = "nextgen"
)

var NexHealthEHRs = []NexHealthEHR{
	NexHealthEHRAthena,
	NexHealthEHRCloud9,
	NexHealthEHRDenticon,
	NexHealthEHRDentrix,
	NexHealthEHRDentrixAscend,
	NexHealthEHREaglesoft,
	NexHealthEHRModMed,
	NexHealthEHROpenDental,
	NexHealthCurve2,
	NexHealthDrChrono,
	NexHealthECW,
	NexHealthEHRNextGen,
}

func (e NexHealthEHR) IsValid() bool {
	switch e {
	case NexHealthEHRAthena, NexHealthEHRCloud9, NexHealthEHRDenticon,
		NexHealthEHRDentrix, NexHealthEHRDentrixAscend, NexHealthCurve2,
		NexHealthEHREaglesoft, NexHealthEHRModMed, NexHealthEHROpenDental,
		NexHealthDrChrono, NexHealthECW, NexHealthEHRNextGen:
		return true
	default:
		return false
	}
}

func (e NexHealthEHR) String() string {
	return string(e)
}
