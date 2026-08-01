package constants

const (
	TestAgentName            = "Test Agent"
	TestAgentLocationAddress = "5345 Magnolia"
)

type TestAgent struct {
	AgentRef       string
	PhoneNumber    string
	PhoneNumberRef string
	EHRSubdomain   string
	EHRLocationRef string
}

var prodTestAgent = TestAgent{
	AgentRef:       "abc",
	PhoneNumber:    "abc",
	PhoneNumberRef: "xyz",
	EHRSubdomain:   "prod-dentist",
	EHRLocationRef: "1234",
}

var nonProdTestAgent = TestAgent{
	AgentRef:       "assistant-cf1ab171-db2b-48b8-a7b6-57930f547615",
	PhoneNumber:    "+13366541083",
	PhoneNumberRef: "3012822334093395781",
	EHRSubdomain:   "local-dentist",
	EHRLocationRef: "354656",
}

func TestAgentFor(landscape Landscape) TestAgent {
	if landscape == LandscapeProd {
		return prodTestAgent
	}
	return nonProdTestAgent
}
