package amino

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"

	"github.com/davecgh/go-spew/spew"
)

//----------------------------------------
// cdc.encodeReflectBinary

/*
This is the main entrypoint for encoding all types in binary form.  This
function calls encodeReflectBinary*, and generally those functions should
only call this one, for all overrides happen here.

The value may be a nil interface, but not a nil pointer.

The argument "bare" is ignored when the value is a primitive type, or a
byteslice or bytearray or generally a list type (except for unpacked lists,
which are more like structs).  EncodeByteSlice() is of course byte-length
prefixed, but EncodeTime() is not, as it is a struct.

For structs and struct-like things like unpacked lists, the "bare" argument
determines whether to include the length-prefix or not.

NOTE: Unlike encodeReflectJSON, rv may not be a pointer.  This is because the
binary representation of pointers depend on the context.  A nil pointer in the
context of a struct field is represented by its presence or absence in the
encoding bytes (w/ bare=false, which normally would produce 0x00), whereas in
the context of a list, (for amino 1.x anyways, which is constrained by proto3),
nil pointers and non-nil pointers to empty structs have the same representation
(0x00).  This is a Proto3 limitation -- choosing repeated fields as the method
of encoding lists is an unfortunate hack.  Amino2 will resolve this issue.

The following contracts apply to all similar encode methods.
CONTRACT: rv is not a pointer
CONTRACT: rv is valid.
*/
func (cdc *Codec) encodeReflectBinary(w io.Writer, info *TypeInfo, rv reflect.Value,
	fopts FieldOptions, bare bool) (err error) {
	if rv.Kind() == reflect.Ptr {
		// Whether to encode nil pointers as 0x00 or not at all depend on the
		// context, so pointers should be handled first by the caller.
		panic("not allowed to be called with a reflect.Ptr")
	}
	if !rv.IsValid() {
		panic("not allowed to be called with invalid / zero Value")
	}
	if printLog {
		spew.Printf("(E) encodeReflectBinary(info: %v, rv: %#v (%v), fopts: %v)\n",
			info, rv.Interface(), rv.Type(), fopts)
		defer func() {
			fmt.Printf("(E) -> err: %v\n", err)
		}()
	}

	// Handle the most special case, "well known".
	if info.IsBinaryWellKnownType {
		var ok bool
		ok, err = encodeReflectBinaryWellKnown(w, info, rv, fopts, bare)
		if ok || err != nil {
			return
		}
	}

	// Handle override if rv implements MarshalAmino.
	if info.IsAminoMarshaler {
		// First, encode rv into repr instance.
		var rrv reflect.Value
		var rinfo *TypeInfo = info.ReprType
		rrv, err = toReprObject(rv)
		if err != nil {
			return
		}
		// Then, encode the repr instance.
		err = cdc.encodeReflectBinary(w, rinfo, rrv, fopts, bare)
		return
	}

	switch info.Type.Kind() {

	//----------------------------------------
	// Complex

	case reflect.Interface:
		err = cdc.encodeReflectBinaryInterface(w, info, rv, fopts, bare)

	case reflect.Array:
		if info.Type.Elem().Kind() == reflect.Uint8 {
			err = cdc.encodeReflectBinaryByteArray(w, info, rv, fopts)
		} else if kind := info.Type.Elem().Kind(); kind == reflect.Slice || kind == reflect.Array {
			// for proto3 compatibility, we do not allow multidimensional arrays,
			// unless the elements involved are bytes (e.g. [][]byte)
			if info.Type.Elem().Elem().Kind() != reflect.Uint8 { // byte is an alias for uint8
				elem := info.Type.Elem()
				for {
					if elem.Kind() == reflect.Slice || elem.Kind() == reflect.Array {
						elem = elem.Elem()
						continue
					}
					if elem.Kind() == reflect.Uint8 { // byte is an alias for uint8
						break
					}
					err = errors.New("multidimensional arrays not allowed")
					break
				}
			}
		} else {
			err = cdc.encodeReflectBinaryList(w, info, rv, fopts, bare)
		}

	case reflect.Slice:
		switch info.Type.Elem().Kind() {

		case reflect.Uint8:
			err = cdc.encodeReflectBinaryByteSlice(w, info, rv, fopts)
		case reflect.Slice, reflect.Array:
			// for proto3 compatibility, we do not allow multidimensional slices,
			// unless the elements involved are bytes (e.g. [][]byte)
			elem := info.Type.Elem()
			for {
				if elem.Kind() == reflect.Slice || elem.Kind() == reflect.Array {
					elem = elem.Elem()
					continue
				}
				if elem.Kind() == reflect.Uint8 { // byte is an alias for uint8
					break
				}
				err = errors.New("multidimensional slices not allowed")
				break
			}
		default:
			err = cdc.encodeReflectBinaryList(w, info, rv, fopts, bare)
		}

	case reflect.Struct:
		err = cdc.encodeReflectBinaryStruct(w, info, rv, fopts, bare)

	//----------------------------------------
	// Signed

	case reflect.Int64:
		if fopts.BinFixed64 {
			err = EncodeInt64(w, rv.Int())
		} else {
			err = EncodeUvarint(w, uint64(rv.Int()))
		}

	case reflect.Int32:
		if fopts.BinFixed32 {
			err = EncodeInt32(w, int32(rv.Int()))
		} else {
			err = EncodeUvarint(w, uint64(rv.Int()))
		}

	case reflect.Int16:
		err = EncodeInt16(w, int16(rv.Int()))

	case reflect.Int8:
		err = EncodeInt8(w, int8(rv.Int()))

	case reflect.Int:
		err = EncodeUvarint(w, uint64(rv.Int()))

	//----------------------------------------
	// Unsigned

	case reflect.Uint64:
		if fopts.BinFixed64 {
			err = EncodeUint64(w, rv.Uint())
		} else {
			err = EncodeUvarint(w, rv.Uint())
		}

	case reflect.Uint32:
		if fopts.BinFixed32 {
			err = EncodeUint32(w, uint32(rv.Uint()))
		} else {
			err = EncodeUvarint(w, rv.Uint())
		}

	case reflect.Uint16:
		err = EncodeUint16(w, uint16(rv.Uint()))

	case reflect.Uint8:
		err = EncodeUint8(w, uint8(rv.Uint()))

	case reflect.Uint:
		err = EncodeUvarint(w, rv.Uint())

	//----------------------------------------
	// Misc

	case reflect.Bool:
		err = EncodeBool(w, rv.Bool())

	case reflect.Float64:
		if !fopts.Unsafe {
			err = errors.New("amino float* support requires `amino:\"unsafe\"`")
			return
		}
		err = EncodeFloat64(w, rv.Float())

	case reflect.Float32:
		if !fopts.Unsafe {
			err = errors.New("amino float* support requires `amino:\"unsafe\"`")
			return
		}
		err = EncodeFloat32(w, float32(rv.Float()))

	case reflect.String:
		err = EncodeString(w, rv.String())

	//----------------------------------------
	// Default

	default:
		panic(fmt.Sprintf("unsupported type %v", info.Type.Kind()))
	}

	return err
}

