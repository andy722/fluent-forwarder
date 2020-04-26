package wal

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *MetaContent) DecodeMsg(dc *msgp.Reader) (err error) {
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
		case "Readers":
			var zb0002 uint32
			zb0002, err = dc.ReadMapHeader()
			if err != nil {
				err = msgp.WrapError(err, "Readers")
				return
			}
			if z.Readers == nil {
				z.Readers = make(map[string]Offset, zb0002)
			} else if len(z.Readers) > 0 {
				for key := range z.Readers {
					delete(z.Readers, key)
				}
			}
			for zb0002 > 0 {
				zb0002--
				var za0001 string
				var za0002 Offset
				za0001, err = dc.ReadString()
				if err != nil {
					err = msgp.WrapError(err, "Readers")
					return
				}
				var zb0003 uint32
				zb0003, err = dc.ReadMapHeader()
				if err != nil {
					err = msgp.WrapError(err, "Readers", za0001)
					return
				}
				for zb0003 > 0 {
					zb0003--
					field, err = dc.ReadMapKeyPtr()
					if err != nil {
						err = msgp.WrapError(err, "Readers", za0001)
						return
					}
					switch msgp.UnsafeString(field) {
					case "Segment":
						{
							var zb0004 int64
							zb0004, err = dc.ReadInt64()
							if err != nil {
								err = msgp.WrapError(err, "Readers", za0001, "Segment")
								return
							}
							za0002.Segment = SegmentID(zb0004)
						}
					case "Position":
						za0002.Position, err = dc.ReadInt64()
						if err != nil {
							err = msgp.WrapError(err, "Readers", za0001, "Position")
							return
						}
					default:
						err = dc.Skip()
						if err != nil {
							err = msgp.WrapError(err, "Readers", za0001)
							return
						}
					}
				}
				z.Readers[za0001] = za0002
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
func (z *MetaContent) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 1
	// write "Readers"
	err = en.Append(0x81, 0xa7, 0x52, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73)
	if err != nil {
		return
	}
	err = en.WriteMapHeader(uint32(len(z.Readers)))
	if err != nil {
		err = msgp.WrapError(err, "Readers")
		return
	}
	for za0001, za0002 := range z.Readers {
		err = en.WriteString(za0001)
		if err != nil {
			err = msgp.WrapError(err, "Readers")
			return
		}
		// map header, size 2
		// write "Segment"
		err = en.Append(0x82, 0xa7, 0x53, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74)
		if err != nil {
			return
		}
		err = en.WriteInt64(int64(za0002.Segment))
		if err != nil {
			err = msgp.WrapError(err, "Readers", za0001, "Segment")
			return
		}
		// write "Position"
		err = en.Append(0xa8, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e)
		if err != nil {
			return
		}
		err = en.WriteInt64(za0002.Position)
		if err != nil {
			err = msgp.WrapError(err, "Readers", za0001, "Position")
			return
		}
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *MetaContent) Msgsize() (s int) {
	s = 1 + 8 + msgp.MapHeaderSize
	if z.Readers != nil {
		for za0001, za0002 := range z.Readers {
			_ = za0002
			s += msgp.StringPrefixSize + len(za0001) + 1 + 8 + msgp.Int64Size + 9 + msgp.Int64Size
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z Offset) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 2
	// string "Segment"
	o = append(o, 0x82, 0xa7, 0x53, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74)
	o = msgp.AppendInt64(o, int64(z.Segment))
	// string "Position"
	o = append(o, 0xa8, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e)
	o = msgp.AppendInt64(o, z.Position)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Offset) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "Segment":
			{
				var zb0002 int64
				zb0002, bts, err = msgp.ReadInt64Bytes(bts)
				if err != nil {
					err = msgp.WrapError(err, "Segment")
					return
				}
				z.Segment = SegmentID(zb0002)
			}
		case "Position":
			z.Position, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Position")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z Offset) Msgsize() (s int) {
	s = 1 + 8 + msgp.Int64Size + 9 + msgp.Int64Size
	return
}

// DecodeMsg implements msgp.Decodable
func (z *SegmentID) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zb0001 int64
		zb0001, err = dc.ReadInt64()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = SegmentID(zb0001)
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z SegmentID) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteInt64(int64(z))
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z SegmentID) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendInt64(o, int64(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *SegmentID) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zb0001 int64
		zb0001, bts, err = msgp.ReadInt64Bytes(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		(*z) = SegmentID(zb0001)
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z SegmentID) Msgsize() (s int) {
	s = msgp.Int64Size
	return
}
