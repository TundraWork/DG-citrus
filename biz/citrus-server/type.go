package citrus_server

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/json"
)

type RawEvent struct {
	Type     EventType `json:"type"`
	ClientId string    `json:"clientId"`
	TargetId string    `json:"targetId"`
	Message  string    `json:"message"`
}

func (e *RawEvent) FromByteArray(data []byte) error {
	err := json.Unmarshal(data, e)
	if err != nil {
		return err
	}
	return nil
}

func (e *RawEvent) ToByteArray() ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (e *RawEvent) ToEvent() (Event, error) {
	var event Event
	switch e.Type {
	case EventTypeHeartbeat:
		event = &EventHeartbeat{}
	case EventTypeBind:
		if e.Message == "targetId" {
			event = &EventBindToServer{}
		} else if e.Message == "DGLAB" {
			event = &EventBindAppToThirdParty{}
		} else {
			return nil, fmt.Errorf("unknown bind message format with message = %s", e.Message)
		}
	case EventTypeBreak:
		event = &EventBreak{}
	case EventTypeError:
		event = &EventError{}
	case EventTypeMsg:
		if strings.HasPrefix(e.Message, "strength-") {
			if len(strings.Split(e.Message, "+")) == 3 {
				event = &EventAdjustStrength{}
			} else if len(strings.Split(e.Message, "+")) == 4 {
				event = &EventReportStrength{}
			} else {
				return nil, fmt.Errorf("unknown message type: strength - unexpected number of values")
			}
		} else if strings.HasPrefix(e.Message, "pulse-") {
			event = &EventExecutePulse{}
		} else if strings.HasPrefix(e.Message, "clear-") {
			event = &EventStopPulse{}
		} else if strings.HasPrefix(e.Message, "feedback-") {
			event = &EventReportFeedback{}
		} else {
			return nil, fmt.Errorf("unknown message type: %s", e.Message)
		}
	default:
		return nil, fmt.Errorf("unknown event type: %s", e.Type)
	}
	err := event.FromRawEvent(e)
	if err != nil {
		return nil, err
	}
	return event, nil
}

type Event interface {
	FromRawEvent(e *RawEvent) error
	ToRawEvent() (*RawEvent, error)
	Process() error
}

type EventType string

const (
	EventTypeHeartbeat EventType = "heartbeat"
	EventTypeBind      EventType = "bind"
	EventTypeMsg       EventType = "msg"
	EventTypeBreak     EventType = "break"
	EventTypeError     EventType = "error"
)

type Channel int

const (
	ChannelUnknown Channel = iota
	ChannelA
	ChannelB
)

type EventHeartbeat struct {
	ClientId ClientSecureId `json:"clientId"`
	TargetId ClientSecureId `json:"targetId"`
}

func (e *EventHeartbeat) FromRawEvent(rawEvent *RawEvent) error {
	e.ClientId = ClientSecureId(rawEvent.ClientId)
	e.TargetId = ClientSecureId(rawEvent.TargetId)
	return nil
}

func (e *EventHeartbeat) ToRawEvent() (*RawEvent, error) {
	return &RawEvent{
		Type:     EventTypeHeartbeat,
		ClientId: string(e.ClientId),
		TargetId: string(e.TargetId),
		Message:  "",
	}, nil
}

type EventBindToServer struct {
	ClientId ClientSecureId `json:"clientId"`
}

func (e *EventBindToServer) FromRawEvent(_ *RawEvent) error {
	return fmt.Errorf("FromRawEvent should never be called for this event type")
}

func (e *EventBindToServer) ToRawEvent() (*RawEvent, error) {
	return &RawEvent{
		Type:     EventTypeBind,
		ClientId: string(e.ClientId),
		Message:  "targetId",
	}, nil
}

type EventBindAppToThirdParty struct {
	ClientId ClientSecureId `json:"clientId"`
	TargetId ClientSecureId `json:"targetId"`
}

func (e *EventBindAppToThirdParty) FromRawEvent(rawEvent *RawEvent) error {
	if rawEvent.Message != "DGLAB" {
		return fmt.Errorf("invalid message payload for bind event")
	}
	e.ClientId = ClientSecureId(rawEvent.ClientId)
	e.TargetId = ClientSecureId(rawEvent.TargetId)
	return nil
}