func (cdc *Codec) encodeReflectBinaryInterface(w io.Writer, iinfo *TypeInfo, rv reflect.Value,
	fopts FieldOptions, bare bool) (err error) {
	if printLog {
		fmt.Println("(e) encodeReflectBinaryInterface")
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}

	// Special case when rv is nil, write nothing or 0x00.
	if rv.IsNil() {
		return writeMaybeBare(w, nil, bare)
	}

	// Get concrete non-pointer reflect value & type.
	var crv = rv.Elem()
	var dcrv = crv
	var crvIsNilPtr = false
	var crvIsPtr = false
	if crv.Kind() == reflect.Ptr {
		crvIsNilPtr = crv.IsNil()
		crvIsPtr = true
		dcrv = crv.Elem()
	}
	if crvIsPtr && dcrv.Kind() == reflect.Interface {
		// See "MARKER: No interface-pointers" in codec.go
		panic("should not happen")
	}
	if crvIsPtr && crvIsNilPtr {
		panic(fmt.Sprintf("Illegal nil-pointer of type %v for registered interface %v. "+
			"For compatibility with other languages, nil-pointer interface values are forbidden.", dcrv.Type(), iinfo.Type))
	}

	// Get *TypeInfo for concrete type.
	var cinfo *TypeInfo
	cinfo, err = cdc.getTypeInfoWLock(crt)
	if err != nil {
		return
	}
	if !cinfo.Registered {
		err = fmt.Errorf("cannot encode unregistered concrete type %v", crt.Elem())
		return
	}

	// For Proto3 compatibility, encode interfaces as google.protobuf.Any
	// Write field #1, TypeURL
	buf := bytes.NewBuffer(nil)
	{
		fnum := uint32(1)
		err = encodeFieldNumberAndTyp3(buf, fnum, Typ3ByteLength)
		if err != nil {
			return
		}
		err = EncodeString(buf, cinfo.TypeURL)
		if err != nil {
			return
		}
	}
	// Write field #2, Value, if not empty/default.
	// writeFieldIfNotEmpty() is not a substitute for this slightly different
	// logic here, because we need to enforce that the value is a []byte type
	// as per google.protobuf.Any.
	{
		// google.protobuf.Any values must be a struct, or an unpacked list which
		// is indistinguishable from a struct.
		var buf2 = bytes.NewBuffer(nil)
		if !cinfo.IsStructOrUnpacked(fopts) {
			writeEmpty := false
			// Encode with an implicit struct, with a single field with number 1.
			// The type of this implicit field determines whether any
			// length-prefixing happens after the typ3 byte.
			// The second FieldOptions is empty, because this isn't a list of
			// Typ3ByteLength things, so however it is encoded, that option is no
			// longer needed.
			if err = cdc.writeFieldIfNotEmpty(buf2, 1, cinfo, FieldOptions{}, FieldOptions{}, dcrv, writeEmpty); err != nil {
				return
			}
		} else {
			// The passed in BinFieldNum is only relevant for when the type is to
			// be encoded unpacked (elements are Typ3ByteLength).  In that case,
			// encodeReflectBinary will repeat the field number as set here, as if
			// encoded with an implicit struct.
			err = cdc.encodeReflectBinary(buf2, cinfo, dcrv, FieldOptions{BinFieldNum: 1}, true)
			if err != nil {
				return
			}
		}
		bz2 := buf2.Bytes()
		if len(bz2) == 0 || len(bz2) == 1 && bz2[0] == 0x00 {
			// Do not write
		} else {
			// Write
			fnum := uint32(2)
			err = encodeFieldNumberAndTyp3(buf, fnum, Typ3ByteLength)
			if err != nil {
				return
			}
			err = EncodeByteSlice(buf, bz2)
			if err != nil {
				return
			}
		}
	}

	return writeMaybeBare(w, buf.Bytes(), bare)
}

