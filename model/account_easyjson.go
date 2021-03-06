// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package model

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson349b126bDecodeConcursModel(in *jlexer.Lexer, out *Premium) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "start":
			out.Start = int64(in.Int64())
		case "finish":
			out.Finish = int64(in.Int64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson349b126bEncodeConcursModel(out *jwriter.Writer, in Premium) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"start\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.Start))
	}
	{
		const prefix string = ",\"finish\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.Finish))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Premium) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson349b126bEncodeConcursModel(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Premium) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson349b126bEncodeConcursModel(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Premium) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson349b126bDecodeConcursModel(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Premium) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson349b126bDecodeConcursModel(l, v)
}
func easyjson349b126bDecodeConcursModel1(in *jlexer.Lexer, out *Account) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.ID = int(in.Int())
		case "email":
			out.Email = string(in.String())
		case "fname":
			out.FName = string(in.String())
		case "sname":
			out.SName = string(in.String())
		case "phone":
			out.Phone = string(in.String())
		case "sex":
			out.Sex = string(in.String())
		case "birth":
			out.Birth = int64(in.Int64())
		case "country":
			out.Country = string(in.String())
		case "city":
			out.City = string(in.String())
		case "joined":
			out.Joined = int64(in.Int64())
		case "status":
			out.Status = string(in.String())
		case "interests":
			if in.IsNull() {
				in.Skip()
				out.Interests = nil
			} else {
				in.Delim('[')
				if out.Interests == nil {
					if !in.IsDelim(']') {
						out.Interests = make([]string, 0, 4)
					} else {
						out.Interests = []string{}
					}
				} else {
					out.Interests = (out.Interests)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Interests = append(out.Interests, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "premium":
			(out.Premium).UnmarshalEasyJSON(in)
		case "likes":
			if in.IsNull() {
				in.Skip()
				out.Likes = nil
			} else {
				in.Delim('[')
				if out.Likes == nil {
					if !in.IsDelim(']') {
						out.Likes = make([]Like, 0, 2)
					} else {
						out.Likes = []Like{}
					}
				} else {
					out.Likes = (out.Likes)[:0]
				}
				for !in.IsDelim(']') {
					var v2 Like
					easyjson349b126bDecodeConcursModel2(in, &v2)
					out.Likes = append(out.Likes, v2)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson349b126bEncodeConcursModel1(out *jwriter.Writer, in Account) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int(int(in.ID))
	}
	{
		const prefix string = ",\"email\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"fname\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.FName))
	}
	{
		const prefix string = ",\"sname\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.SName))
	}
	{
		const prefix string = ",\"phone\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Phone))
	}
	{
		const prefix string = ",\"sex\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Sex))
	}
	{
		const prefix string = ",\"birth\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.Birth))
	}
	{
		const prefix string = ",\"country\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Country))
	}
	{
		const prefix string = ",\"city\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.City))
	}
	{
		const prefix string = ",\"joined\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.Joined))
	}
	{
		const prefix string = ",\"status\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Status))
	}
	{
		const prefix string = ",\"interests\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Interests == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v3, v4 := range in.Interests {
				if v3 > 0 {
					out.RawByte(',')
				}
				out.String(string(v4))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"premium\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		(in.Premium).MarshalEasyJSON(out)
	}
	{
		const prefix string = ",\"likes\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Likes == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v5, v6 := range in.Likes {
				if v5 > 0 {
					out.RawByte(',')
				}
				easyjson349b126bEncodeConcursModel2(out, v6)
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Account) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson349b126bEncodeConcursModel1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Account) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson349b126bEncodeConcursModel1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Account) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson349b126bDecodeConcursModel1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Account) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson349b126bDecodeConcursModel1(l, v)
}
func easyjson349b126bDecodeConcursModel2(in *jlexer.Lexer, out *Like) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "ts":
			out.Ts = float64(in.Float64())
		case "id":
			out.ID = int64(in.Int64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson349b126bEncodeConcursModel2(out *jwriter.Writer, in Like) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"ts\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Float64(float64(in.Ts))
	}
	{
		const prefix string = ",\"id\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.ID))
	}
	out.RawByte('}')
}
