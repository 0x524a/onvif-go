package events

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/mickeyzzc/onvif-go/v2/types"
)

// Tolerant wsnt:Notify decoding.
//
// Field devices disagree on the wire shape of a Notify POST: prefix dialects
// vary per vendor, some wrap the whole response envelope inside another
// envelope, some POST the Notify element as the document root, some emit
// literal Name/Value child pairs instead of SimpleItem attributes, and some
// heartbeat empty Notifies. A path-anchored xml.Unmarshal cannot see through
// those variations, so the Notify decoder walks a local-name-keyed tree
// instead (the parser this package's Push-Notify support was originally
// field-hardened against — ported from the MiBeeNvr reference
// implementation, keep the dialect tolerance intact).

// xmlNode is a local-name-keyed XML tree node: prefix dialects differ between
// devices, so all matching in this file is by local name.
type xmlNode struct {
	name     string
	attrs    []xml.Attr
	textBuf  strings.Builder
	children []*xmlNode
	next     *xmlNode // next sibling (Name/Value pair parsing)
	parent   *xmlNode
}

func localXMLName(full string) string {
	if idx := strings.Index(full, ":"); idx >= 0 {
		return full[idx+1:]
	}
	return full
}

// parseXMLTree builds a synthetic root whose children are the document's
// top-level elements.
func parseXMLTree(data []byte) (*xmlNode, error) {
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	root := &xmlNode{name: "#document"}
	stack := []*xmlNode{root}
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		top := stack[len(stack)-1]
		switch t := tok.(type) {
		case xml.StartElement:
			node := &xmlNode{name: localXMLName(t.Name.Local), attrs: append([]xml.Attr(nil), t.Attr...), parent: top}
			if n := len(top.children); n > 0 {
				top.children[n-1].next = node
			}
			top.children = append(top.children, node)
			stack = append(stack, node)
		case xml.EndElement:
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			top.textBuf.Write(t)
		}
	}
	return root, nil
}

func (n *xmlNode) text() string { return n.textBuf.String() }

// attr returns an attribute value by local name (namespace ignored).
func (n *xmlNode) attr(local string) string {
	for _, a := range n.attrs {
		if localXMLName(a.Name.Local) == local {
			return a.Value
		}
	}
	return ""
}

// firstLocal returns the first direct child with the given local name.
func (n *xmlNode) firstLocal(name string) *xmlNode {
	for _, c := range n.children {
		if c.name == name {
			return c
		}
	}
	return nil
}

// collectLocal gathers every node with the given local name in the subtree.
func (n *xmlNode) collectLocal(name string, out *[]*xmlNode) {
	for _, c := range n.children {
		if c.name == name {
			*out = append(*out, c)
		}
		c.collectLocal(name, out)
	}
}

// findFault searches the tree for a SOAP Fault and maps it to an error.
func findNotifyFault(root *xmlNode) error {
	var fault *xmlNode
	var walk func(n *xmlNode)
	walk = func(n *xmlNode) {
		if fault != nil {
			return
		}
		if n.name == "Fault" {
			fault = n
			return
		}
		for _, c := range n.children {
			walk(c)
		}
	}
	walk(root)
	if fault == nil {
		return nil
	}
	text := "SOAP fault"
	if code := fault.firstLocal("Code"); code != nil {
		if v := code.firstLocal("Value"); v != nil {
			val := strings.TrimSpace(v.text())
			if val != "" {
				text += ": " + val
			}
		}
		if sub := code.firstLocal("Subcode"); sub != nil {
			if v := sub.firstLocal("Value"); v != nil {
				if subval := strings.TrimSpace(v.text()); subval != "" {
					text += ": " + subval
				}
			}
		}
	}
	if reason := fault.firstLocal("Reason"); reason != nil {
		if txt := reason.firstLocal("Text"); txt != nil {
			if r := strings.TrimSpace(txt.text()); r != "" {
				text += ": " + r
			}
		}
	}
	return errors.New(text)
}

