// Standard motion topics: semantic parsing of NotificationMessage.

package events

import (
	"strconv"
	"strings"

	"github.com/mickeyzzc/onvif-go/v2/types"
)

// MotionAlarm is the motion state carried by a standard motion notification.
type MotionAlarm struct {
	Active bool
	Score  float64 // 0–100; 0 when the device does not provide one
	Source string  // Source SimpleItem value, e.g. "CSI" or a video source token
}

// motionTopics maps each recognized motion topic, normalized by
// normalizeTopic, to the Data item that carries its boolean state.
var motionTopics = map[string]string{
	"videosource/motionalarm":              "State",
	"ruleengine/cellmotiondetector/motion": "IsMotion",
}

// normalizeTopic lower-cases a topic path and drops the namespace prefix of
// every segment, so tns1:, tnsalias: and vendor prefixes all compare equal.
func normalizeTopic(topic string) string {
	segments := strings.Split(strings.TrimSpace(topic), "/")
	for i, seg := range segments {
		if j := strings.LastIndexByte(seg, ':'); j >= 0 {
			seg = seg[j+1:]
		}

		segments[i] = strings.ToLower(seg)
	}

	return strings.Join(segments, "/")
}

// IsMotionTopic reports whether topic is a motion topic ParseMotionAlarm
// recognizes. Matching is case-insensitive and ignores namespace prefixes.
func IsMotionTopic(topic string) bool {
	_, ok := motionTopics[normalizeTopic(topic)]

	return ok
}

// ParseMotionAlarm extracts the motion state from a notification. Two
// standard topics are recognized:
//
//   - tns1:VideoSource/MotionAlarm → Data "State"
//   - tns1:RuleEngine/CellMotionDetector/Motion → Data "IsMotion"
//
// Many cameras (all TP-Link models tested) emit only the second one; some
// (Hikvision) emit both. The state is read from the Data slot only. Score is
// an optional vendor extension in Data; a missing or unparseable Score is 0.
// Source is the Source item named "Source", else "VideoSourceConfigurationToken",
// else the first Source item.
//
// Any other topic, or a missing or unparseable state item, returns
// (MotionAlarm{}, false).
func ParseMotionAlarm(msg NotificationMessage) (MotionAlarm, bool) {
	key, ok := motionTopics[normalizeTopic(msg.Topic)]
	if !ok {
		return MotionAlarm{}, false
	}

	raw, ok := simpleItemValue(msg.Message.Data, key)
	if !ok {
		return MotionAlarm{}, false
	}

	active, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return MotionAlarm{}, false
	}

	alarm := MotionAlarm{Active: active, Source: motionSource(msg.Message)}

	if s, ok := simpleItemValue(msg.Message.Data, "Score"); ok {
		if score, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
			alarm.Score = score
		}
	}

	return alarm, true
}

func motionSource(m EventMessage) string {
	for _, name := range []string{"Source", "VideoSourceConfigurationToken"} {
		if v, ok := simpleItemValue(m.Source, name); ok {
			return v
		}
	}

	if len(m.Source) > 0 {
		return m.Source[0].Value
	}

	return ""
}

func simpleItemValue(items []types.SimpleItem, name string) (string, bool) {
	for _, it := range items {
		if it.Name == name {
			return it.Value, true
		}
	}

	return "", false
}
