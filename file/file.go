package file

import (
	"errors"
	"log"
)


var (
	SIZE_PIECE uint32 = 1400
	ERR_RES_OUT_OF_RANGE = errors.New("index out of range")
	ERR_RES_NOT_EXIST = errors.New("resource not exist")
)

type File struct {
	Id string
	nPieces uint32
	data []byte
}

type Reader interface {
	Read (id string) (*File, error)
}


func MakeFile(id string, data []byte) *File{
	f := &File{}
	f.Id = id
	f.nPieces = (uint32(len(data)) + SIZE_PIECE - 1)/SIZE_PIECE
	f.data = data
	return f
}

func (file *File) GetPiece(index uint32) ([]byte) {
	if index >= file.nPieces {
		log.Panic ("Invalid index=", index, "nPieces=", file.nPieces)
	}

	var s  = index * SIZE_PIECE
	var e  = (index + 1) * SIZE_PIECE
	var l  = uint32(len(file.data))
	if e > l {
		e = l
	}
	return file.data[s:e]
}

func (file *File) PiecesNum () uint32 {
	return file.nPieces
}

