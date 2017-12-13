package gossiperServer

import (
	"github.com/pauarge/peerster/gossiper/common"
	"io/ioutil"
	"fmt"
	"strconv"
	"bytes"
	"crypto/sha256"
)

func (g *Gossiper) reconstructFile(filename string) bool {
	metafile, err := ioutil.ReadFile(common.TmpFilePath + filename + ".meta")
	if err != nil {
		fmt.Println("NO METAFILE FOUND")
		return false
	}

	g.downloadMetasLock.RLock()
	if ch, ok := g.downloadMetas[filename]; ok {
		ch <- true
	}
	g.downloadMetasLock.RUnlock()

	g.filesLock.Lock()
	tmp := g.files[filename]
	tmp.Metafile = metafile
	hash := sha256.New()
	hash.Write(metafile)
	tmp.Metahash = hash.Sum(nil)
	g.files[filename] = tmp
	g.filesLock.Unlock()

	hashes := split(metafile, common.HashLen)
	var file []byte
	for i := range hashes {
		f, err := ioutil.ReadFile(common.TmpFilePath + filename + ".part" + strconv.Itoa(i))
		if err != nil {
			fmt.Printf("CHUNK %d NOT FOUND\n", i)
			return false
		}
		file = append(file, f...)
	}

	err = ioutil.WriteFile(common.FilePath+filename, file, 0644)
	check(err)
	fmt.Printf("RECONSTRUCTED file %s\n", filename)
	return true
}

func (g *Gossiper) getChunkNum(filename string, hash []byte) int {
	g.filesLock.RLock()
	file := g.files[filename]
	g.filesLock.RUnlock()

	hashes := split(file.Metafile, common.HashLen)
	for i := range hashes {
		if bytes.Equal(hash, hashes[i]) {
			return i
		}
	}

	return -1
}

func (g *Gossiper) getNextChunk(filename string) int {
	g.filesLock.RLock()
	file, ok := g.files[filename]
	g.filesLock.RUnlock()

	if ok {
		hashes := split(file.Metafile, common.HashLen)
		m := g.getLocalFiles()

		for i := range hashes {
			if !m[filename+".part"+strconv.Itoa(i)] {
				return i
			}
		}

		return -1
	} else {
		return 0
	}
}

func (g *Gossiper) getChunkMap(filename string) []uint64 {
	g.filesLock.RLock()
	file, ok := g.files[filename]
	g.filesLock.RUnlock()

	var res []uint64

	if ok {
		hashes := split(file.Metafile, common.HashLen)
		m := g.getLocalFiles()

		for i := range hashes {
			if m[filename+".part"+strconv.Itoa(i)] {
				res = append(res, uint64(i))
			}
		}
	}

	return res
}

func (g *Gossiper) getLocalFiles() map[string]bool {
	files, err := ioutil.ReadDir(common.TmpFilePath)
	check(err)
	m := make(map[string]bool)
	for _, f := range files {
		if !f.IsDir() {
			m[f.Name()] = true
		}
	}
	return m
}
