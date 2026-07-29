package ptr_test

import (
	"testing"

	"github.com/ali-gulzar/speechory-core/pkg/ptr"
)

type address struct {
	city string
}

type user struct {
	address *address
}

func TestTo(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		v := 42
		p := ptr.To(v)
		if p == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *p != v {
			t.Fatalf("expected %d, got %d", v, *p)
		}
	})

	t.Run("string", func(t *testing.T) {
		v := "hello"
		p := ptr.To(v)
		if p == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *p != v {
			t.Fatalf("expected %q, got %q", v, *p)
		}
	})

	t.Run("bool", func(t *testing.T) {
		p := ptr.To(true)
		if p == nil {
			t.Fatal("expected non-nil pointer")
		}
		if !*p {
			t.Fatal("expected true")
		}
	})

	t.Run("struct", func(t *testing.T) {
		a := address{city: "NYC"}
		p := ptr.To(a)
		if p == nil {
			t.Fatal("expected non-nil pointer")
		}
		if p.city != a.city {
			t.Fatalf("expected %q, got %q", a.city, p.city)
		}
	})

	t.Run("returns distinct pointer each call", func(t *testing.T) {
		p1 := ptr.To(1)
		p2 := ptr.To(1)
		if p1 == p2 {
			t.Fatal("expected distinct pointers")
		}
	})
}

func TestToPtrOrNil(t *testing.T) {
	t.Run("non-zero string returns pointer", func(t *testing.T) {
		p := ptr.ToPtrOrNil("hello")
		if p == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *p != "hello" {
			t.Fatalf("expected hello, got %q", *p)
		}
	})

	t.Run("zero string returns nil", func(t *testing.T) {
		if ptr.ToPtrOrNil("") != nil {
			t.Fatal("expected nil for empty string")
		}
	})

	t.Run("non-zero int returns pointer", func(t *testing.T) {
		p := ptr.ToPtrOrNil(42)
		if p == nil {
			t.Fatal("expected non-nil pointer")
		}
		if *p != 42 {
			t.Fatalf("expected 42, got %d", *p)
		}
	})

	t.Run("zero int returns nil", func(t *testing.T) {
		if ptr.ToPtrOrNil(0) != nil {
			t.Fatal("expected nil for zero int")
		}
	})

	t.Run("non-zero struct returns pointer", func(t *testing.T) {
		a := address{city: "NYC"}
		p := ptr.ToPtrOrNil(a)
		if p == nil {
			t.Fatal("expected non-nil pointer")
		}
		if p.city != "NYC" {
			t.Fatalf("expected NYC, got %q", p.city)
		}
	})

	t.Run("zero struct returns nil", func(t *testing.T) {
		if ptr.ToPtrOrNil(address{}) != nil {
			t.Fatal("expected nil for zero struct")
		}
	})

	t.Run("struct field via From then ToPtrOrNil", func(t *testing.T) {
		u := user{address: &address{city: "NYC"}}
		city := ptr.ToPtrOrNil(ptr.From(u.address).city)
		if city == nil || *city != "NYC" {
			t.Fatalf("expected NYC, got %v", city)
		}
	})

	t.Run("nil struct pointer field returns nil", func(t *testing.T) {
		u := user{address: nil}
		city := ptr.ToPtrOrNil(ptr.From(u.address).city)
		if city != nil {
			t.Fatalf("expected nil for zero city, got %q", *city)
		}
	})
}

func TestFrom(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		v := 42
		if got := ptr.From(&v); got != v {
			t.Fatalf("expected %d, got %d", v, got)
		}
	})

	t.Run("string", func(t *testing.T) {
		v := "hello"
		if got := ptr.From(&v); got != v {
			t.Fatalf("expected %q, got %q", v, got)
		}
	})

	t.Run("struct", func(t *testing.T) {
		a := address{city: "NYC"}
		if got := ptr.From(&a); got.city != a.city {
			t.Fatalf("expected %q, got %q", a.city, got.city)
		}
	})

	t.Run("nil pointer returns zero value", func(t *testing.T) {
		var p *int
		if got := ptr.From(p); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
	})

	t.Run("nil struct pointer returns zero struct", func(t *testing.T) {
		u := user{address: nil}
		if got := ptr.From(u.address); got != (address{}) {
			t.Fatalf("expected zero address, got %+v", got)
		}
	})

	t.Run("struct field access via From then To", func(t *testing.T) {
		u := user{address: &address{city: "NYC"}}
		city := ptr.To(ptr.From(u.address).city)
		if *city != "NYC" {
			t.Fatalf("expected NYC, got %q", *city)
		}
	})

	t.Run("nil struct pointer field access returns zero", func(t *testing.T) {
		u := user{address: nil}
		city := ptr.To(ptr.From(u.address).city)
		if *city != "" {
			t.Fatalf("expected empty string, got %q", *city)
		}
	})

	t.Run("roundtrip with To", func(t *testing.T) {
		v := 99
		if got := ptr.From(ptr.To(v)); got != v {
			t.Fatalf("expected %d, got %d", v, got)
		}
	})
}
