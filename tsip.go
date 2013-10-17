package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const MSG_MAX_LEN = 256

const (
	EXPECT_START = iota
	EXPECT_DONE
	IN_MESSAGE
)

type handleable interface {
	Handle()
}

type Action struct {
	MatchSequence []byte
	decodeInto    handleable
}

type PrimaryTimingPacket struct {
	Subcode    uint8
	TimeOfWeek uint32
	WeekNumber uint16
	UTCOffset  int16
	TimingFlag byte
	Seconds    uint8
	Minutes    uint8
	Hours      uint8
	DayOfMonth uint8
	Month      uint8
	Year       uint16
}

type SecondaryTimingPacket struct {
	Subcode              uint8
	ReceiverMode         uint8
	DiscipliningMode     uint8
	SelfSurveyProgress   uint8
	HoldoverDuration     uint32
	CriticalAlarms       uint16
	MinorAlarms          uint16
	GPSDecodeStatus      uint8
	DiscipliningActivity uint8
	SpareStatus1         uint8
	SpareStatus2         uint8
	PPSOffset            float32
	TenMhzOffset         float32
	DACValue             uint32
	DACVoltage           float32
	Temperature          float32
	Latitude             float64
	Longitude            float64
	Altitude             float64
	Spare                int64
}

func (p *SecondaryTimingPacket) Handle() {
	fmt.Printf("Secondary packet:  RCV %d, DIS %d, SUR %d PPS-OFFSET: %f CriticalAlarm: %x MinorAlarm: %x  Temp: %f\n", p.ReceiverMode, p.DiscipliningMode, p.SelfSurveyProgress, p.PPSOffset, p.CriticalAlarms, p.MinorAlarms, p.Temperature)
}

func (p *PrimaryTimingPacket) Handle() {
	fmt.Printf("Primary Timing Packet:  %d/%d/%d %d:%d:%d  (GPS offset %d)\n", p.Year, p.Month, p.DayOfMonth, p.Hours, p.Minutes, p.Seconds, p.UTCOffset)

}

var actions []Action

func init() {
	// HUMAN:  The parser requires that you append these in descending
	// order of MatchSequence length.
	actions = append(actions, Action{[]byte{0x8f, 0xab}, &PrimaryTimingPacket{}})
	actions = append(actions, Action{[]byte{0x8f, 0xac}, &SecondaryTimingPacket{}})
}

func handleMsg(msg []byte) {
	var p handleable

	for _, a := range actions {
		alen := len(a.MatchSequence)
		if bytes.Equal(msg[0:alen], a.MatchSequence) {
			p = a.decodeInto
			break
		}
	}

	r := bytes.NewReader(msg[1:])
	binary.Read(r, binary.BigEndian, p)
	p.Handle()
}

func main() {
	fmt.Println("connecting to serial server")
	conn, err := net.Dial("tcp", "192.168.2.111:6001")
	if err != nil {
		fmt.Println("could not connect:", err)
		return
	}
	br := bufio.NewReader(conn)
	// Find a start of message
	for {
		c, _ := br.ReadByte()
		if c == 0x10 {
			c, _ := br.ReadByte()
			if c == 0x3 {
				break
			}
		}
	}
	state := 0
	var msg [MSG_MAX_LEN]byte
	msgptr := 0
	for {
		c, _ := br.ReadByte()
		if c == 0x10 {
			// Attempt to de-stuff DLEs if they're in message data
			nextbytes, _ := br.Peek(1)
			if nextbytes[0] == 0x10 {
				c, _ = br.ReadByte()
			} else {
				if state == EXPECT_START {
					msgptr = 0
					state = IN_MESSAGE
					continue
				} else {
					state = EXPECT_DONE
					continue
				}
			}
		}

		if state == EXPECT_DONE {
			if c != 0x03 {
				fmt.Println("Error:  Expected to be done, got", c)
				return
			} else {
				handleMsg(msg[0:msgptr])
				state = EXPECT_START
			}
		}
		if msgptr < MSG_MAX_LEN {
			msg[msgptr] = c
			msgptr++
		} // Else silently discard the rest of the message.  *shrug*
		// This should not happen.
	}
}