func (cdc *Codec) encodeReflectBinaryByteArray(w io.Writer, info *TypeInfo, rv reflect.Value,
	fopts FieldOptions) (err error) {
	ert := info.Type.Elem()
	if ert.Kind() != reflect.Uint8 {
		panic("should not happen")
	}
	length := info.Type.Len()

	// If rv is an interface, get the elem.

	// Get byteslice.
	var byteslice []byte
	if rv.CanAddr() {
		byteslice = rv.Slice(0, length).Bytes()
	} else {
		byteslice = make([]byte, length)
		reflect.Copy(reflect.ValueOf(byteslice), rv) // XXX: looks expensive!
	}

	// Write byte-length prefixed byteslice.
	err = EncodeByteSlice(w, byteslice)
	return
}

func (cdc *Codec) encodeReflectBinaryList(w io.Writer, info *TypeInfo, rv reflect.Value,
	fopts FieldOptions, bare bool) (err error) {
	if printLog {
		fmt.Println("(e) encodeReflectBinaryList")
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}
	ert := info.Type.Elem()
	if ert.Kind() == reflect.Uint8 {
		panic("should not happen")
	}
	einfo, err := cdc.getTypeInfoWLock(ert)
	if err != nil {
		return
	}

	// Proto3 byte-length prefixing incurs alloc cost on the encoder.
	// Here we incur it for unpacked form for ease of dev.
	buf := bytes.NewBuffer(nil)

	// If elem is not already a ByteLength type, write in packed form.
	// This is a Proto wart due to Proto backwards compatibility issues.
	// Amino2 will probably migrate to use the List typ3.  Please?  :)
	typ3 := einfo.GetTyp3(fopts)
	if typ3 != Typ3ByteLength {
		// Write elems in packed form.
		for i := 0; i < rv.Len(); i++ {
			var erv = rv.Index(i)
			// If pointer, get dereferenced element value (or zero).
			if ert.Kind() == reflect.Ptr {
				if erv.IsNil() {
					erv = reflect.New(ert.Elem()).Elem()
				}
			}
			// Write the element value.
			err = cdc.encodeReflectBinary(buf, einfo, erv, fopts, false)
			if err != nil {
				return
			}
		}
	} else { // typ3 == Typ3ByteLength
		// NOTE: ert is for the element value, while einfo.Type is dereferenced.
		isErtStructPointer := ert.Kind() == reflect.Ptr && einfo.Type.Kind() == reflect.Struct

		// Write elems in unpacked form.
		for i := 0; i < rv.Len(); i++ {
			// Write elements as repeated fields of the parent struct.
			err = encodeFieldNumberAndTyp3(buf, fopts.BinFieldNum, Typ3ByteLength)
			if err != nil {
				return
			}
			// Get dereferenced element value and info.
			var erv = rv.Index(i)
			if isDefaultValue(erv) {
				// Special case if:
				//  - erv is a struct pointer and
				//  - field option has EmptyElements set
				if isErtStructPointer && fopts.EmptyElements {
					// NOTE: Not sure what to do here, but for future-proofing,
					// we explicitly fail on nil pointers, just like
					// Proto3's Golang client does.
					// This also makes it easier to upgrade to Amino2
					// which would enable the encoding of nil structs.
					return errors.New("nil struct pointers not supported when empty_elements field tag is set")
				}
				// Nothing to encode, so the length is 0.
				err = EncodeByte(buf, byte(0x00))
				if err != nil {
					return
				}
			} else {
				// Write the element value as a ByteLength prefixed.

				// NOTE: In case of any (nested) inner lists in unpacked form,
				// we again pass in BinFieldNum=1, but each inner list of field
				// num = 1 items will still be byte-length prefixed, so
				// multidimensional lists are still supported.
				//
				// In proto3 this would be resented with an auto-generated
				// message which holds a list at field number 1.
				efopts := fopts
				efopts.BinFieldNum = 1
				err = cdc.encodeReflectBinary(buf, einfo, erv, efopts, false)
				if err != nil {
					return
				}
			}
		}
	}

	return writeMaybeBare(w, buf.Bytes(), bare)
}

