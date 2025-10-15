package iso

import (
	"fmt"

	"github.com/moov-io/iso8583"
	"github.com/moov-io/iso8583/encoding"
	"github.com/moov-io/iso8583/field"
	"github.com/moov-io/iso8583/padding"
	"github.com/moov-io/iso8583/prefix"
)

var Spec87Hex *iso8583.MessageSpec = &iso8583.MessageSpec{
	Fields: map[int]field.Field{
		0: field.NewString(&field.Spec{
			Length:      4,
			Description: "Message Type Indicator",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		1: field.NewBitmap(&field.Spec{
			Description: "Bitmap",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.Hex.Fixed,
		}),
		2: field.NewString(&field.Spec{
			Length:      19,
			Description: "Primary Account Number",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LL,
		}),
		3: field.NewString(&field.Spec{
			Length:      6,
			Description: "Processing Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
		}),
		4: field.NewString(&field.Spec{
			Length:      12,
			Description: "Transaction Amount",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
		}),
		5: field.NewString(&field.Spec{
			Length:      12,
			Description: "Settlement Amount",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
		}),
		6: field.NewString(&field.Spec{
			Length:      12,
			Description: "Billing Amount",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
		}),
		7: field.NewString(&field.Spec{
			Length:      10,
			Description: "Transmission Date & Time",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		8: field.NewString(&field.Spec{
			Length:      8,
			Description: "Billing Fee Amount",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		9: field.NewString(&field.Spec{
			Length:      8,
			Description: "Settlement Conversion Rate",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		10: field.NewString(&field.Spec{
			Length:      8,
			Description: "Cardholder Billing Conversion Rate",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		11: field.NewString(&field.Spec{
			Length:      6,
			Description: "Systems Trace Audit Number (STAN)",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes, nil
			}),
		}),
		12: field.NewString(&field.Spec{
			Length:      6,
			Description: "Local Transaction Time",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		13: field.NewString(&field.Spec{
			Length:      4,
			Description: "Local Transaction Date",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		14: field.NewString(&field.Spec{
			Length:      4,
			Description: "Expiration Date",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		15: field.NewString(&field.Spec{
			Length:      4,
			Description: "Settlement Date",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		16: field.NewString(&field.Spec{
			Length:      4,
			Description: "Currency Conversion Date",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		17: field.NewString(&field.Spec{
			Length:      4,
			Description: "Capture Date",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		18: field.NewString(&field.Spec{
			Length:      4,
			Description: "Merchant Type",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		19: field.NewString(&field.Spec{
			Length:      3,
			Description: "Acquiring Institution Country Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length+1)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length + 1

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		20: field.NewString(&field.Spec{
			Length:      3,
			Description: "PAN Extended Country Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length+1)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length + 1

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		21: field.NewString(&field.Spec{
			Length:      3,
			Description: "Forwarding Institution Country Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length+1)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length + 1

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		22: field.NewString(&field.Spec{
			Length:      3,
			Description: "Point of Sale (POS) Entry Mode",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length+1)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length + 1

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		23: field.NewString(&field.Spec{
			Length:      3,
			Description: "Card Sequence Number (CSN)",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length+1)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length + 1

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		24: field.NewString(&field.Spec{
			Length:      3,
			Description: "Function Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length+1)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length + 1

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		25: field.NewString(&field.Spec{
			Length:      2,
			Description: "Point of Service Condition Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		26: field.NewString(&field.Spec{
			Length:      2,
			Description: "Point of Service PIN Capture Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		27: field.NewString(&field.Spec{
			Length:      1,
			Description: "Authorizing Identification Response Length",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length+1)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length + 1

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		28: field.NewString(&field.Spec{
			Length:      8,
			Description: "Transaction Fee Amount",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		29: field.NewString(&field.Spec{
			Length:      8,
			Description: "Settlement Fee Amount",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		30: field.NewString(&field.Spec{
			Length:      8,
			Description: "Transaction Processing Fee Amount",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		31: field.NewString(&field.Spec{
			Length:      8,
			Description: "Settlement Processing Fee Amount",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		32: field.NewString(&field.Spec{
			Length:      11,
			Description: "Acquiring Institution Identification Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LL,
			Pad:         padding.Right('F'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				// if spec.Pad != nil {
				// 	value = spec.Pad.Pad(value, spec.Length)
				// }

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				// if len(encodedValue) == maxLength {
				// 	encodedValue = spec.Pad.Pad(encodedValue, maxLength+1)
				// }
				if len(encodedValue)%2 != 0 {
					encodedValue = spec.Pad.Pad(encodedValue, len(encodedValue)+1)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if valueLength%2 != 0 {
					prefBytes = prefBytes + 1
				}

				return value, read + prefBytes, nil
			}),
		}),
		33: field.NewString(&field.Spec{
			Length:      11,
			Description: "Forwarding Institution Identification Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LL,
			Pad:         padding.Right('F'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				// if spec.Pad != nil {
				// 	value = spec.Pad.Pad(value, spec.Length)
				// }

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				if len(encodedValue) == maxLength {
					encodedValue = spec.Pad.Pad(encodedValue, maxLength+1)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if valueLength == maxEncodedValueLength {
					prefBytes = prefBytes + 1
				}

				return value, read + prefBytes, nil
			}),
		}),
		34: field.NewString(&field.Spec{
			Length:      28,
			Description: "Extended Primary Account Number",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LL,
		}),
		35: field.NewString(&field.Spec{
			Length:      37,
			Description: "Track 2 Data",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LL,
			Pad:         padding.Right('F'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				// if spec.Pad != nil {
				// 	value = spec.Pad.Pad(value, spec.Length)
				// }

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				if len(encodedValue) == maxLength {
					encodedValue = spec.Pad.Pad(encodedValue, maxLength+1)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if valueLength == maxEncodedValueLength {
					prefBytes = prefBytes + 1
				}

				return value, read + prefBytes, nil
			}),
		}),
		36: field.NewString(&field.Spec{
			Length:      104,
			Description: "Track 3 Data",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LLL,
		}),
		37: field.NewString(&field.Spec{
			Length:      12,
			Description: "Retrieval Reference Number",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.Fixed,
		}),
		38: field.NewString(&field.Spec{
			Length:      6,
			Description: "Authorization Identification Response",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.Fixed,
		}),
		39: field.NewString(&field.Spec{
			Length:      2,
			Description: "Response Code",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.Fixed,
		}),
		40: field.NewString(&field.Spec{
			Length:      3,
			Description: "Service Restriction Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length+1)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length + 1

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		41: field.NewString(&field.Spec{
			Length:      8,
			Description: "Card Acceptor Terminal Identification",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.Fixed,
		}),
		42: field.NewString(&field.Spec{
			Length:      15,
			Description: "Card Acceptor Identification Code",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.Fixed,
		}),
		43: field.NewString(&field.Spec{
			Length:      40,
			Description: "Card Acceptor Name/Location",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Right(' '),
		}),
		44: field.NewString(&field.Spec{
			Length:      99,
			Description: "Additional Data",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.LL,
		}),
		45: field.NewString(&field.Spec{
			Length:      76,
			Description: "Track 1 Data",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LL,
		}),
		46: field.NewString(&field.Spec{
			Length:      999,
			Description: "Additional data (ISO)",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		47: field.NewString(&field.Spec{
			Length:      999,
			Description: "Additional data (National)",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		48: field.NewString(&field.Spec{
			Length:      999,
			Description: "Additional data (Private)",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		49: field.NewString(&field.Spec{
			Length:      3,
			Description: "Transaction Currency Code",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.Fixed,
		}),
		50: field.NewString(&field.Spec{
			Length:      3,
			Description: "Settlement Currency Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length+1)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length + 1

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		51: field.NewString(&field.Spec{
			Length:      3,
			Description: "Cardholder Billing Currency Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length+1)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length + 1

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		52: field.NewString(&field.Spec{
			Length:      16,
			Description: "PIN Data",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		53: field.NewString(&field.Spec{
			Length:      16,
			Description: "Security Related Control Information",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		54: field.NewString(&field.Spec{
			Length:      120,
			Description: "Additional Amounts",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength * 2

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		55: field.NewString(&field.Spec{
			Length:      999,
			Description: "ICC Data – EMV Having Multiple Tags",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength * 2

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		56: field.NewString(&field.Spec{
			Length:      999,
			Description: "Reserved (ISO)",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		57: field.NewString(&field.Spec{
			Length:      999,
			Description: "Reserved (National)",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		58: field.NewString(&field.Spec{
			Length:      999,
			Description: "Reserved (National)",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)-1)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}
				zero := make([]byte, 1)
				lengthPrefix = append(zero, lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		59: field.NewString(&field.Spec{
			Length:      999,
			Description: "Reserved (National)",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		60: field.NewString(&field.Spec{
			Length:      999,
			Description: "Reserved (National)",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		61: field.NewString(&field.Spec{
			Length:      999,
			Description: "Reserved (Private)",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		62: field.NewString(&field.Spec{
			Length:      999,
			Description: "Reserved (Private)",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		63: field.NewString(&field.Spec{
			Length:      999,
			Description: "Reserved (Private)",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		64: field.NewString(&field.Spec{
			Length:      8,
			Description: "Message Authentication Code (MAC)",
			Enc:         encoding.ASCII,
			Pref:        prefix.Hex.Fixed,
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength * 2

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes, nil
			}),
		}),
		70: field.NewString(&field.Spec{
			Length:      3,
			Description: "Network management information code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length+1)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length + 1

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
		90: field.NewString(&field.Spec{
			Length:      42,
			Description: "Original Data Elements",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		98: field.NewString(&field.Spec{
			Length:      25,
			Description: "Original Data Elements",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.Fixed,
		}),
		100: field.NewString(&field.Spec{
			Length:      11,
			Description: "Original Data Elements",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LL,
			Pad:         padding.Right('F'),
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				// if spec.Pad != nil {
				// 	value = spec.Pad.Pad(value, spec.Length)
				// }

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue))
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				if len(encodedValue)%2 != 0 {
					encodedValue = spec.Pad.Pad(encodedValue, len(encodedValue)+1)
				}

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if valueLength%2 != 0 {
					prefBytes = prefBytes + 1
				}

				return value, read + prefBytes, nil
			}),
		}),
		102: field.NewString(&field.Spec{
			Length:      19,
			Description: "Original Data Elements",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.LL,
		}),
		103: field.NewString(&field.Spec{
			Length:      19,
			Description: "Original Data Elements",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.LL,
		}),
		126: field.NewString(&field.Spec{
			Length:      999,
			Description: "Original Data Elements",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.ASCII.LLL,
			Packer: field.PackerFunc(func(value []byte, spec *field.Spec) ([]byte, error) {
				if spec.Pad != nil {
					value = spec.Pad.Pad(value, spec.Length)
				}

				encodedValue, err := spec.Enc.Encode(value)
				if err != nil {
					return nil, fmt.Errorf("failed to encode content: %w", err)
				}

				// Encode the length of the packed data, not the length of the value
				maxLength := spec.Length

				// Encode the length of the encoded value
				lengthPrefix, err := spec.Pref.EncodeLength(maxLength, len(encodedValue)/2)
				if err != nil {
					return nil, fmt.Errorf("failed to encode length: %w", err)
				}

				lengthPrefix = append([]byte("0"), lengthPrefix...)

				return append(lengthPrefix, encodedValue...), nil
			}),
			Unpacker: field.UnpackerFunc(func(packedFieldValue []byte, spec *field.Spec) ([]byte, int, error) {
				maxEncodedValueLength := spec.Length

				encodedValueLength, prefBytes, err := spec.Pref.DecodeLength(maxEncodedValueLength, packedFieldValue[1:])
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode length: %w", err)
				}

				// for BCD encoding, the length of the packed data is twice the length of the encoded value
				valueLength := encodedValueLength

				// Decode the packed data length
				value, read, err := spec.Enc.Decode(packedFieldValue[prefBytes+1:], valueLength)
				if err != nil {
					return nil, 0, fmt.Errorf("failed to decode content: %w", err)
				}

				if spec.Pad != nil {
					value = spec.Pad.Unpad(value)
				}

				return value, read + prefBytes + 1, nil
			}),
		}),
	},
}
