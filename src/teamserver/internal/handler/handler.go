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

package handler

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/panjf2000/gnet"
	"log"
	"math/rand"
	"sync"
	bn "teamserver/internal/beacon"
	"teamserver/internal/conf"
	"teamserver/internal/proto/encode"
	pb "teamserver/internal/proto/protobuf"
	"teamserver/internal/store"
	"teamserver/pkg/crypto"
	"teamserver/pkg/mq"
	"time"
)

//MsgHandler represents all beacons msg handler functions
type MsgHandler struct {
	Session  sync.Map
	cmdQueue *mq.Client
}

//NewMsgHandler return a msghandler
func NewMsgHandler(mqclient *mq.Client) *MsgHandler {
	handler := &MsgHandler{cmdQueue: mqclient}
	go handler.pushTask()
	return handler
}

//HandleMsg handler beacon msg,return send data
func (hm *MsgHandler) HandleMsg(msg []byte, c gnet.Conn, conntype pb.CONN_TYPE) (rsp []byte, err error) {

	dp := encode.NewDataPack()
	netio, err := dp.Unpack(msg)
	if err != nil {
		return nil, err
	}

	if !netio.IsEncrypted() {
		return hm.auth(netio, c, conntype)
	}

	return hm.msgdispatch(netio, conntype)
}

//HandleClose handler disconnect
func (hm *MsgHandler) HandleClose(c gnet.Conn) {
	sessionID := c.Context().(uint64)
	hm.Session.Delete(sessionID)
}

func (hm *MsgHandler) auth(netio encode.INetIOData, c gnet.Conn, connType pb.CONN_TYPE) (rsp []byte, err error) {
	task := &pb.TaskData{}
	err = proto.Unmarshal(netio.GetData(), task)
	if err != nil {
		return
	}
	switch pb.MSGID(task.MsgId) {
	case pb.MSGID_PUBKEY_REQ:
		rsp, err = hm.onReqPubKey(task, c, connType)
		return
	case pb.MSGID_AUTH_REQ:
		rsp, err = hm.onReqAuth(netio.GetSessionId(), task, c, connType)
		return
	default:
		return rsp, errors.New("no msgid")
	}
}

func (hm *MsgHandler) msgdispatch(netio encode.INetIOData, connType pb.CONN_TYPE) (rsp []byte, err error) {
	sessionID := netio.GetSessionId()
	beacon, ok := hm.Session.Load(sessionID)
	if !ok {
		return rsp, errors.New("no session id")
	}
	sessionKey := beacon.(*bn.Beacon).SessionKey
	taskData, err := crypto.Xchacha20(sessionKey, netio.GetData())
	if err != nil {
		return rsp, errors.New("error session key")
	}

	task := &pb.TaskData{}
	err = proto.Unmarshal(taskData, task)
	if err != nil {
		return
	}
	var taskRspData []byte
	switch pb.MSGID(task.MsgId) {
	case pb.MSGID_HOST_INFO_RSP:
		taskRspData, err = hm.onRspData(task)
		break
	case pb.MSGID_HEAT_BEAT_REQ:
		{
			taskRspData, err = hm.onQuerytask(task)

			ipaddr := beacon.(*bn.Beacon).Conn.RemoteAddr()
			hm.onUpdateBeacon(task, ipaddr.String())
			break
		}
	default:
		taskRspData, err = hm.onRspData(task)
		break
	}

	encTaskData, err := crypto.Xchacha20(sessionKey, taskRspData)
	if err != nil {
		return
	}

	dp := encode.NewDataPack()
	msg := encode.NewNetioData(true, sessionID, encTaskData)
	rsp, _ = dp.Pack(msg, connType)
	return
}