// CONTRACT: info.Type.Elem().Kind() == reflect.Uint8
func (cdc *Codec) encodeReflectBinaryByteSlice(w io.Writer, info *TypeInfo, rv reflect.Value,
	fopts FieldOptions) (err error) {
	if printLog {
		fmt.Println("(e) encodeReflectBinaryByteSlice")
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}
	ert := info.Type.Elem()
	if ert.Kind() != reflect.Uint8 {
		panic("should not happen")
	}

	// Write byte-length prefixed byte-slice.
	var byteslice = rv.Bytes()
	err = EncodeByteSlice(w, byteslice)
	return
}

func (cdc *Codec) encodeReflectBinaryStruct(w io.Writer, info *TypeInfo, rv reflect.Value,
	fopts FieldOptions, bare bool) (err error) {
	if printLog {
		fmt.Println("(e) encodeReflectBinaryBinaryStruct")
		defer func() {
			fmt.Printf("(e) -> err: %v\n", err)
		}()
	}

	// Proto3 incurs a cost in writing non-root structs.
	// Here we incur it for root structs as well for ease of dev.
	buf := bytes.NewBuffer(nil)

	for _, field := range info.Fields {
		// Get type info for field.
		var finfo *TypeInfo
		finfo, err = cdc.getTypeInfoWLock(field.Type)
		if err != nil {
			return
		}
		// Get dereferenced field value and info.
		var frv = rv.Field(field.Index)
		var dfrv = frv
		var frvIsPtr = frv.Kind() == reflect.Ptr
		if frvIsPtr {
			dfrv = frv.Elem()
		}
		if isDefaultValue(frv) && !field.WriteEmpty {
			// Do not encode default value fields
			// (except when `amino:"write_empty"` is set).
			continue
		}
		if field.UnpackedList {
			// Write repeated field entries for each list item.
			err = cdc.encodeReflectBinaryList(buf, finfo, dfrv, field.FieldOptions, true)
			if err != nil {
				return
			}
		} else {
			// write empty if explicitly set or if this is a pointer:
			writeEmpty := field.WriteEmpty || frvIsPtr
			err = cdc.writeFieldIfNotEmpty(buf, field.BinFieldNum, finfo, fopts, field.FieldOptions, dfrv, writeEmpty)
			if err != nil {
				return
			}
		}
	}

	return writeMaybeBare(w, buf.Bytes(), bare)
}

