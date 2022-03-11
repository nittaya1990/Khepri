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

package encode

const (
	//DefaultHeadLength default header size
	DefaultHeadLength = 17
)

//NetIOData is a communication head struct between beacon and teamserver
type NetIOData struct {
	Size      uint32 //data size not contain NetIOData size
	Encrypted bool   //encrypt flag
	SessionID uint64 //session id
	Reserved1 int32  //reserved, now this is a random int
	Data      []byte //data
}

//INetIOData netio data interface
type INetIOData interface {
	getDataLen() uint32
	IsEncrypted() bool
	GetData() []byte
	GetSessionId() uint64

	setDataLen(uint32)
	setEncrypted(bool)
	setData([]byte)
	setSessionId(uint64)
}

//NewNetioData return a NetIOData object
func NewNetioData(encrypted bool, sessionID uint64, data []byte) *NetIOData {
	return &NetIOData{
		Size:      uint32(len(data)),
		Encrypted: encrypted,
		SessionID: sessionID,
		Data:      data,
	}
}

func (msg *NetIOData) getDataLen() uint32 {
	return msg.Size
}

//IsEncrypted return msg is encrypt
func (msg *NetIOData) IsEncrypted() bool {
	return msg.Encrypted
}

//GetSessionId return msg session id
func (msg *NetIOData) GetSessionId() uint64 {
	return msg.SessionID
}

//GetData return msg data
func (msg *NetIOData) GetData() []byte {
	return msg.Data
}

func (msg *NetIOData) setDataLen(len uint32) {
	msg.Size = len
}

func (msg *NetIOData) setEncrypted(encrypted bool) {
	msg.Encrypted = encrypted
}

func (msg *NetIOData) setSessionId(sessionid uint64) {
	msg.SessionID = sessionid
}

func (msg *NetIOData) setData(data []byte) {
	msg.Data = data
}