// parseItems reads SimpleItem children (Name/Value attributes — the ONVIF
// event payload form) plus literal Name/Value child pairs some devices emit.
func parseItems(parent *xmlNode) []types.SimpleItem {
	var items []types.SimpleItem
	for _, child := range parent.children {
		switch child.name {
		case "SimpleItem", "ElementItem":
			name := child.attr("Name")
			if name != "" {
				items = append(items, types.SimpleItem{Name: name, Value: child.attr("Value")})
			}
		case "Name":
			// paired with a following Value sibling
			if child.next != nil && child.next.name == "Value" {
				items = append(items, types.SimpleItem{
					Name:  strings.TrimSpace(child.text()),
					Value: child.next.text(),
				})
			}
		}
	}
	return items
}

// notifyMessageFromNode maps one NotificationMessage tree node to the public
// shape. The canonical ONVIF payload is double-layer — the wsnt:Message
// wrapper around an inner payload element (tt:Message) that carries @UtcTime
// and the Source/Key/Data groups; devices that inline them on wsnt:Message
// directly are tolerated: each facet reads whichever layer has it.
func notifyMessageFromNode(msg *xmlNode) NotificationMessage {
	out := NotificationMessage{}
	if topic := msg.firstLocal("Topic"); topic != nil {
		out.Topic = strings.TrimSpace(topic.text())
	}
	if pr := msg.firstLocal("ProducerReference"); pr != nil {
		if addr := pr.firstLocal("Address"); addr != nil {
			out.ProducerAddress = strings.TrimSpace(addr.text())
		}
	}
	if m := msg.firstLocal("Message"); m != nil {
		layer := m
		if inner := m.firstLocal("Message"); inner != nil {
			layer = inner
		}
		for _, cand := range []*xmlNode{layer, m} {
			if out.Message.UtcTime.IsZero() {
				if ts, ok := parseWSNTime(cand.attr("UtcTime")); ok {
					out.Message.UtcTime = ts
				}
			}
			if out.Message.PropertyOperation == "" {
				out.Message.PropertyOperation = cand.attr("PropertyOperation")
			}
		}
		for _, group := range []string{"Source", "Key", "Data"} {
			for _, cand := range []*xmlNode{layer, m} {
				if g := cand.firstLocal(group); g != nil {
					if got := parseItems(g); len(got) > 0 {
						switch group {
						case "Source":
							out.Message.Source = got
						case "Key":
							out.Message.Key = got
						default:
							out.Message.Data = got
						}
						break
					}
				}
			}
		}
	}
	return out
}

// parseNotifyTree decodes a wsnt:Notify SOAP body (what a device POSTs to the
// consumer endpoint) into messages. Fault bodies surface as errors; a
// well-formed Notify with zero NotificationMessage elements is valid (devices
// heartbeat empty Notifies) and yields an empty slice.
func parseNotifyTree(body []byte) ([]NotificationMessage, error) {
	root, err := parseXMLTree(body)
	if err != nil {
		return nil, fmt.Errorf("parse Notify: %w", err)
	}
	if len(root.children) == 0 {
		return nil, fmt.Errorf("parse Notify: body contains no XML elements")
	}
	if fault := findNotifyFault(root); fault != nil {
		return nil, fault
	}
	// There must be a Notify element somewhere — an envelope without one is
	// garbage, while a Notify with zero NotificationMessage elements is a
	// valid heartbeat delivery.
	var notifies []*xmlNode
	root.collectLocal("Notify", &notifies)
	if len(notifies) == 0 {
		return nil, fmt.Errorf("parse Notify: no Notify element")
	}
	var nodes []*xmlNode
	root.collectLocal("NotificationMessage", &nodes)
	msgs := make([]NotificationMessage, 0, len(nodes))
	for _, n := range nodes {
		msgs = append(msgs, notifyMessageFromNode(n))
	}
	return msgs, nil
}
