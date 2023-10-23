package torrent

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
)

const (
	PieceLength = 20
)

type Meta struct {
	Raw      interface{}
	InfoHash string
	Announce string `json:"announce"`
	Info     struct {
		Length      int    `json:"length"`
		Name        string `json:"name"`
		PieceLength int    `json:"piece length"`
		Pieces      string `json:"pieces"`
	} `json:"info"`
}

type MetaInfo struct {
	URL         string
	Length      int
	InfoHash    string
	PieceLength int
	PieceHashes []string
}

func Hash(blob string) string {
	sum := sha1.Sum([]byte(blob))
	return hex.EncodeToString(sum[:])
}

func NewMetaFromFile(metaFilePath string) (*Meta, error) {
	data, err := os.ReadFile(metaFilePath)
	if err != nil {
		return nil, err
	}

	return NewMetaFromBencode(string(data))
}

func NewMetaFromBencode(bencodeData string) (*Meta, error) {
	decoded, _, err := DecodeBencode(string(bencodeData))
	if err != nil {
		return nil, err
	}

	return NewMeta(decoded)
}

func NewMeta(benocodedData interface{}) (*Meta, error) {
	metaDict := benocodedData.(Dict)
	info := metaDict["info"].(Dict)
	pieces := info["pieces"].(string)

	meta := Meta{}
	marshaled, err := json.Marshal(benocodedData)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(marshaled, &meta)
	if err != nil {
		return nil, err
	}

	// @todo(kotsmile) fix problem with converting unicode charecters to plain text
	meta.Info.Pieces = pieces
	meta.Raw = benocodedData

	infoHash, err := meta.GetHashInfo()
	if err != nil {
		return nil, err
	}

	meta.InfoHash = infoHash

	return &meta, nil
}

func (meta *Meta) GetPieces() ([]string, error) {
	piecesArr := []byte(meta.Info.Pieces)
	pieces := make([]string, 0, 10)
	if len(piecesArr)%PieceLength != 0 {
		return nil, errors.New("length pieces is not divisible by 20")
	}

	for index := 0; index+PieceLength <= len(piecesArr); index += PieceLength {
		pieces = append(
			pieces,
			hex.EncodeToString(
				piecesArr[index:index+PieceLength],
			),
		)
	}

	return pieces, nil
}

func (meta *Meta) GetHashInfo() (string, error) {
	decodedMap, ok := meta.Raw.(Dict)
	if !ok {
		return "", errors.New("not a dict")
	}

	infoValue, ok := decodedMap["info"]
	if !ok {
		return "", errors.New("no 'info' key in dict")
	}

	encodedInfo, err := EncodeBencode(infoValue)
	if err != nil {
		return "", err
	}

	return Hash(encodedInfo), nil
}

func (meta *Meta) GetInfo() (*MetaInfo, error) {

	metaInfo := &MetaInfo{}

	metaInfo.InfoHash = meta.InfoHash
	metaInfo.URL = meta.Announce
	metaInfo.Length = meta.Info.Length
	metaInfo.PieceLength = meta.Info.PieceLength

	pieces, err := meta.GetPieces()
	if err != nil {
		return nil, err
	}

	metaInfo.PieceHashes = pieces

	return metaInfo, nil
}
