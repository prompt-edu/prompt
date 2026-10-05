// Package standaloneModule describes services that are not a course phase type but still keep
// data per course phase or per person. Core asks them in its phase deletion and privacy fan-outs
// exactly like a phase module, through the endpoints the prompt-sdk registrars serve.
package standaloneModule

type Module struct {
	// Name labels the module in the privacy records and in error messages.
	Name string
	// BaseURL is the root the SDK endpoints are served under, such as https://host/ai/api.
	BaseURL string
}