func (e *EventBindAppToThirdParty) ToRawEvent() (*RawEvent, error) {
	return nil, fmt.Errorf("ToRawEvent should never be called for this event type")
}

type EventBindResult struct {
	ClientId ClientSecureId `json:"clientId"`
	TargetId ClientSecureId `json:"targetId"`
	Code     int            `json:"code"`
}

func (e *EventBindResult) FromRawEvent(_ *RawEvent) error {
	return fmt.Errorf("FromRawEvent should never be called for this event type")
}

func (e *EventBindResult) ToRawEvent() (*RawEvent, error) {
	return &RawEvent{
		Type:     EventTypeBind,
		ClientId: string(e.ClientId),
		TargetId: string(e.TargetId),
		Message:  strconv.Itoa(e.Code),
	}, nil
}

type EventBreak struct {
	ClientId ClientSecureId `json:"clientId"`
	TargetId ClientSecureId `json:"targetId"`
}

func (e *EventBreak) FromRawEvent(rawEvent *RawEvent) error {
	e.ClientId = ClientSecureId(rawEvent.ClientId)
	e.TargetId = ClientSecureId(rawEvent.TargetId)
	return nil
}

func (e *EventBreak) ToRawEvent() (*RawEvent, error) {
	return &RawEvent{
		Type:     EventTypeBreak,
		ClientId: string(e.ClientId),
		TargetId: string(e.TargetId),
		Message:  "209",
	}, nil
}

type EventError struct {
	ClientId ClientSecureId `json:"clientId"`
	TargetId ClientSecureId `json:"targetId"`
	Message  string         `json:"message"`
}

func (e *EventError) FromRawEvent(rawEvent *RawEvent) error {
	e.ClientId = ClientSecureId(rawEvent.ClientId)
	e.TargetId = ClientSecureId(rawEvent.TargetId)
	e.Message = rawEvent.Message
	return nil
}

func (e *EventError) ToRawEvent() (*RawEvent, error) {
	return &RawEvent{
		Type:     EventTypeError,
		ClientId: string(e.ClientId),
		TargetId: string(e.TargetId),
		Message:  e.Message,
	}, nil
}

type EventReportStrength struct {
	ClientId ClientSecureId     `json:"clientId"`
	TargetId ClientSecureId     `json:"targetId"`
	Strength DataReportStrength `json:"strength"`
}
type DataReportStrength struct {
	ChannelAValue int `json:"channelAValue"`
	ChannelBValue int `json:"channelBValue"`
	ChannelALimit int `json:"channelALimit"`
	ChannelBLimit int `json:"channelBLimit"`
}

func (e *EventReportStrength) FromRawEvent(rawEvent *RawEvent) error {
	e.ClientId = ClientSecureId(rawEvent.ClientId)
	e.TargetId = ClientSecureId(rawEvent.TargetId)
	parts := strings.SplitN(rawEvent.Message, "-", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid report strength data format: missing delimiter")
	}
	values := strings.Split(parts[1], "+")
	var err error
	e.Strength.ChannelAValue, err = strconv.Atoi(values[0])
	if err != nil {
		return fmt.Errorf("error parsing data for report strength: %s", err)
	}
	e.Strength.ChannelBValue, err = strconv.Atoi(values[1])
	if err != nil {
		return fmt.Errorf("error parsing data for report strength: %s", err)
	}
	e.Strength.ChannelALimit, err = strconv.Atoi(values[2])
	if err != nil {
		return fmt.Errorf("error parsing data for report strength: %s", err)
	}
	e.Strength.ChannelBLimit, err = strconv.Atoi(values[3])
	if err != nil {
		return fmt.Errorf("error parsing data for report strength: %s", err)
	}
	return nil
}

func (e *EventReportStrength) ToRawEvent() (*RawEvent, error) {
	return &RawEvent{
		Type:     EventTypeMsg,
		ClientId: string(e.ClientId),
		TargetId: string(e.TargetId),
		Message:  fmt.Sprintf("strength-%d+%d+%d+%d", e.Strength.ChannelAValue, e.Strength.ChannelBValue, e.Strength.ChannelALimit, e.Strength.ChannelBLimit),
	}, nil
}

