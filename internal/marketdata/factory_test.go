package marketdata

import "testing"

func TestNewProvider_FakeByDefault(t *testing.T) {
	if _, ok := NewProvider("fake", "").(*FakeProvider); !ok {
		t.Error("default/fake should return a FakeProvider")
	}
	if _, ok := NewProvider("unknown", "key").(*FakeProvider); !ok {
		t.Error("unknown provider should fall back to FakeProvider")
	}
}

func TestNewProvider_TwelveDataNeedsKey(t *testing.T) {
	// No key -> stay on fake so the app always boots.
	if _, ok := NewProvider("twelvedata", "").(*FakeProvider); !ok {
		t.Error("twelvedata without a key should fall back to FakeProvider")
	}
	// With a key -> the caching-wrapped real provider.
	if _, ok := NewProvider("twelvedata", "a-key").(*CachingProvider); !ok {
		t.Error("twelvedata with a key should return a CachingProvider")
	}
}
