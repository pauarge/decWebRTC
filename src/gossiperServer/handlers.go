package gossiperServer

import (
	"github.com/pauarge/peerster/gossiper/common"
	"fmt"
	"net"
	"strconv"
	"os"
	"crypto/sha256"
	"io"
	"io/ioutil"
	"bytes"
	"strings"
	"path/filepath"
	"log"
	"time"
)

func (g *Gossiper) HandlePeerMessage(msg common.PeerMessage) {
	fmt.Println("CLIENT " + msg.Text + " " + g.Name)

	msg.Id = g.counter
	rmsg := common.RumorMessage{Origin: g.Name, Id: g.counter, Text: msg.Text}

	g.wantLock.Lock()
	g.want[g.Name] = g.counter + 1
	g.wantLock.Unlock()

	g.MessagesLock.Lock()
	g.Messages[common.MapKey{Origin: g.Name, MessageId: g.counter}] = rmsg
	g.MessagesLock.Unlock()

	g.MsgOrderLock.Lock()
	g.MsgOrder = append(g.MsgOrder, common.MapKey{Origin: g.Name, MessageId: g.counter})
	g.MsgOrderLock.Unlock()

	g.counter += 1
	g.iterativeRumorMongering("", rmsg)
}

func (g *Gossiper) handleStatusPacket(msg common.StatusPacket, relay *net.UDPAddr) {
	relayStr := getRelayStr(relay)

	g.channelsLock.RLock()
	if ch, ok := g.channels[relayStr]; ok {
		ch <- true
	}
	g.channelsLock.RUnlock()

	remoteWant := parseWant(msg)
	g.wantLock.RLock()
	localWant := g.want
	synced := true
	var dest common.MapKey

	statusStr := "STATUS from " + relayStr + " |"
	for i := range remoteWant {
		statusStr += " origin " + i + " nextID " + strconv.Itoa(int(remoteWant[i]))
		if remoteWant[i] > localWant[i] {
			synced = false
		} else if remoteWant[i] < localWant[i] {
			dest = common.MapKey{Origin: i, MessageId: remoteWant[i]}
			break
		}
	}

	for i := range localWant {
		if remoteWant[i] > localWant[i] {
			synced = false
		} else if remoteWant[i] < localWant[i] {
			if remoteWant[i] == 0 {
				remoteWant[i] = 1
			}
			dest = common.MapKey{Origin: i, MessageId: remoteWant[i]}
			break
		}
	}
	g.wantLock.RUnlock()

	fmt.Println(statusStr)
	if dest != (common.MapKey{}) {
		g.MessagesLock.RLock()
		msg := g.Messages[dest]
		g.MessagesLock.RUnlock()
		g.rumorMongering(relayStr, msg)
	} else if synced {
		fmt.Println("IN SYNC WITH " + relayStr)
	} else {
		g.sendStatusPacket(relay)
	}
}

func (g *Gossiper) handleRumorMessage(msg common.RumorMessage, relay *net.UDPAddr) {
	g.wantLock.RLock()
	wantMsgOrigin := g.want[msg.Origin]
	g.wantLock.RUnlock()

	if wantMsgOrigin > msg.Id && msg.Origin != g.Name && msg.LastPort == nil && msg.LastIP == nil {
		g.NextHopLock.Lock()
		g.NextHop[msg.Origin] = relay
		g.NextHopLock.Unlock()
	} else if wantMsgOrigin == msg.Id || wantMsgOrigin == 0 {
		if msg.LastPort != nil && msg.LastIP != nil {
			g.PeersLock.Lock()
			g.Peers[msg.LastIP.String()+":"+strconv.Itoa(*msg.LastPort)] = true
			g.PeersLock.Unlock()
		} else {
			fmt.Println("DIRECT-ROUTE FOR " + msg.Origin + ": " + getRelayStr(relay))
		}

		var route bool
		if msg.Text != "" {
			fmt.Println("RUMOR origin " + msg.Origin + " from " + getRelayStr(relay) + " ID " +
				strconv.Itoa(int(msg.Id)) + " contents " + msg.Text)
			g.MsgOrderLock.Lock()
			g.MsgOrder = append(g.MsgOrder, common.MapKey{Origin: msg.Origin, MessageId: msg.Id})
			g.MsgOrderLock.Unlock()
			route = false
		} else {
			fmt.Println("ROUTE MESSAGE from " + getRelayStr(relay))
			route = true
		}

		g.wantLock.Lock()
		g.want[msg.Origin] = msg.Id + 1
		g.wantLock.Unlock()

		g.MessagesLock.Lock()
		g.Messages[common.MapKey{Origin: msg.Origin, MessageId: msg.Id}] = msg
		g.MessagesLock.Unlock()

		g.NextHopLock.Lock()
		g.NextHop[msg.Origin] = relay
		g.NextHopLock.Unlock()

		g.sendStatusPacket(relay)

		if route || !g.noforward {
			msg.LastIP = &relay.IP
			msg.LastPort = &relay.Port
			go g.iterativeRumorMongering(getRelayStr(relay), msg)
		} else {
			fmt.Println("Not forwarding private message")
		}
	}
}

