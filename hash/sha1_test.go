package hash

import "testing"

func TestShortChecksum(t *testing.T) {
	t.Parallel()

	sum := "1234567890abcdef"

	if got := ShortChecksum(sum, 8); got != "12345678" {
		t.Fatalf("ShortChecksum truncation = %q, want %q", got, "12345678")
	}

	if got := ShortChecksum(sum, 32); got != sum {
		t.Fatalf("ShortChecksum full = %q, want %q", got, sum)
	}

	if got := ShortChecksum(sum, 0); got != "" {
		t.Fatalf("ShortChecksum zero = %q, want empty", got)
	}
}
