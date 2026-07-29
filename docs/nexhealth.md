# Working with NexHealth

NexHealth is the PMS/EHR sync provider behind `constants.EHRNexHealth`. Their
Synchronizer product (https://synchronizer.io) polls each practice's EHR and
exposes it through NexHealth's REST API.

**Full API/docs index for agents:** https://docs.nexhealth.com/llms.txt — fetch
this first when you need anything not covered below (it links every guide and
API reference page on their site). Key pages it points to:

- API reference (all endpoints): https://docs.nexhealth.com/reference
- Onboardings endpoints: https://docs.nexhealth.com/reference/postonboardings,
  https://docs.nexhealth.com/reference/getonboardings,
  https://docs.nexhealth.com/reference/getonboardingshashid
- Supported health record systems: https://docs.nexhealth.com/docs/supported-health-record-systems
- Synchronizer install/troubleshooting guides: `docs/nexhealth-synchronizer-installation-guide-1`,
  `docs/troubleshooting-installs`

## EHR name constants

Supported EHR/PMS systems for onboarding live in `internal/constants/nexhealth.go`
as `constants.NexHealthEHR*`. Passing the matching value as `ehr_name` when
creating an onboarding lets the practice skip manually picking their system.

Only `NexHealthEHRDentrix` ("dentrix") is confirmed directly by NexHealth's own
example in the create-onboarding reference. The rest are slugified from display
names on the supported-systems page and have **not** been confirmed against a
live onboarding call — verify the exact `ehr_name` value against a real
request (or NexHealth support) before depending on one in production code.

## Existing integration points in this repo

- `internal/constants/ehr.go` — `EHR` enum (`EHRNexHealth`)
- `internal/biz/ehr.go` / `internal/contracts/ehr.go` / `internal/storage/ehr.go` —
  NexHealth subdomain + location ID plumbing for a practice's EHR config
- `internal/services/ehr/nexhealth/nexhealth.go` — NexHealth Synchronizer API client
- `config/local.json` / `config/dev.json` — `base_url: https://nexhealth.info`
