/*
Package extnet provides a portable interface for network I/O, including SCTP.
*/
package extnet

const (
	// RxBufferSize is network recieve queue size
	RxBufferSize = 10240
	// BacklogSize is accept queue size
	BacklogSize = 128

	// CloseNotifyPpid is used close listener
	CloseNotifyPpid uint32 = 4294967295
	// CloseNotifyCotext is used close listener
	CloseNotifyCotext uint32 = 4294967295
)

// Notificator is called when error or trace event are occured
var Notificator func(e error)

type sndrcvInfo struct {
	stream     uint16
	ssn        uint16
	flags      uint16
	ppid       uint32
	context    uint32
	timetolive uint32
	tsn        uint32
	cumtsn     uint32
	assocID    assocT
}

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

var ppidStr = map[int]string{
	0:  "Reserved by SCTP",
	1:  "IUA",
	2:  "M2UA",
	3:  "M3UA",
	4:  "SUA",
	5:  "M2PA",
	6:  "V5UA",
	7:  "H.248",
	8:  "BICC/Q.2150.3",
	9:  "TALI",
	10: "DUA",
	11: "ASAP",
	12: "ENRP",
	13: "H.323",
	14: "Q.IPC/Q.2150.3",
	15: "SIMCO",
	16: "DDP Segment Chunk",
	17: "DDP Stream Session Control",
	18: "S1AP",
	19: "RUA",
	20: "HNBAP",
	21: "ForCES-HP",
	22: "ForCES-MP",
	23: "ForCES-LP",
	24: "SBc-AP",
	25: "NBAP",
	27: "X2AP",
	28: "IRCP",
	29: "LCS-AP",
	30: "MPICH2",
	31: "SABP",
	32: "FGP",
	33: "PPP",
	34: "CALCAPP",
	35: "SSP",
	36: "NPMP-CONTROL",
	37: "NPMP-DATA",
	38: "ECHO",
	39: "DISCARD",
	40: "DAYTIME",
	41: "CHARGEN",
	42: "3GPP RNA",
	43: "3GPP M2AP",
	44: "3GPP M3AP",
	45: "SSH over SCTP",
	46: "Diameter in a SCTP DATA chunk",
	47: "Diameter in a DTLS/SCTP DATA chunk",
	48: "R14P. BER Encoded ASN.1 over SCTP",
	50: "WebRTC DCEP",
	51: "WebRTC String",
	52: "WebRTC Binary Partial (deprecated)",
	53: "WebRTC Binary",
	54: "WebRTC String Partial (deprecated)",
	55: "3GPP PUA",
	56: "WebRTC String Empty",
	57: "WebRTC Binary Empty",
	58: "3GPP XwAP",
	59: "3GPP Xw-Control Plane"}
