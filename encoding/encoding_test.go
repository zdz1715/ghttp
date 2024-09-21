package encoding

import "testing"

func TestGetCodec(t *testing.T) {
	c := GetCodec("")
	t.Log(c)
}
