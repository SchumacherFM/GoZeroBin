package gzbLibs

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

const (
	FILE_EXT = ".json"
)

type ZeroBin struct {
	Expiration       int64
	BurnAfterReading int
	PasteId          string
	PasteData        []byte
	DataDir          string
}

type ZeroContainer struct {
	Data []byte
	Meta map[string]int64
}

func (p *ZeroBin) Save() error {

	if "" == p.PasteId {
		h := sha1.New()
		h.Write(p.PasteData)
		hashByte := []byte(hex.EncodeToString(h.Sum(nil)))
		p.PasteId = string(hashByte[0:20]) // substring to 20 chars
	}

	ut := time.Now().Unix()
	meta := map[string]int64{
		"expire_date":        p.Expiration + ut,
		"postdate":           ut,
		"burn_after_reading": int64(p.BurnAfterReading),
	}

	jsonData, _ := json.Marshal(ZeroContainer{
		Meta: meta,
		Data: p.PasteData,
	})
	filename := p.DataDir + p.PasteId + FILE_EXT
	return ioutil.WriteFile(filename, jsonData, 0600)
}

func (p *ZeroBin) Delete() error {
	filename := p.DataDir + p.PasteId + FILE_EXT
	p.Expiration = 0
	p.PasteId = ""
	p.PasteData = []byte("")
	return os.Remove(filename)
}

func LoadZeroBin(dataDir, pasteId string) (*ZeroBin, error) {
	if "" != pasteId {
		filename := dataDir + pasteId + FILE_EXT
		body, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		container := ZeroContainer{}
		json.Unmarshal(body, &container)

		zeroBin := container.Data

		return &ZeroBin{
			Expiration:       container.Meta["expire_date"],
			BurnAfterReading: int(container.Meta["burn_after_reading"]),
			PasteId:          pasteId,
			PasteData:        zeroBin,
			DataDir:          dataDir,
		}, nil
	}
	return &ZeroBin{DataDir: dataDir}, nil
}
