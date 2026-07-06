package model

import "testing"

func TestBlockKindString(t *testing.T) {
	if BlockFunc.String() != "func" || BlockMethod.String() != "method" || BlockType.String() != "type" {
		t.Fatal("block kind strings")
	}
	if BlockKind(99).String() != "?" {
		t.Fatal("unknown kind")
	}
}
