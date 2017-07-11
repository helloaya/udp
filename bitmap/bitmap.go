package bitmap


import (
	"log"
)


type Bitmap struct {
	Bits []byte
	Start uint32
	End uint32
}


func MakeBitmap (start uint32, end uint32) *Bitmap {
	b := &Bitmap{}
	b.Bits = make ([]byte, (end - start + 8) /8)
	b.Start = start
	b.End = end
	return b
}

func (b *Bitmap) Update(bits []byte) {
	if len(b.Bits) != len(bits) {
		log.Printf ("Update bits failed, invalid len %d\n", len(bits))
		return
	}
	b.Bits = bits
}

func (b *Bitmap) Getbit(index uint32) bool {
	if index < b.Start || index > b.End {
		log.Printf ("Setbit failed, index %d out of range", index)
		return false
	}
	l:= (index - b.Start)
	i := l / 8
	o := l % 8
	if 1 == ((b.Bits[i] >> o) & 1) {
		return true
	}
	return false
}

