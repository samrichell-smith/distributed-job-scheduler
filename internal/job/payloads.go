package job

type AddNumbersPayload struct {
	X int
	Y int
}

type ReverseStringPayload struct {
	Text string
}

type ResizeImagePayload struct {
	URL    string
	Width  int
	Height int
}