func (g *Gossiper) HandlePrivateMessageClient(msg common.PrivateMessage) {
	msg.Origin = g.Name
	msg.Id = 0
	msg.HopLimit = common.MaxHops
	g.PrivateMessagesLock.Lock()
	g.PrivateMessages[msg.Destination] = append(g.PrivateMessages[msg.Destination], msg)
	g.PrivateMessagesLock.Unlock()
	p := common.GossipPacket{Private: &msg}
	g.sendPrivateMessage(msg.Destination, p)
}

func (g *Gossiper) handlePrivateMessage(msg common.PrivateMessage) {
	msg.HopLimit -= 1
	if msg.Destination == g.Name {
		g.PrivateMessagesLock.Lock()
		g.PrivateMessages[msg.Origin] = append(g.PrivateMessages[msg.Origin], msg)
		g.PrivateMessagesLock.Unlock()
		fmt.Println("PRIVATE: " + msg.Origin + ":" + strconv.Itoa(int(msg.HopLimit)) + ":" + msg.Text)
	} else if msg.HopLimit >= 0 && !g.noforward {
		p := common.GossipPacket{Private: &msg}
		g.sendPrivateMessage(msg.Destination, p)
	}
}

func (g *Gossiper) HandleDataRequestClient(msg common.DataRequest) {
	if msg.Destination != "" {
		fmt.Printf("DOWNLOADING metafile of %s from %s\n", msg.FileName, msg.Destination)
		msg.Origin = g.Name
		msg.HopLimit = common.MaxHops
		g.filesLock.Lock()
		g.files[msg.FileName] = common.StoredFile{Metahash: msg.HashValue}
		g.filesLock.Unlock()
		p := common.GossipPacket{DataRequest: &msg}
		g.sendDataRequest(msg.Destination, p, common.DataRequestRetries)
	} else {
		fmt.Println("UPLOADING FILE TO SYSTEM")
		g.HandleFileUpload(msg.FileName)
	}
}

func (g *Gossiper) HandleFileUpload(path string) {
	file, err := os.Open(common.FilePath + path)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	chunkCount := (fileInfo.Size() / common.ChunkSize) + 1

	var metafile []byte
	hash := sha256.New()
	buff := make([]byte, common.ChunkSize)
	for i := int64(0); i < chunkCount; i++ {
		hash.Reset()
		var chunkReaded int64
		for chunkReaded = 0; chunkReaded < common.ChunkSize; {
			m := int64(common.ChunkSize)
			if chunkReaded+m > common.ChunkSize {
				m = common.ChunkSize - chunkReaded
			}

			n, err := file.Read(buff[0:m])
			if err != nil {
				if err == io.EOF {
					break
				}

				fmt.Println("Cannot read file, ", path, ", error:", err.Error())
				os.Exit(1)
			}
			hash.Write(buff[0:n])
			chunkReaded += int64(n)
		}

		err = ioutil.WriteFile(common.TmpFilePath+path+".part"+strconv.Itoa(int(i)), buff[0:chunkReaded], 0644)
		check(err)

		h := hash.Sum(nil)
		metafile = append(metafile, h...)
	}

	err = ioutil.WriteFile(common.TmpFilePath+path+".meta", metafile, 0644)
	check(err)

	metahash := sha256.New()
	metahash.Write(metafile)

	g.filesLock.Lock()
	g.files[path] = common.StoredFile{Size: fileInfo.Size(), Metafile: metafile, Metahash: metahash.Sum(nil)}
	g.filesLock.Unlock()
}

