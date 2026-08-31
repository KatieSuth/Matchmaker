package apilink

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/KatieSuth/MatchmakerAPI/internal/cryptoutil"
)

// DefaultKeyID is stored on rows encrypted with API_LINK_ENCRYPTION_KEY when
// API_LINK_ENCRYPTION_KEY_ID is unset.
const DefaultKeyID = "1"

var keyIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// Keyring holds the current wrapping key (used for new writes) plus any previous
// keys still needed to open existing api_links rows.
type Keyring struct {
	currentID string
	keys      map[string][]byte
}

// NewKeyring builds a keyring. currentID identifies currentKey in api_links.key_id.
// previous maps older key IDs to 32-byte AES keys; it must not reuse currentID.
func NewKeyring(currentID string, currentKey []byte, previous map[string][]byte) (*Keyring, error) {
	if err := validateKeyID(currentID); err != nil {
		return nil, fmt.Errorf("current key id: %w", err)
	}
	if len(currentKey) != 32 {
		return nil, fmt.Errorf("current key must be 32 bytes")
	}

	keys := make(map[string][]byte, 1+len(previous))
	keys[currentID] = append([]byte(nil), currentKey...)

	for id, key := range previous {
		if err := validateKeyID(id); err != nil {
			return nil, fmt.Errorf("previous key id %q: %w", id, err)
		}
		if id == currentID {
			return nil, fmt.Errorf("previous key id %q duplicates the current key id", id)
		}
		if len(key) != 32 {
			return nil, fmt.Errorf("previous key %q must be 32 bytes", id)
		}
		if _, exists := keys[id]; exists {
			return nil, fmt.Errorf("duplicate key id %q", id)
		}
		keys[id] = append([]byte(nil), key...)
	}

	return &Keyring{currentID: currentID, keys: keys}, nil
}

// CurrentID is written to api_links.key_id on encrypt.
func (k *Keyring) CurrentID() string {
	return k.currentID
}

// currentKey returns the AES-256 key used for new Seal operations.
func (k *Keyring) currentKey() []byte {
	return k.keys[k.currentID]
}

// key returns the AES-256 key for id, or an error if that id is not in the ring.
func (k *Keyring) key(id string) ([]byte, error) {
	key, ok := k.keys[id]
	if !ok {
		return nil, fmt.Errorf("unknown encryption key id %q", id)
	}
	return key, nil
}

// ParsePreviousKeys parses API_LINK_ENCRYPTION_PREVIOUS_KEYS: comma-separated
// `id:hex` entries (64 hex chars each). Empty input yields an empty map.
func ParsePreviousKeys(s string) (map[string][]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return map[string][]byte{}, nil
	}

	keys := make(map[string][]byte)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, hexKey, ok := strings.Cut(part, ":")
		if !ok || id == "" || hexKey == "" {
			return nil, fmt.Errorf("previous key entry %q must be id:hex", part)
		}
		id = strings.TrimSpace(id)
		hexKey = strings.TrimSpace(hexKey)
		if err := validateKeyID(id); err != nil {
			return nil, fmt.Errorf("previous key id: %w", err)
		}
		if _, exists := keys[id]; exists {
			return nil, fmt.Errorf("duplicate previous key id %q", id)
		}
		key, err := cryptoutil.ParseAES256Key(hexKey)
		if err != nil {
			return nil, fmt.Errorf("previous key %q: %w", id, err)
		}
		keys[id] = key
	}
	return keys, nil
}

// validateKeyID reports whether id is a safe keyring identifier (no colon or comma,
// so PREVIOUS_KEYS parsing stays unambiguous).
func validateKeyID(id string) error {
	if !keyIDPattern.MatchString(id) {
		return fmt.Errorf("must match %s", keyIDPattern.String())
	}
	return nil
}
