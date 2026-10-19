package shortcut

// Carbon codes identify physical positions; Windows scan codes are mapped below.
var keys = []struct {
	ID     string
	Carbon int64
	VK     uint32
}{
	{"KeyA", 0, 0x41}, {"KeyS", 1, 0x53}, {"KeyD", 2, 0x44}, {"KeyF", 3, 0x46}, {"KeyH", 4, 0x48}, {"KeyG", 5, 0x47}, {"KeyZ", 6, 0x5a}, {"KeyX", 7, 0x58}, {"KeyC", 8, 0x43}, {"KeyV", 9, 0x56}, {"KeyB", 11, 0x42}, {"KeyQ", 12, 0x51}, {"KeyW", 13, 0x57}, {"KeyE", 14, 0x45}, {"KeyR", 15, 0x52}, {"KeyY", 16, 0x59}, {"KeyT", 17, 0x54},
	{"Digit1", 18, 0x31}, {"Digit2", 19, 0x32}, {"Digit3", 20, 0x33}, {"Digit4", 21, 0x34}, {"Digit6", 22, 0x36}, {"Digit5", 23, 0x35}, {"Equal", 24, 0xbb}, {"Digit9", 25, 0x39}, {"Digit7", 26, 0x37}, {"Minus", 27, 0xbd}, {"Digit8", 28, 0x38}, {"Digit0", 29, 0x30}, {"BracketRight", 30, 0xdd}, {"KeyO", 31, 0x4f}, {"KeyU", 32, 0x55}, {"BracketLeft", 33, 0xdb}, {"KeyI", 34, 0x49}, {"KeyP", 35, 0x50}, {"Enter", 36, 0x0d}, {"KeyL", 37, 0x4c}, {"KeyJ", 38, 0x4a}, {"Quote", 39, 0xde}, {"KeyK", 40, 0x4b}, {"Semicolon", 41, 0xba}, {"Backslash", 42, 0xdc}, {"Comma", 43, 0xbc}, {"Slash", 44, 0xbf}, {"KeyN", 45, 0x4e}, {"KeyM", 46, 0x4d}, {"Period", 47, 0xbe}, {"Tab", 48, 9}, {"Space", 49, 0x20}, {"Backquote", 50, 0xc0}, {"Backspace", 51, 8}, {"Escape", 53, 0x1b}, {"Fn", 63, 0},
	{"NumpadDecimal", 65, 0x6e}, {"NumpadMultiply", 67, 0x6a}, {"NumpadAdd", 69, 0x6b}, {"NumpadClear", 71, 0x0c}, {"NumpadDivide", 75, 0x6f}, {"NumpadEnter", 76, 0x10d}, {"NumpadSubtract", 78, 0x6d}, {"NumpadEqual", 81, 0}, {"Numpad0", 82, 0x60}, {"Numpad1", 83, 0x61}, {"Numpad2", 84, 0x62}, {"Numpad3", 85, 0x63}, {"Numpad4", 86, 0x64}, {"Numpad5", 87, 0x65}, {"Numpad6", 88, 0x66}, {"Numpad7", 89, 0x67}, {"Numpad8", 91, 0x68}, {"Numpad9", 92, 0x69},
	{"F1", 122, 0x70}, {"F2", 120, 0x71}, {"F3", 99, 0x72}, {"F4", 118, 0x73}, {"F5", 96, 0x74}, {"F6", 97, 0x75}, {"F7", 98, 0x76}, {"F8", 100, 0x77}, {"F9", 101, 0x78}, {"F10", 109, 0x79}, {"F11", 103, 0x7a}, {"F12", 111, 0x7b}, {"F13", 105, 0x7c}, {"F14", 107, 0x7d}, {"F15", 113, 0x7e}, {"F16", 106, 0x7f}, {"F17", 64, 0x80}, {"F18", 79, 0x81}, {"F19", 80, 0x82}, {"F20", 90, 0x83}, {"F21", 644, 0x84}, {"F22", 645, 0x85}, {"F23", 646, 0x86}, {"F24", 647, 0x87}, {"Insert", 557, 0x2d},
	{"Home", 115, 0x24}, {"PageUp", 116, 0x21}, {"Delete", 117, 0x2e}, {"End", 119, 0x23}, {"PageDown", 121, 0x22}, {"ArrowLeft", 123, 0x25}, {"ArrowRight", 124, 0x27}, {"ArrowDown", 125, 0x28}, {"ArrowUp", 126, 0x26},
}

func Key(code int64) string {
	if code == 179 {
		return "Fn"
	}
	for _, k := range keys {
		if k.Carbon == code {
			return k.ID
		}
	}
	return ""
}

// VK is the Windows virtual key of a key ID, 0 for one Windows lacks.
func VK(id string) uint32 {
	for _, k := range keys {
		if k.ID == id {
			return k.VK
		}
	}
	return 0
}
func WindowsKey(vk uint32) int64 {
	for _, k := range keys {
		if k.VK != 0 && k.VK == vk {
			return k.Carbon
		}
	}
	return int64(vk) + 512
}
func Supported(id, os string) bool {
	for _, k := range keys {
		if k.ID == id {
			if os == "windows" {
				return k.VK != 0
			}
			return k.Carbon < 256
		}
	}
	return false
}
func Modifier(code int64) (Mods, bool) {
	switch code {
	case 56:
		return LeftShift, true
	case 60:
		return RightShift, true
	case 59:
		return LeftCtrl, true
	case 62:
		return RightCtrl, true
	case 58:
		return LeftAlt, true
	case 61:
		return RightAlt, true
	case 55:
		return LeftMeta, true
	case 54:
		return RightMeta, true
	}
	return 0, false
}

// WindowsVK uses hardware scan positions for typing keys; NumLock must not
// turn a keypad key into the main navigation key in a saved binding.
func WindowsVK(vk, scan uint32, extended bool) uint32 {
	if vk == 0x0d && extended {
		return 0x10d
	}
	if !extended {
		if key, ok := scanVK[scan]; ok {
			return key
		}
		if key, ok := keypadVK[scan]; ok {
			return key
		}
	}
	return vk
}

var scanVK = map[uint32]uint32{
	0x02: '1', 0x03: '2', 0x04: '3', 0x05: '4', 0x06: '5', 0x07: '6', 0x08: '7', 0x09: '8', 0x0a: '9', 0x0b: '0', 0x0c: 0xbd, 0x0d: 0xbb,
	0x10: 'Q', 0x11: 'W', 0x12: 'E', 0x13: 'R', 0x14: 'T', 0x15: 'Y', 0x16: 'U', 0x17: 'I', 0x18: 'O', 0x19: 'P', 0x1a: 0xdb, 0x1b: 0xdd,
	0x1e: 'A', 0x1f: 'S', 0x20: 'D', 0x21: 'F', 0x22: 'G', 0x23: 'H', 0x24: 'J', 0x25: 'K', 0x26: 'L', 0x27: 0xba, 0x28: 0xde, 0x29: 0xc0, 0x2b: 0xdc,
	0x2c: 'Z', 0x2d: 'X', 0x2e: 'C', 0x2f: 'V', 0x30: 'B', 0x31: 'N', 0x32: 'M', 0x33: 0xbc, 0x34: 0xbe, 0x35: 0xbf,
}
var keypadVK = map[uint32]uint32{0x47: 0x67, 0x48: 0x68, 0x49: 0x69, 0x4b: 0x64, 0x4c: 0x65, 0x4d: 0x66, 0x4f: 0x61, 0x50: 0x62, 0x51: 0x63, 0x52: 0x60, 0x53: 0x6e}
