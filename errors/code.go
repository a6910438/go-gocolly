package errors

type code uint

const (
	errUnknown        code = 0
	errRecordNotFound code = 10000
)

func (c code) String() string {
	switch c {
	case errUnknown:
		return "unknown"
	case errRecordNotFound:
		return "record not found"
	}
	return ""
}
