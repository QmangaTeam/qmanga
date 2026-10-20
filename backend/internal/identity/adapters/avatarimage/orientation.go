package avatarimage

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
)

// Container readers only slice existing bounded input; lengths never drive
// allocations. TIFF orientation is shared across all accepted image formats.
func sourceOrientation(raw []byte, format string) uint16 {
	switch format {
	case "jpeg":
		return jpegOrientation(raw)
	case "png":
		return pngOrientation(raw)
	case "webp":
		// Browser WebP decoders currently preserve pixel orientation even with
		// EXIF. Match that preview; JPEG/PNG metadata is applied above.
		return 1
	}
	return 1
}
func pngOrientation(raw []byte) uint16 {
	if len(raw) < 8 || !bytes.Equal(raw[:8], []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		return 1
	}
	for offset := 8; offset <= len(raw)-12; {
		length := uint64(binary.BigEndian.Uint32(raw[offset : offset+4]))
		if length > uint64(len(raw)-offset-12) {
			return 1
		}
		end := offset + 8 + int(length)
		kind := string(raw[offset+4 : offset+8])
		if kind == "eXIf" {
			if crc32.ChecksumIEEE(raw[offset+4:end]) != binary.BigEndian.Uint32(raw[end:end+4]) {
				return 1
			}
			return tiffOrientation(raw[offset+8 : end])
		}
		if kind == "IEND" {
			return 1
		}
		offset = end + 4
	}
	return 1
}

// jpegOrientation reads only the bounded JPEG APP1/IFD0 orientation scalar.
// It never follows sub-IFDs or allocates from untrusted metadata counts/offsets.
// Invalid or absent metadata keeps the decoded pixel orientation.
func jpegOrientation(raw []byte) uint16 {
	if len(raw) < 2 || raw[0] != 0xff || raw[1] != 0xd8 {
		return 1
	}
	for offset := 2; offset < len(raw); {
		if raw[offset] != 0xff {
			return 1
		}
		for offset < len(raw) && raw[offset] == 0xff {
			offset++
		}
		if offset >= len(raw) {
			return 1
		}
		marker := raw[offset]
		offset++
		if marker == 0xda || marker == 0xd9 {
			return 1
		}
		if marker == 0x01 || (marker >= 0xd0 && marker <= 0xd7) {
			continue
		}
		if offset+2 > len(raw) {
			return 1
		}
		size := int(binary.BigEndian.Uint16(raw[offset : offset+2]))
		if size < 2 || size > len(raw)-offset {
			return 1
		}
		payload := raw[offset+2 : offset+size]
		if marker == 0xe1 && bytes.HasPrefix(payload, []byte("Exif\x00\x00")) {
			return tiffOrientation(payload[6:])
		}
		offset += size
	}
	return 1
}
func tiffOrientation(raw []byte) uint16 {
	if len(raw) < 8 {
		return 1
	}
	var order binary.ByteOrder
	switch string(raw[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 1
	}
	if order.Uint16(raw[2:4]) != 42 {
		return 1
	}
	offset := uint64(order.Uint32(raw[4:8]))
	if offset < 8 || offset+2 > uint64(len(raw)) {
		return 1
	}
	count := uint64(order.Uint16(raw[offset : offset+2]))
	offset += 2
	if count > (uint64(len(raw))-offset)/12 {
		return 1
	}
	for index := uint64(0); index < count; index++ {
		entry := raw[offset+index*12 : offset+(index+1)*12]
		if order.Uint16(entry[:2]) != 0x0112 {
			continue
		}
		if order.Uint16(entry[2:4]) != 3 || order.Uint32(entry[4:8]) != 1 {
			return 1
		}
		value := order.Uint16(entry[8:10])
		if value >= 1 && value <= 8 {
			return value
		}
		return 1
	}
	return 1
}

// A coordinate view applies EXIF before crop/scale, without allocating a second
// full-resolution image. The existing decoded-pixel and source-byte limits stay
// authoritative; only the final 256px avatar is allocated after decode.
type orientedImage struct {
	source      image.Image
	orientation uint16
}

func (img orientedImage) ColorModel() color.Model { return img.source.ColorModel() }
func (img orientedImage) Bounds() image.Rectangle {
	bounds := img.source.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if img.orientation >= 5 {
		w, h = h, w
	}
	return image.Rect(0, 0, w, h)
}
func (img orientedImage) At(x, y int) color.Color {
	bounds := img.source.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	switch img.orientation {
	case 2:
		x = w - 1 - x
	case 3:
		x, y = w-1-x, h-1-y
	case 4:
		y = h - 1 - y
	case 5:
		x, y = y, x
	case 6:
		x, y = y, h-1-x
	case 7:
		x, y = w-1-y, h-1-x
	case 8:
		x, y = w-1-y, x
	}
	return img.source.At(bounds.Min.X+x, bounds.Min.Y+y)
}
