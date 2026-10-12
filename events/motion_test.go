package events

import (
	"context"
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/types"
)

func motionMsg(topic string, source, data []types.SimpleItem) NotificationMessage {
	return NotificationMessage{Topic: topic, Message: EventMessage{Source: source, Data: data}}
}

func item(name, value string) types.SimpleItem { return types.SimpleItem{Name: name, Value: value} }

func TestParseMotionAlarm(t *testing.T) {
	tests := []struct {
		name   string
		msg    NotificationMessage
		want   MotionAlarm
		wantOK bool
	}{
		{
			name: "MotionAlarm State true with Score",
			msg: motionMsg("tns1:VideoSource/MotionAlarm",
				[]types.SimpleItem{item("Source", "CSI")},
				[]types.SimpleItem{item("State", "true"), item("Score", "87")}),
			want:   MotionAlarm{Active: true, Score: 87, Source: "CSI"},
			wantOK: true,
		},
		{
			name: "MotionAlarm garbage Score tolerated as 0",
			msg: motionMsg("tns1:VideoSource/MotionAlarm",
				[]types.SimpleItem{item("Source", "VideoSource_1")},
				[]types.SimpleItem{item("State", "true"), item("Score", "n/a")}),
			want:   MotionAlarm{Active: true, Source: "VideoSource_1"},
			wantOK: true,
		},
		{
			name: "CellMotionDetector IsMotion true",
			msg: motionMsg("tns1:RuleEngine/CellMotionDetector/Motion",
				[]types.SimpleItem{
					item("VideoSourceConfigurationToken", "VideoSourceToken"),
					item("VideoAnalyticsConfigurationToken", "VideoAnalyticsToken"),
					item("Rule", "MyMotionDetectorRule"),
				},
				[]types.SimpleItem{item("IsMotion", "true")}),
			want:   MotionAlarm{Active: true, Source: "VideoSourceToken"},
			wantOK: true,
		},
		{
			name: "CellMotionDetector IsMotion false",
			msg: motionMsg("tns1:RuleEngine/CellMotionDetector/Motion", nil,
				[]types.SimpleItem{item("IsMotion", "false")}),
			want:   MotionAlarm{},
			wantOK: true,
		},
		{
			name: "CellMotionDetector without IsMotion rejected",
			msg: motionMsg("tns1:RuleEngine/CellMotionDetector/Motion", nil,
				[]types.SimpleItem{item("State", "true")}),
		},
		{
			name: "MotionAlarm without State rejected",
			msg: motionMsg("tns1:VideoSource/MotionAlarm", nil,
				[]types.SimpleItem{item("IsMotion", "true")}),
		},
		{
			name: "state key in the Source slot is not read",
			msg: motionMsg("tns1:RuleEngine/CellMotionDetector/Motion",
				[]types.SimpleItem{item("IsMotion", "true")}, nil),
		},
		{
			name: "unparseable state rejected",
			msg: motionMsg("tns1:VideoSource/MotionAlarm", nil,
				[]types.SimpleItem{item("State", "maybe")}),
		},
		{
			name: "upper-case topic accepted",
			msg: motionMsg("TNS1:RULEENGINE/CELLMOTIONDETECTOR/MOTION", nil,
				[]types.SimpleItem{item("IsMotion", "true")}),
			want:   MotionAlarm{Active: true},
			wantOK: true,
		},
		{
			name: "other rule-engine topic rejected",
			msg: motionMsg("tns1:RuleEngine/TamperDetector/Tamper", nil,
				[]types.SimpleItem{item("IsTamper", "true")}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseMotionAlarm(tt.msg)
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("ParseMotionAlarm = (%+v, %v), want (%+v, %v)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestIsMotionTopic(t *testing.T) {
	tests := []struct {
		topic string
		want  bool
	}{
		{"tns1:VideoSource/MotionAlarm", true},
		{"tns1:RuleEngine/CellMotionDetector/Motion", true},
		{"TNS1:RULEENGINE/CELLMOTIONDETECTOR/MOTION", true},
		{"tnsalias:VideoSource/MotionAlarm", true},
		{"tns1:RuleEngine/tns1:CellMotionDetector/tns1:Motion", true},
		{" tns1:VideoSource/MotionAlarm ", true},
		{"tns1:RuleEngine/TamperDetector/Tamper", false},
		{"tns1:Device/Humidity", false},
		{"tns1:VideoSource/MotionAlarmExtra", false},
		{"tns1:RuleEngine/CellMotionDetector", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := IsMotionTopic(tt.topic); got != tt.want {
			t.Errorf("IsMotionTopic(%q) = %v, want %v", tt.topic, got, tt.want)
		}
	}
}

// pullMessagesHikvisionDualMotion is a field capture from a Hikvision
// DS-IPC-T12HV3-IA (V5.7.12), which emits both standard motion topics.
const pullMessagesHikvisionDualMotion = `<PullMessagesResponse xmlns="http://www.onvif.org/ver10/events/wsdl" xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2" xmlns:tt="http://www.onvif.org/ver10/schema">
<CurrentTime>2026-10-12T05:07:04Z</CurrentTime>
<TerminationTime>2026-10-12T05:08:04Z</TerminationTime>
<wsnt:NotificationMessage><wsnt:Topic Dialect="http://www.onvif.org/ver10/tev/topicExpression/ConcreteSet">tns1:VideoSource/MotionAlarm</wsnt:Topic>
<wsnt:Message><tt:Message UtcTime="2026-10-12T05:07:04Z" PropertyOperation="Initialized"><tt:Source><tt:SimpleItem Name="Source" Value="VideoSource_1"/>
</tt:Source>
<tt:Data><tt:SimpleItem Name="State" Value="false"/>
</tt:Data>
</tt:Message>
</wsnt:Message>
</wsnt:NotificationMessage>
<wsnt:NotificationMessage><wsnt:Topic Dialect="http://www.onvif.org/ver10/tev/topicExpression/ConcreteSet">tns1:RuleEngine/CellMotionDetector/Motion</wsnt:Topic>
<wsnt:Message><tt:Message UtcTime="2026-10-12T05:07:04Z" PropertyOperation="Initialized"><tt:Source><tt:SimpleItem Name="VideoSourceConfigurationToken" Value="VideoSourceToken"/>
<tt:SimpleItem Name="VideoAnalyticsConfigurationToken" Value="VideoAnalyticsToken"/>
<tt:SimpleItem Name="Rule" Value="MyMotionDetectorRule"/>
</tt:Source>
<tt:Data><tt:SimpleItem Name="IsMotion" Value="false"/>
</tt:Data>
</tt:Message>
</wsnt:Message>
</wsnt:NotificationMessage>
</PullMessagesResponse>`

// TestParseMotionAlarmFromPullMessages runs field captures through the real
// PullMessages decode path, so the helper is tested on what the library
// actually delivers.
func TestParseMotionAlarmFromPullMessages(t *testing.T) {
	caller := newFakeCaller(func(string, string) (string, error) {
		return pullMessagesHikvisionDualMotion, nil
	})

	messages, err := New(caller).PullMessages(context.Background(), "http://fake/sub/1", 10*time.Second, 5)
	if err != nil {
		t.Fatalf("PullMessages: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("messages = %d, want 2", len(messages))
	}

	want := []MotionAlarm{
		{Active: false, Source: "VideoSource_1"},
		{Active: false, Source: "VideoSourceToken"},
	}
	for i, msg := range messages {
		got, ok := ParseMotionAlarm(msg)
		if !ok || got != want[i] {
			t.Errorf("message %d (%s): ParseMotionAlarm = (%+v, %v), want (%+v, true)", i, msg.Topic, got, ok, want[i])
		}
	}
}
