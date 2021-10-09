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
	DefaultHeadLength = 17
)

//NetioData is a communication head struct between beacon and teamserver
type NetioData struct {
	Size      uint32 //data size not contain NetioData size
	Encrypted bool   //encrypt flag
	SessionId uint64 //session id
	Reserved1 int32  //reserved, now this is a random int
	Data      []byte //data
}

type INetioData interface {
	getDataLen() uint32
	IsEncrypted() bool
	GetData() []byte
	GetSessionId() uint64

	setDataLen(uint32)
	setEncrypted(bool)
	setData([]byte)
	setSessionId(uint64)
}

func NewNetioData(encrypted bool, sessionid uint64, data []byte) *NetioData {
	return &NetioData{
		Size:      uint32(len(data)),
		Encrypted: encrypted,
		SessionId: sessionid,
		Data:      data,
	}
}

func (msg *NetioData) getDataLen() uint32 {
	return msg.Size
}

func (msg *NetioData) IsEncrypted() bool {
	return msg.Encrypted
}

func (msg *NetioData) GetSessionId() uint64 {
	return msg.SessionId
}

func (msg *NetioData) GetData() []byte {
	return msg.Data
}

func (msg *NetioData) setDataLen(len uint32) {
	msg.Size = len
}

func (msg *NetioData) setEncrypted(encrypted bool) {
	msg.Encrypted = encrypted
}

func (msg *NetioData) setSessionId(sessionid uint64) {
	msg.SessionId = sessionid
}

func (msg *NetioData) setData(data []byte) {
	msg.Data = data
}
