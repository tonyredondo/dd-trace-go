package tracer

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *ciTestCyclePayload) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "version":
			z.Version, err = dc.ReadInt32()
			if err != nil {
				err = msgp.WrapError(err, "Version")
				return
			}
		case "metadata":
			var zb0002 uint32
			zb0002, err = dc.ReadMapHeader()
			if err != nil {
				err = msgp.WrapError(err, "Metadata")
				return
			}
			if z.Metadata == nil {
				z.Metadata = make(map[string]map[string]string, zb0002)
			} else if len(z.Metadata) > 0 {
				for key := range z.Metadata {
					delete(z.Metadata, key)
				}
			}
			for zb0002 > 0 {
				zb0002--
				var za0001 string
				var za0002 map[string]string
				za0001, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Metadata")
					return
				}
				var zb0003 uint32
				zb0003, err = dc.ReadMapHeader()
				if err != nil {
					err = msgp.WrapError(err, "Metadata", za0001)
					return
				}
				if za0002 == nil {
					za0002 = make(map[string]string, zb0003)
				} else if len(za0002) > 0 {
					for key := range za0002 {
						delete(za0002, key)
					}
				}
				for zb0003 > 0 {
					zb0003--
					var za0003 string
					var za0004 string
					za0003, err = dc.ReadString()
					if err != nil {
						err = msgp.WrapError(err, "Metadata", za0001)
						return
					}
					za0004, err = dc.ReadString()
					if err != nil {
						err = msgp.WrapError(err, "Metadata", za0001, za0003)
						return
					}
					za0002[za0003] = za0004
				}
				z.Metadata[za0001] = za0002
			}
		case "events":
			err = z.Events.DecodeMsg(dc)
			if err != nil {
				err = msgp.WrapError(err, "Events")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ciTestCyclePayload) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "version"
	err = en.Append(0x83, 0xa7, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteInt32(z.Version)
	if err != nil {
		err = msgp.WrapError(err, "Version")
		return
	}
	// write "metadata"
	err = en.Append(0xa8, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61)
	if err != nil {
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Metadata)))
	if err != nil {
		err = msgp.WrapError(err, "Metadata")
		return
	}
	for za0001, za0002 := range z.Metadata {
		err = en.WriteString(za0001)
		if err != nil {
			err = msgp.WrapError(err, "Metadata")
			return
		}
		err = en.WriteMapHeader(uint32(len(za0002)))
		if err != nil {
			err = msgp.WrapError(err, "Metadata", za0001)
			return
		}
		for za0003, za0004 := range za0002 {
			err = en.WriteString(za0003)
			if err != nil {
				err = msgp.WrapError(err, "Metadata", za0001)
				return
			}
			err = en.WriteString(za0004)
			if err != nil {
				err = msgp.WrapError(err, "Metadata", za0001, za0003)
				return
			}
		}
	}
	// write "events"
	err = en.Append(0xa6, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73)
	if err != nil {
		return
	}
	err = z.Events.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "Events")
		return
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ciTestCyclePayload) Msgsize() (s int) {
	s = 1 + 8 + msgp.Int32Size + 9 + msgp.MapHeaderSize
	if z.Metadata != nil {
		for za0001, za0002 := range z.Metadata {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + msgp.MapHeaderSize
			if za0002 != nil {
				for za0003, za0004 := range za0002 {
					_ = za0004
					s += msgp.StringPrefixSize + len(za0003) + msgp.StringPrefixSize + len(za0004)
				}
			}
		}
	}
	s += 7 + z.Events.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ciTestCyclePayloadList) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0002 uint32
	zb0002, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(ciTestCyclePayloadList, zb0002)
	}
	for zb0001 := range *z {
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				err = msgp.WrapError(err, zb0001)
				return
			}
			(*z)[zb0001] = nil
		} else {
			if (*z)[zb0001] == nil {
				(*z)[zb0001] = new(ciTestCyclePayload)
			}
			err = (*z)[zb0001].DecodeMsg(dc)
			if err != nil {
				err = msgp.WrapError(err, zb0001)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ciTestCyclePayloadList) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0003 := range z {
		if z[zb0003] == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			err = z[zb0003].EncodeMsg(en)
			if err != nil {
				err = msgp.WrapError(err, zb0003)
				return
			}
		}
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z ciTestCyclePayloadList) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zb0003 := range z {
		if z[zb0003] == nil {
			s += msgp.NilSize
		} else {
			s += z[zb0003].Msgsize()
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ciVisibilityEvent) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "type":
			z.Type, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Type")
				return
			}
		case "version":
			z.Version, err = dc.ReadInt32()
			if err != nil {
				err = msgp.WrapError(err, "Version")
				return
			}
		case "content":
			err = z.Content.DecodeMsg(dc)
			if err != nil {
				err = msgp.WrapError(err, "Content")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *ciVisibilityEvent) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "type"
	err = en.Append(0x83, 0xa4, 0x74, 0x79, 0x70, 0x65)
	if err != nil {
		return
	}
	err = en.WriteString(z.Type)
	if err != nil {
		err = msgp.WrapError(err, "Type")
		return
	}
	// write "version"
	err = en.Append(0xa7, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteInt32(z.Version)
	if err != nil {
		err = msgp.WrapError(err, "Version")
		return
	}
	// write "content"
	err = en.Append(0xa7, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74)
	if err != nil {
		return
	}
	err = z.Content.EncodeMsg(en)
	if err != nil {
		err = msgp.WrapError(err, "Content")
		return
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ciVisibilityEvent) Msgsize() (s int) {
	s = 1 + 5 + msgp.StringPrefixSize + len(z.Type) + 8 + msgp.Int32Size + 8 + z.Content.Msgsize()
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ciVisibilityEvents) DecodeMsg(dc *msgp.Reader) (err error) {
	var zb0002 uint32
	zb0002, err = dc.ReadArrayHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if cap((*z)) >= int(zb0002) {
		(*z) = (*z)[:zb0002]
	} else {
		(*z) = make(ciVisibilityEvents, zb0002)
	}
	for zb0001 := range *z {
		if dc.IsNil() {
			err = dc.ReadNil()
			if err != nil {
				err = msgp.WrapError(err, zb0001)
				return
			}
			(*z)[zb0001] = nil
		} else {
			if (*z)[zb0001] == nil {
				(*z)[zb0001] = new(ciVisibilityEvent)
			}
			var field []byte
			_ = field
			var zb0003 uint32
			zb0003, err = dc.ReadMapHeader()
			if err != nil {
				err = msgp.WrapError(err, zb0001)
				return
			}
			for zb0003 > 0 {
				zb0003--
				field, err = dc.ReadMapKeyPtr()
				if err != nil {
					err = msgp.WrapError(err, zb0001)
					return
				}
				switch msgp.UnsafeString(field) {
				case "type":
					(*z)[zb0001].Type, err = dc.ReadString()
					if err != nil {
						err = msgp.WrapError(err, zb0001, "Type")
						return
					}
				case "version":
					(*z)[zb0001].Version, err = dc.ReadInt32()
					if err != nil {
						err = msgp.WrapError(err, zb0001, "Version")
						return
					}
				case "content":
					err = (*z)[zb0001].Content.DecodeMsg(dc)
					if err != nil {
						err = msgp.WrapError(err, zb0001, "Content")
						return
					}
				default:
					err = dc.Skip()
					if err != nil {
						err = msgp.WrapError(err, zb0001)
						return
					}
				}
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ciVisibilityEvents) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteArrayHeader(uint32(len(z)))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0004 := range z {
		if z[zb0004] == nil {
			err = en.WriteNil()
			if err != nil {
				return
			}
		} else {
			// map header, size 3
			// write "type"
			err = en.Append(0x83, 0xa4, 0x74, 0x79, 0x70, 0x65)
			if err != nil {
				return
			}
			err = en.WriteString(z[zb0004].Type)
			if err != nil {
				err = msgp.WrapError(err, zb0004, "Type")
				return
			}
			// write "version"
			err = en.Append(0xa7, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
			if err != nil {
				return
			}
			err = en.WriteInt32(z[zb0004].Version)
			if err != nil {
				err = msgp.WrapError(err, zb0004, "Version")
				return
			}
			// write "content"
			err = en.Append(0xa7, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74)
			if err != nil {
				return
			}
			err = z[zb0004].Content.EncodeMsg(en)
			if err != nil {
				err = msgp.WrapError(err, zb0004, "Content")
				return
			}
		}
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z ciVisibilityEvents) Msgsize() (s int) {
	s = msgp.ArrayHeaderSize
	for zb0004 := range z {
		if z[zb0004] == nil {
			s += msgp.NilSize
		} else {
			s += 1 + 5 + msgp.StringPrefixSize + len(z[zb0004].Type) + 8 + msgp.Int32Size + 8 + z[zb0004].Content.Msgsize()
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *tslvSpan) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "test_session_id":
			z.SessionId, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "SessionId")
				return
			}
		case "test_module_id":
			z.ModuleId, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "ModuleId")
				return
			}
		case "test_suite_id":
			z.SuiteId, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "SuiteId")
				return
			}
		case "itr_correlation_id":
			z.CorrelationId, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "CorrelationId")
				return
			}
		case "name":
			z.Name, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Name")
				return
			}
		case "service":
			z.Service, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Service")
				return
			}
		case "resource":
			z.Resource, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Resource")
				return
			}
		case "type":
			z.Type, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Type")
				return
			}
		case "start":
			z.Start, err = dc.ReadInt64()
			if err != nil {
				err = msgp.WrapError(err, "Start")
				return
			}
		case "duration":
			z.Duration, err = dc.ReadInt64()
			if err != nil {
				err = msgp.WrapError(err, "Duration")
				return
			}
		case "span_id":
			z.SpanID, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "SpanID")
				return
			}
		case "trace_id":
			z.TraceID, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "TraceID")
				return
			}
		case "parent_id":
			z.ParentID, err = dc.ReadUint64()
			if err != nil {
				err = msgp.WrapError(err, "ParentID")
				return
			}
		case "error":
			z.Error, err = dc.ReadInt32()
			if err != nil {
				err = msgp.WrapError(err, "Error")
				return
			}
		case "meta":
			var zb0002 uint32
			zb0002, err = dc.ReadMapHeader()
			if err != nil {
				err = msgp.WrapError(err, "Meta")
				return
			}
			if z.Meta == nil {
				z.Meta = make(map[string]string, zb0002)
			} else if len(z.Meta) > 0 {
				for key := range z.Meta {
					delete(z.Meta, key)
				}
			}
			for zb0002 > 0 {
				zb0002--
				var za0001 string
				var za0002 string
				za0001, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Meta")
					return
				}
				za0002, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Meta", za0001)
					return
				}
				z.Meta[za0001] = za0002
			}
		case "metrics":
			var zb0003 uint32
			zb0003, err = dc.ReadMapHeader()
			if err != nil {
				err = msgp.WrapError(err, "Metrics")
				return
			}
			if z.Metrics == nil {
				z.Metrics = make(map[string]float64, zb0003)
			} else if len(z.Metrics) > 0 {
				for key := range z.Metrics {
					delete(z.Metrics, key)
				}
			}
			for zb0003 > 0 {
				zb0003--
				var za0003 string
				var za0004 float64
				za0003, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Metrics")
					return
				}
				za0004, err = dc.ReadFloat64()
				if err != nil {
					err = msgp.WrapError(err, "Metrics", za0003)
					return
				}
				z.Metrics[za0003] = za0004
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *tslvSpan) EncodeMsg(en *msgp.Writer) (err error) {
	// omitempty: check for empty values
	zb0001Len := uint32(16)
	var zb0001Mask uint16 /* 16 bits */
	_ = zb0001Mask
	if z.SessionId == 0 {
		zb0001Len--
		zb0001Mask |= 0x1
	}
	if z.ModuleId == 0 {
		zb0001Len--
		zb0001Mask |= 0x2
	}
	if z.SuiteId == 0 {
		zb0001Len--
		zb0001Mask |= 0x4
	}
	if z.CorrelationId == "" {
		zb0001Len--
		zb0001Mask |= 0x8
	}
	if z.SpanID == 0 {
		zb0001Len--
		zb0001Mask |= 0x400
	}
	if z.TraceID == 0 {
		zb0001Len--
		zb0001Mask |= 0x800
	}
	if z.ParentID == 0 {
		zb0001Len--
		zb0001Mask |= 0x1000
	}
	if z.Meta == nil {
		zb0001Len--
		zb0001Mask |= 0x4000
	}
	if z.Metrics == nil {
		zb0001Len--
		zb0001Mask |= 0x8000
	}
	// variable map header, size zb0001Len
	err = en.WriteMapHeader(zb0001Len)
	if err != nil {
		return
	}
	if zb0001Len == 0 {
		return
	}
	if (zb0001Mask & 0x1) == 0 { // if not empty
		// write "test_session_id"
		err = en.Append(0xaf, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64)
		if err != nil {
			return
		}
		err = en.WriteUint64(z.SessionId)
		if err != nil {
			err = msgp.WrapError(err, "SessionId")
			return
		}
	}
	if (zb0001Mask & 0x2) == 0 { // if not empty
		// write "test_module_id"
		err = en.Append(0xae, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x5f, 0x69, 0x64)
		if err != nil {
			return
		}
		err = en.WriteUint64(z.ModuleId)
		if err != nil {
			err = msgp.WrapError(err, "ModuleId")
			return
		}
	}
	if (zb0001Mask & 0x4) == 0 { // if not empty
		// write "test_suite_id"
		err = en.Append(0xad, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x73, 0x75, 0x69, 0x74, 0x65, 0x5f, 0x69, 0x64)
		if err != nil {
			return
		}
		err = en.WriteUint64(z.SuiteId)
		if err != nil {
			err = msgp.WrapError(err, "SuiteId")
			return
		}
	}
	if (zb0001Mask & 0x8) == 0 { // if not empty
		// write "itr_correlation_id"
		err = en.Append(0xb2, 0x69, 0x74, 0x72, 0x5f, 0x63, 0x6f, 0x72, 0x72, 0x65, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64)
		if err != nil {
			return
		}
		err = en.WriteString(z.CorrelationId)
		if err != nil {
			err = msgp.WrapError(err, "CorrelationId")
			return
		}
	}
	// write "name"
	err = en.Append(0xa4, 0x6e, 0x61, 0x6d, 0x65)
	if err != nil {
		return
	}
	err = en.WriteString(z.Name)
	if err != nil {
		err = msgp.WrapError(err, "Name")
		return
	}
	// write "service"
	err = en.Append(0xa7, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65)
	if err != nil {
		return
	}
	err = en.WriteString(z.Service)
	if err != nil {
		err = msgp.WrapError(err, "Service")
		return
	}
	// write "resource"
	err = en.Append(0xa8, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65)
	if err != nil {
		return
	}
	err = en.WriteString(z.Resource)
	if err != nil {
		err = msgp.WrapError(err, "Resource")
		return
	}
	// write "type"
	err = en.Append(0xa4, 0x74, 0x79, 0x70, 0x65)
	if err != nil {
		return
	}
	err = en.WriteString(z.Type)
	if err != nil {
		err = msgp.WrapError(err, "Type")
		return
	}
	// write "start"
	err = en.Append(0xa5, 0x73, 0x74, 0x61, 0x72, 0x74)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.Start)
	if err != nil {
		err = msgp.WrapError(err, "Start")
		return
	}
	// write "duration"
	err = en.Append(0xa8, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.Duration)
	if err != nil {
		err = msgp.WrapError(err, "Duration")
		return
	}
	if (zb0001Mask & 0x400) == 0 { // if not empty
		// write "span_id"
		err = en.Append(0xa7, 0x73, 0x70, 0x61, 0x6e, 0x5f, 0x69, 0x64)
		if err != nil {
			return
		}
		err = en.WriteUint64(z.SpanID)
		if err != nil {
			err = msgp.WrapError(err, "SpanID")
			return
		}
	}
	if (zb0001Mask & 0x800) == 0 { // if not empty
		// write "trace_id"
		err = en.Append(0xa8, 0x74, 0x72, 0x61, 0x63, 0x65, 0x5f, 0x69, 0x64)
		if err != nil {
			return
		}
		err = en.WriteUint64(z.TraceID)
		if err != nil {
			err = msgp.WrapError(err, "TraceID")
			return
		}
	}
	if (zb0001Mask & 0x1000) == 0 { // if not empty
		// write "parent_id"
		err = en.Append(0xa9, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64)
		if err != nil {
			return
		}
		err = en.WriteUint64(z.ParentID)
		if err != nil {
			err = msgp.WrapError(err, "ParentID")
			return
		}
	}
	// write "error"
	err = en.Append(0xa5, 0x65, 0x72, 0x72, 0x6f, 0x72)
	if err != nil {
		return
	}
	err = en.WriteInt32(z.Error)
	if err != nil {
		err = msgp.WrapError(err, "Error")
		return
	}
	if (zb0001Mask & 0x4000) == 0 { // if not empty
		// write "meta"
		err = en.Append(0xa4, 0x6d, 0x65, 0x74, 0x61)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(z.Meta)))
		if err != nil {
			err = msgp.WrapError(err, "Meta")
			return
		}
		for za0001, za0002 := range z.Meta {
			err = en.WriteString(za0001)
			if err != nil {
				err = msgp.WrapError(err, "Meta")
				return
			}
			err = en.WriteString(za0002)
			if err != nil {
				err = msgp.WrapError(err, "Meta", za0001)
				return
			}
		}
	}
	if (zb0001Mask & 0x8000) == 0 { // if not empty
		// write "metrics"
		err = en.Append(0xa7, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73)
		if err != nil {
			return
		}
		err = en.WriteMapHeader(uint32(len(z.Metrics)))
		if err != nil {
			err = msgp.WrapError(err, "Metrics")
			return
		}
		for za0003, za0004 := range z.Metrics {
			err = en.WriteString(za0003)
			if err != nil {
				err = msgp.WrapError(err, "Metrics")
				return
			}
			err = en.WriteFloat64(za0004)
			if err != nil {
				err = msgp.WrapError(err, "Metrics", za0003)
				return
			}
		}
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *tslvSpan) Msgsize() (s int) {
	s = 3 + 16 + msgp.Uint64Size + 15 + msgp.Uint64Size + 14 + msgp.Uint64Size + 19 + msgp.StringPrefixSize + len(z.CorrelationId) + 5 + msgp.StringPrefixSize + len(z.Name) + 8 + msgp.StringPrefixSize + len(z.Service) + 9 + msgp.StringPrefixSize + len(z.Resource) + 5 + msgp.StringPrefixSize + len(z.Type) + 6 + msgp.Int64Size + 9 + msgp.Int64Size + 8 + msgp.Uint64Size + 9 + msgp.Uint64Size + 10 + msgp.Uint64Size + 6 + msgp.Int32Size + 5 + msgp.MapHeaderSize
	if z.Meta != nil {
		for za0001, za0002 := range z.Meta {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + msgp.StringPrefixSize + len(za0002)
		}
	}
	s += 8 + msgp.MapHeaderSize
	if z.Metrics != nil {
		for za0003, za0004 := range z.Metrics {
			_ = za0004
			s += msgp.StringPrefixSize + len(za0003) + msgp.Float64Size
		}
	}
	return
}