type EventAdjustStrength struct {
	ClientId ClientSecureId     `json:"clientId"`
	TargetId ClientSecureId     `json:"targetId"`
	Strength DataAdjustStrength `json:"strength"`
}
type DataAdjustStrength struct {
	Channel Channel            `json:"channel"`
	Type    AdjustStrengthType `json:"type"`
	Value   int                `json:"value"`
}
type AdjustStrengthType int

const (
	AdjustStrengthTypeDecrease = iota
	AdjustStrengthTypeIncrease
	AdjustStrengthTypeSet
)

func (e *EventAdjustStrength) FromRawEvent(rawEvent *RawEvent) error {
	e.ClientId = ClientSecureId(rawEvent.ClientId)
	e.TargetId = ClientSecureId(rawEvent.TargetId)
	parts := strings.SplitN(rawEvent.Message, "-", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid strength data format: missing delimiter")
	}
	values := strings.Split(parts[1], "+")
	var err error
	channel, err := strconv.Atoi(values[0])
	if err != nil {
		return fmt.Errorf("error parsing data for adjust strength: %s", err)
	}
	e.Strength.Channel = Channel(channel)
	mode, err := strconv.Atoi(values[1])
	if err != nil {
		return fmt.Errorf("error parsing data for adjust strength: %s", err)
	}
	e.Strength.Type = AdjustStrengthType(mode)
	e.Strength.Value, err = strconv.Atoi(values[2])
	if err != nil {
		return fmt.Errorf("error parsing data for adjust strength: %s", err)
	}
	return nil
}

func (e *EventAdjustStrength) ToRawEvent() (*RawEvent, error) {
	return &RawEvent{
		Type:     EventTypeMsg,
		ClientId: string(e.ClientId),
		TargetId: string(e.TargetId),
		Message:  fmt.Sprintf("strength-%d+%d+%d", e.Strength.Channel, e.Strength.Type, e.Strength.Value),
	}, nil
}

type EventExecutePulse struct {
	ClientId       ClientSecureId  `json:"clientId"`
	TargetId       ClientSecureId  `json:"targetId"`
	Channel        Channel         `json:"channel"`
	PulseSequences []PulseSequence `json:"pulseSequences"`
}
type WaveformFrequency int
type WaveformStrength int
type WaveformFrequencySequence [4]WaveformFrequency
type WaveformStrengthSequence [4]WaveformStrength
type PulseSequence struct {
	FrequencySequence WaveformFrequencySequence `json:"frequencySequence"`
	StrengthSequence  WaveformStrengthSequence  `json:"strengthSequence"`
}

func (e *EventExecutePulse) FromRawEvent(rawEvent *RawEvent) error {
	e.ClientId = ClientSecureId(rawEvent.ClientId)
	e.TargetId = ClientSecureId(rawEvent.TargetId)
	values := strings.Split(rawEvent.Message, ":")
	if len(values) != 2 {
		return fmt.Errorf("invalid pulse data format: missing pulse sequence")
	}
	channel, err := strconv.Atoi(values[0])
	if err != nil {
		return fmt.Errorf("invalid pulse data format: failed to parse channel")
	}
	e.Channel = Channel(channel)
	var pulseSequenceHexes []string
	if err := json.Unmarshal([]byte(values[1]), &pulseSequenceHexes); err != nil {
		return fmt.Errorf("invalid pulse data format: failed to parse pulse sequences as JSON")
	}
	for _, pulseSequenceHex := range pulseSequenceHexes {
		bytes, err := hex.DecodeString(pulseSequenceHex)
		if err != nil {
			return fmt.Errorf("invalid pulse data format: failed to decode pulse sequence hex")
		}
		if len(bytes) != 8 {
			return fmt.Errorf("invalid pulse data format: unexpected pulse sequence length")
		}
		var pulseSequence PulseSequence
		for i := 0; i < 4; i++ {
			pulseSequence.FrequencySequence[i] = WaveformFrequency(bytes[i])
			pulseSequence.StrengthSequence[i] = WaveformStrength(bytes[i+4])
		}
		e.PulseSequences = append(e.PulseSequences, pulseSequence)
	}
	return nil
}