func (g *Gossiper) handleDataRequest(msg common.DataRequest, relay *net.UDPAddr) {
	msg.HopLimit -= 1
	if msg.Destination == g.Name {
		g.filesLock.RLock()
		f, ok := g.files[msg.FileName]
		g.filesLock.RUnlock()

		if ok {
			if bytes.Equal(msg.HashValue, f.Metahash) {
				res := common.DataReply{
					Origin:      g.Name,
					Destination: msg.Origin,
					HopLimit:    common.MaxHops,
					FileName:    msg.FileName + ".meta",
					HashValue:   f.Metahash,
					Data:        f.Metafile}
				p := common.GossipPacket{DataReply: &res}
				g.sendPrivateMessage(msg.Origin, p)
			} else {
				i := g.getChunkNum(msg.FileName, msg.HashValue)
				if i >= 0 {
					fn := msg.FileName + ".part" + strconv.Itoa(i)
					data, err := ioutil.ReadFile(common.TmpFilePath + fn)
					if err != nil {
						fmt.Println("DID NOT FOUND THAT CHUNK")
						return
					}
					res := common.DataReply{
						Origin:      g.Name,
						Destination: msg.Origin,
						HopLimit:    common.MaxHops,
						FileName:    fn,
						HashValue:   msg.HashValue,
						Data:        data}
					p := common.GossipPacket{DataReply: &res}
					g.sendPrivateMessage(msg.Origin, p)
				}
			}
		}
	} else if msg.HopLimit > 0 && !g.noforward {
		p := common.GossipPacket{DataRequest: &msg}
		g.sendPrivateMessage(msg.Destination, p)
	}
}

func (g *Gossiper) handleDataReply(msg common.DataReply, relay *net.UDPAddr) {
	msg.HopLimit -= 1
	if msg.Destination == g.Name {
		hash := sha256.New()
		hash.Write(msg.Data)
		if bytes.Equal(hash.Sum(nil), msg.HashValue) {
			g.dataRequestsLock.RLock()
			if ch, ok := g.dataRequests[getRelayStr(relay)]; ok {
				ch <- true
			}
			g.dataRequestsLock.RUnlock()

			err := ioutil.WriteFile(common.TmpFilePath+msg.FileName, msg.Data, 0644)
			check(err)
			fn := strings.TrimSuffix(msg.FileName, filepath.Ext(msg.FileName))
			if !g.reconstructFile(fn) {
				i := g.getNextChunk(fn)
				fmt.Printf("DOWNLOADING %s chunk %d from %s\n", fn, i, msg.Origin)
				g.filesLock.Lock()
				res := common.DataRequest{
					Origin:      g.Name,
					Destination: msg.Origin,
					HopLimit:    common.MaxHops,
					FileName:    fn,
					HashValue:   split(g.files[fn].Metafile, 32)[i],
				}
				g.filesLock.Unlock()
				p := common.GossipPacket{DataRequest: &res}
				g.sendDataRequest(msg.Origin, p, common.DataRequestRetries)
			}
		}
	} else if msg.HopLimit > 0 && !g.noforward {
		p := common.GossipPacket{DataReply: &msg}
		g.sendPrivateMessage(msg.Destination, p)
	}
}

func (g *Gossiper) HandleKeywords(keywords, budget string) {
	b, err := strconv.Atoi(budget)
	if err != nil {
		b = common.DefaultBudget
	}

	kList := strings.Split(keywords, ",")
	msg := common.SearchRequest{
		Origin:   g.Name,
		Budget:   uint64(b) + 1,
		Keywords: kList,
	}
	g.handleSearchRequest(msg, nil)
}

