/*
 * Copyright (c) 2021.  https://github.com/geemion
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/panjf2000/gnet"
	"log"
	"sync"
	"teamserver/internal/handler"
	pb "teamserver/internal/proto/protobuf"
)

//ServerMgr is a struct to save servers info, it's a map
type ServerMgr struct {
	serverInfo sync.Map //key is a server name, value is server addr
}

//StartServer is to start a server
func (s *ServerMgr) StartServer(name string, addr string, msghandler *handler.MsgHandler) (err error) {
	if s.isRunningServer(name) {
		return errors.New("Duplicate listener name")
	}
	go func() {
		server := newBeaconServer(addr, false, false, nil, msghandler)
		err = server.Run()
		if err != nil {
			log.Print(err)
		}
		s.serverInfo.Delete(name)
	}()

	s.serverInfo.Store(name, addr)
	return nil
}

//StopServer is to stop a server
func (s *ServerMgr) StopServer(name string) (err error) {
	addr, ok := s.serverInfo.Load(name)
	if !ok {
		return errors.New("no server named:" + name + "is running")
	}

	return gnet.Stop(context.TODO(), addr.(string))
}

//GetRunningServer return all running server info
func (s *ServerMgr) GetRunningServer() (data []byte, err error) {
	info := &pb.ServerInfo{}

	s.serverInfo.Range(func(key, value interface{}) bool {
		name := key.(string)
		addr := value.(string)
		item := &pb.ServerItem{}
		item.Name = name
		item.Addr = addr
		info.Server = append(info.Server, item)
		return true
	})

	data, err = proto.Marshal(info)
	return
}

//isRunningServer return server run status by name
func (s *ServerMgr) isRunningServer(name string) bool {
	running := false
	s.serverInfo.Range(func(key, value interface{}) bool {
		tempName := key.(string)
		if tempName == name {
			running = true
			return false
		}
		return true
	})
	return running
}
