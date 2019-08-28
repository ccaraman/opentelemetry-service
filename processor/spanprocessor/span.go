// Copyright 2019, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spanprocessor

import (
	"context"
	"strconv"
	"strings"

	tracepb "github.com/census-instrumentation/opencensus-proto/gen-go/trace/v1"

	"github.com/open-telemetry/opentelemetry-service/consumer"
	"github.com/open-telemetry/opentelemetry-service/consumer/consumerdata"
	"github.com/open-telemetry/opentelemetry-service/oterr"
	"github.com/open-telemetry/opentelemetry-service/processor"
)

type spanProcessor struct {
	nextConsumer consumer.TraceConsumer
	config       Config
	lenkeys      int
}

// NewTraceProcessor returns the span processor.
func NewTraceProcessor(nextConsumer consumer.TraceConsumer, config Config) (processor.TraceProcessor, error) {
	if nextConsumer == nil {
		return nil, oterr.ErrNilNextConsumer
	}

	sp := &spanProcessor{
		nextConsumer: nextConsumer,
		config:       config,
		lenkeys:      len(config.Rename.FromAttributes),
	}

	return sp, nil
}

func (sp *spanProcessor) ConsumeTraceData(ctx context.Context, td consumerdata.TraceData) error {
	for _, span := range td.Spans {
		if span == nil || span.Attributes == nil || len(span.Attributes.AttributeMap) == 0 {
			continue
		}
		// Name the span using attribute values.
		sp.NameSpan(span)
	}
	return sp.nextConsumer.ConsumeTraceData(ctx, td)
}

func (sp *spanProcessor) NameSpan(span *tracepb.Span) {
	// Currently, there is no preallocation for the length of the string builder
	// backing the new span name. If this does become a performance issue, the
	// methods Grow() and Len() and the length returned by WriteString() can be
	// used to improve allocation patterns.
	var sb strings.Builder
	for i, key := range sp.config.Rename.FromAttributes {
		attribute, found := span.Attributes.AttributeMap[key]

		// If one of the keys isn't found, the span name is not updated.
		if !found {
			return
		}

		// Note: WriteString() always return a nil error so there is no error checking
		// for this method call.
		// https://golang.org/src/strings/builder.go?s=3425:3477#L110

		// Include the separator before appending an attribute value if:
		// this isn't the first value(ie i == 0) loop through the FromAttributes
		// and
		// the separator isn't an empty string.
		if i > 0 && sp.config.Rename.Separator != "" {
			sb.WriteString(sp.config.Rename.Separator)
		}

		// Ideally with proto converting to the internal format for attributes
		// there shouldn't be any map entries with a nil value. However,
		// if there is a bad translation, this might be possible.
		if attribute == nil {
			sb.WriteString("<nil-attribute-value>")
			continue
		}

		switch value := attribute.Value.(type) {
		case *tracepb.AttributeValue_StringValue:
			sb.WriteString(value.StringValue.GetValue())
		case *tracepb.AttributeValue_BoolValue:
			sb.WriteString(strconv.FormatBool(value.BoolValue))
		case *tracepb.AttributeValue_DoubleValue:
			sb.WriteString(strconv.FormatFloat(value.DoubleValue, 'f', -1, 64))
		case *tracepb.AttributeValue_IntValue:
			sb.WriteString(strconv.FormatInt(value.IntValue, 10))
		default:
			sb.WriteString("<unknown-attribute-type>")
		}
	}
	span.Name = &tracepb.TruncatableString{Value: sb.String()}

}

func (sp *spanProcessor) AppendToSlice(span *tracepb.Span) {
	nameParts := make([]string, 0, sp.lenkeys)
	for _, key := range sp.config.Rename.FromAttributes {
		attrib, ok := span.Attributes.AttributeMap[key]
		if !ok {
			break
		}
		var namePart string
		if attrib == nil {
			namePart = "<nil-attribute-value>"
		} else {
			switch attribValue := attrib.Value.(type) {
			case *tracepb.AttributeValue_StringValue:
				namePart = attribValue.StringValue.GetValue()
			case *tracepb.AttributeValue_IntValue:
				namePart = strconv.FormatInt(attribValue.IntValue, 10)
			case *tracepb.AttributeValue_BoolValue:
				namePart = strconv.FormatBool(attribValue.BoolValue)
			case *tracepb.AttributeValue_DoubleValue:
				namePart = strconv.FormatFloat(attribValue.DoubleValue, 'f', -1, 64)
			default:
				namePart = "<unknown-attribute-type>"
			}
		}
		nameParts = append(nameParts, namePart)
	}
	if len(nameParts) == sp.lenkeys {
		if span.Name == nil {
			span.Name = &tracepb.TruncatableString{}
		}
		span.Name.Value = strings.Join(nameParts, sp.config.Rename.Separator)
	}
}

func (sp *spanProcessor) IndexInsert(span *tracepb.Span) {
	nameParts := make([]string, sp.lenkeys)
	for i, key := range sp.config.Rename.FromAttributes {
		attrib, ok := span.Attributes.AttributeMap[key]
		if !ok {
			break
		}
		var namePart string
		if attrib == nil {
			namePart = "<nil-attribute-value>"
		} else {
			switch attribValue := attrib.Value.(type) {
			case *tracepb.AttributeValue_StringValue:
				namePart = attribValue.StringValue.GetValue()
			case *tracepb.AttributeValue_IntValue:
				namePart = strconv.FormatInt(attribValue.IntValue, 10)
			case *tracepb.AttributeValue_BoolValue:
				namePart = strconv.FormatBool(attribValue.BoolValue)
			case *tracepb.AttributeValue_DoubleValue:
				namePart = strconv.FormatFloat(attribValue.DoubleValue, 'f', -1, 64)
			default:
				namePart = "<unknown-attribute-type>"
			}
		}
		nameParts[i] = namePart
	}
	if len(nameParts) == sp.lenkeys {
		if span.Name == nil {
			span.Name = &tracepb.TruncatableString{}
		}
		span.Name.Value = strings.Join(nameParts, sp.config.Rename.Separator)
	}
}