func (g *Gossiper) handleSearchRequest(msg common.SearchRequest, relay *net.UDPAddr) {
	if msg.Origin == "" {
		// request comes from CLI
		msg.Origin = g.Name
	}

	g.searchRequestsLock.RLock()
	_, ok := g.searchRequests[common.MapKeySearchRequest{Origin: msg.Origin, Keywords: strings.Join(msg.Keywords, "")}]
	g.searchRequestsLock.RUnlock()

	if !ok {
		g.searchRequestsLock.Lock()
		g.searchRequests[common.MapKeySearchRequest{Origin: msg.Origin, Keywords: strings.Join(msg.Keywords, "")}] = true
		g.searchRequestsLock.Unlock()

		g.filesLock.RLock()
		files := g.files
		g.filesLock.RUnlock()

		if msg.Origin != g.Name {
			res := common.SearchReply{
				Origin:      g.Name,
				Destination: msg.Origin,
				HopLimit:    common.MaxHops,
			}

			for i := range msg.Keywords {
				for k, v := range files {
					if strings.Contains(k, msg.Keywords[i]) {
						cm := g.getChunkMap(k)
						sr := common.SearchResult{
							FileName:     k,
							MetafileHash: v.Metahash,
							ChunkMap:     cm,
						}
						res.Results = append(res.Results, &sr)
						fmt.Printf("FOUND match %s at %s budget=%d metafile=%x chunks=%v\n",
							k, g.Name, msg.Budget, v.Metahash, cm)
					}
				}
			}

			p := common.GossipPacket{SearchReply: &res}
			g.sendPrivateMessage(msg.Origin, p)
		}

		msg.Budget -= 1
		if msg.Budget > 0 {
			g.NextHopLock.RLock()
			peers := g.NextHop
			g.NextHopLock.RUnlock()

			if uint64(len(peers)) <= msg.Budget {
				msg.Budget = 1
				p := common.GossipPacket{SearchRequest: &msg}
				i := 0
				for k := range peers {
					if uint64(i) >= msg.Budget {
						break
					}
					g.sendPrivateMessage(k, p)
					i++
				}
			} else {
				m := msg.Budget % uint64(len(peers))
				msg.Budget = msg.Budget / uint64(len(peers))
				i := 0
				for k := range peers {
					if uint64(i) < m {
						msg.Budget += 1
					}
					p := common.GossipPacket{SearchRequest: &msg}
					g.sendPrivateMessage(k, p)
					i++
				}
			}
		}

		ticker := time.NewTicker(time.Millisecond * common.DuplicSearchRequestTime)
		go func() {
			for range ticker.C {
				g.searchRequestsLock.Lock()
				delete(g.searchRequests, common.MapKeySearchRequest{Origin: msg.Origin, Keywords: strings.Join(msg.Keywords, "")})
				g.searchRequestsLock.Unlock()
				return
			}
		}()
	}
}

func (g *Gossiper) handleSearchReply(msg common.SearchReply, relay *net.UDPAddr) {
	msg.HopLimit -= 1
	if msg.Destination == g.Name {
		for i := range msg.Results {
			fn := common.TmpFilePath + msg.Results[i].FileName + ".meta"
			_, err := ioutil.ReadFile(fn)
			if err != nil {
				fmt.Printf("DOWNLOADING metafile of %s from %s\n", msg.Results[i].FileName, msg.Origin)
				res := common.DataRequest{
					Origin:      g.Name,
					Destination: msg.Origin,
					HopLimit:    common.MaxHops,
					FileName:    msg.Results[i].FileName,
					HashValue:   msg.Results[i].MetafileHash,
				}
				p := common.GossipPacket{DataRequest: &res}
				g.sendDataRequest(msg.Origin, p, common.DataRequestRetries)

				g.downloadMetasLock.Lock()
				g.downloadMetas[msg.Results[i].FileName] = make(chan bool)
				ch := g.downloadMetas[msg.Results[i].FileName]
				g.downloadMetasLock.Unlock()
				_ = <-ch
			}

			for j := range msg.Results[i].ChunkMap {
				fn := common.TmpFilePath + msg.Results[i].FileName + ".part" + strconv.Itoa(int(msg.Results[i].ChunkMap[j]))
				_, err := ioutil.ReadFile(fn)
				if err != nil {
					g.filesLock.Lock()
					res := common.DataRequest{
						Origin:      g.Name,
						Destination: msg.Origin,
						HopLimit:    common.MaxHops,
						FileName:    msg.Results[i].FileName,
						HashValue:   split(g.files[msg.Results[i].FileName].Metafile, 32)[i],
					}
					g.filesLock.Unlock()
					fmt.Printf("DOWNLOADING %s chunk %d from %s\n", fn, i, msg.Origin)
					p := common.GossipPacket{DataRequest: &res}
					g.sendDataRequest(msg.Origin, p, common.DataRequestRetries)
				}
			}

			g.reconstructFile(msg.Results[i].FileName)
		}
	} else if msg.HopLimit > 0 && !g.noforward {
		p := common.GossipPacket{SearchReply: &msg}
		g.sendPrivateMessage(msg.Destination, p)
	}
}