//----------------------------------------
// Misc.

// Write field key.
func encodeFieldNumberAndTyp3(w io.Writer, num uint32, typ Typ3) (err error) {
	if (typ & 0xF8) != 0 {
		panic(fmt.Sprintf("invalid Typ3 byte %v", typ))
	}
	if num > (1<<29 - 1) {
		panic(fmt.Sprintf("invalid field number %v", num))
	}

	// Pack Typ3 and field number.
	var value64 = (uint64(num) << 3) | uint64(typ)

	// Write uvarint value for field and Typ3.
	var buf [10]byte
	n := binary.PutUvarint(buf[:], value64)
	_, err = w.Write(buf[0:n])
	return
}

func (cdc *Codec) writeFieldIfNotEmpty(
	buf *bytes.Buffer,
	fieldNum uint32,
	finfo *TypeInfo,
	structsFopts FieldOptions, // the wrapping struct's FieldOptions if any
	fieldOpts FieldOptions, // the field's FieldOptions
	derefedVal reflect.Value,
	isWriteEmpty bool,
) error {
	lBeforeKey := buf.Len()
	// Write field key (number and type).
	err := encodeFieldNumberAndTyp3(buf, fieldNum, finfo.GetTyp3(fieldOpts))
	if err != nil {
		return err
	}
	lBeforeValue := buf.Len()

	// Write field value from rv.
	err = cdc.encodeReflectBinary(buf, finfo, derefedVal, fieldOpts, false)
	if err != nil {
		return err
	}
	lAfterValue := buf.Len()

	if !isWriteEmpty && lBeforeValue == lAfterValue-1 && buf.Bytes()[buf.Len()-1] == 0x00 {
		// rollback typ3/fieldnum and last byte if
		// not a pointer and empty:
		buf.Truncate(lBeforeKey)
	}
	return nil
}

// NOTE: This is slightly less efficient than recursing as in the
// implementation for encodeReflectBinaryWelKnown.
func writeMaybeBare(w io.Writer, bz []byte, bare bool) (err error) {
	// Special case
	if len(bz) == 0 {
		if bare {
			return
		} else {
			_, err = w.Write([]byte{0x00})
		}
		return
	}
	// General case
	if bare {
		// Write byteslice without byte-length prefixing.
		_, err = w.Write(bz)
	} else {
		// Write byte-length prefixed byteslice.
		err = EncodeByteSlice(w, bz)
	}
	return err
}