func (e *EventExecutePulse) ToRawEvent() (*RawEvent, error) {
	pulseSequenceHexes := make([]string, 0, len(e.PulseSequences))
	for _, pulseSequence := range e.PulseSequences {
		var bytes [8]byte
		for i := 0; i < 4; i++ {
			bytes[i] = byte(pulseSequence.FrequencySequence[i])
			bytes[i+4] = byte(pulseSequence.StrengthSequence[i])
		}
		pulseSequenceHexes = append(pulseSequenceHexes, hex.EncodeToString(bytes[:]))
	}
	pulseSequencesJson, err := json.Marshal(pulseSequenceHexes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pulse sequences as JSON: %s", err)
	}
	return &RawEvent{
		Type:     EventTypeMsg,
		ClientId: string(e.ClientId),
		TargetId: string(e.TargetId),
		Message:  fmt.Sprintf("%d:%s", e.Channel, pulseSequencesJson),
	}, nil
}

type EventStopPulse struct {
	ClientId ClientSecureId `json:"clientId"`
	TargetId ClientSecureId `json:"targetId"`
	Channel  Channel        `json:"channel"`
}

func (e *EventStopPulse) FromRawEvent(rawEvent *RawEvent) error {
	e.ClientId = ClientSecureId(rawEvent.ClientId)
	e.TargetId = ClientSecureId(rawEvent.TargetId)
	channel, err := strconv.Atoi(rawEvent.Message)
	if err != nil {
		return err
	}
	e.Channel = Channel(channel)
	return nil
}

func (e *EventStopPulse) ToRawEvent() (*RawEvent, error) {
	return &RawEvent{
		Type:     EventTypeMsg,
		ClientId: string(e.ClientId),
		TargetId: string(e.TargetId),
		Message:  fmt.Sprintf("clear-%d", e.Channel),
	}, nil
}

type EventReportFeedback struct {
	ClientId ClientSecureId `json:"clientId"`
	TargetId ClientSecureId `json:"targetId"`
	Button   ButtonIndex    `json:"button"`
}
type ButtonIndex int

const (
	ButtonIndexChannelA1 = iota
	ButtonIndexChannelA2
	ButtonIndexChannelA3
	ButtonIndexChannelA4
	ButtonIndexChannelA5
	ButtonIndexChannelB1
	ButtonIndexChannelB2
	ButtonIndexChannelB3
	ButtonIndexChannelB4
	ButtonIndexChannelB5
)

func (e *EventReportFeedback) FromRawEvent(rawEvent *RawEvent) error {
	e.ClientId = ClientSecureId(rawEvent.ClientId)
	e.TargetId = ClientSecureId(rawEvent.TargetId)
	button, err := strconv.Atoi(strings.Split(rawEvent.Message, "-")[1])
	if err != nil {
		return err
	}
	e.Button = ButtonIndex(button)
	return nil
}

func (e *EventReportFeedback) ToRawEvent() (*RawEvent, error) {
	return &RawEvent{
		Type:     EventTypeMsg,
		ClientId: string(e.ClientId),
		TargetId: string(e.TargetId),
		Message:  fmt.Sprintf("feedback-%d", e.Button),
	}, nil
}

func _() {
	// ignore unused constants, should be removed after adding tests
	_, _, _, _, _ = EventTypeHeartbeat, EventTypeBind, EventTypeMsg, EventTypeBreak, EventTypeError
	_, _, _ = ChannelA, ChannelB, ChannelUnknown
	_, _, _ = AdjustStrengthTypeDecrease, AdjustStrengthTypeIncrease, AdjustStrengthTypeSet
	_, _, _, _, _ = ButtonIndexChannelA1, ButtonIndexChannelA2, ButtonIndexChannelA3, ButtonIndexChannelA4, ButtonIndexChannelA5
	_, _, _, _, _ = ButtonIndexChannelB1, ButtonIndexChannelB2, ButtonIndexChannelB3, ButtonIndexChannelB4, ButtonIndexChannelB5
}