func (hm *MsgHandler) onReqPubKey(taskReq *pb.TaskData, c gnet.Conn, connType pb.CONN_TYPE) (rsp []byte, err error) {

	authKey := &pb.AuthRsaKey{Pe: conf.GlobalConf.RsaEncode.E, Pn: conf.GlobalConf.RsaEncode.N}
	authData, _ := proto.Marshal(authKey)

	taskRsp := &pb.TaskData{MsgId: int32(pb.MSGID_PUBKEY_RSP), BeaconId: taskReq.BeaconId, TaskId: 123445, ByteValue: authData}
	taskData, _ := proto.Marshal(taskRsp)

	nano := time.Now().UnixNano()
	rand.Seed(nano)
	sessionID := rand.Uint64()

	c.SetContext(sessionID)

	dp := encode.NewDataPack()
	msg := encode.NewNetioData(false, sessionID, taskData)
	rsp, _ = dp.Pack(msg, connType)

	return rsp, nil
}

func (hm *MsgHandler) onReqAuth(sessionID uint64, taskReq *pb.TaskData, c gnet.Conn, connType pb.CONN_TYPE) (rsp []byte, err error) {
	encSessionKey := taskReq.GetByteValue()
	key, err := conf.GlobalConf.RsaEncode.PrivateDecode(encSessionKey)
	if err != nil {
		return
	}
	beacon := &bn.Beacon{ID: taskReq.BeaconId, SessionKey: key, ConnType: connType, Conn: c}
	hm.Session.Store(sessionID, beacon)

	taskRsp := &pb.TaskData{MsgId: int32(pb.MSGID_AUTH_RSP), BeaconId: taskReq.BeaconId, TaskId: 123445, ByteValue: nil}
	taskData, _ := proto.Marshal(taskRsp)

	dp := encode.NewDataPack()
	msg := encode.NewNetioData(false, sessionID, taskData)
	rsp, _ = dp.Pack(msg, connType)
	return
}

//todo: save rsp to db
func (hm *MsgHandler) onRspData(taskRsp *pb.TaskData) (rsp []byte, err error) {
	cmdRsp := pb.CommandRsp{
		TaskId:    taskRsp.TaskId,
		BeaconId:  taskRsp.BeaconId,
		MsgId:     taskRsp.MsgId,
		ByteValue: taskRsp.ByteValue,
	}
	hm.cmdQueue.Publish(conf.CmdRspTopic, cmdRsp)
	return
}

//todo:tcp push cmd
func (hm *MsgHandler) onQuerytask(taskRsp *pb.TaskData) (rsp []byte, err error) {
	taskData, err := store.GetTask(taskRsp.BeaconId)
	if err != nil {
		return
	}
	rsp, err = proto.Marshal(&taskData)
	return
}

func (hm *MsgHandler) onUpdateBeacon(taskRsp *pb.TaskData, ipAddr string) {
	store.UpdateBeacon(taskRsp.BeaconId, ipAddr, taskRsp.ByteValue)
}

func (hm *MsgHandler) pushTask() {

	taskCh, err := hm.cmdQueue.Subscribe(conf.CmdReqTopic)
	if err != nil {
		log.Print(err)
		return
	}

	defer hm.cmdQueue.Unsubscribe(conf.CmdReqTopic, taskCh)

	for {
		req, ok := (hm.cmdQueue.GetPayLoad(taskCh)).(*pb.CommandReq)
		if !ok {
			continue
		}

		var beacon *bn.Beacon = nil
		var sessionID uint64
		hm.Session.Range(func(key, value interface{}) bool {
			temp := value.(*bn.Beacon)
			if temp.ID == req.GetBeaconId() {
				sessionID = key.(uint64)
				beacon = temp
				ok = true
				return true
			}
			return false
		})

		if beacon != nil {
			task, err := store.GetTask(req.BeaconId)
			if err != nil {
				return
			}
			data, err := proto.Marshal(&task)
			encTaskData, err := crypto.Xchacha20(beacon.SessionKey, data)
			if err != nil {
				return
			}

			dp := encode.NewDataPack()
			msg := encode.NewNetioData(true, sessionID, encTaskData)
			msgRsp, _ := dp.Pack(msg, beacon.ConnType)

			if beacon.ConnType == pb.CONN_TYPE_CONNNAME_TCP {
				beacon.Conn.AsyncWrite(msgRsp)
			} else {
				beacon.Conn.SendTo(msgRsp)
			}
		}
	}
}
