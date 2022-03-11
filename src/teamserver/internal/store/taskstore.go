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

package store

import (
	"errors"
	pb "teamserver/internal/proto/protobuf"
	"time"
)

type taskStatus int32

const (
	statusCreate   taskStatus = 0
	statusDispatch taskStatus = 1
	statusDone     taskStatus = 2
)

//TaskStore represents beacon task in database
type TaskStore struct {
	CreatedAt time.Time  //task create time
	UpdatedAt time.Time  //task update time
	TaskID    uint64     `gorm:"AUTO_INCREMENT;primary_key"` //taskid
	MsgId     int32      //msgid
	BeaconId  string     //beaconid
	ReqParam  []byte     //request param data
	RspParam  []byte     //resp  data
	Status    taskStatus //task status
}

//AddTask add a task to database
func AddTask(msgID int32, beaconID string, reqParam []byte) (err error) {
	task := TaskStore{
		MsgId:    msgID,
		BeaconId: beaconID,
		ReqParam: reqParam,
		Status:   statusCreate,
	}
	db := instance()

	db.AutoMigrate(&TaskStore{})

	if err := db.Create(&task).Error; err != nil {
		return err
	}
	return
}

//GetTask return a task from database
func GetTask(beaconID string) (data pb.TaskData, err error) {
	task := TaskStore{}
	db := instance()

	if db.Where("beacon_id = ? and status = ?", beaconID, statusCreate).First(&task).RecordNotFound() {
		err = errors.New("no task")
		return
	}

	data.MsgId = task.MsgId
	data.BeaconId = task.BeaconId
	data.ByteValue = task.ReqParam
	data.TaskId = task.TaskID

	db.Model(&task).Update(TaskStore{
		Status: statusDispatch,
	})
	return
}

//UpdateTask update task in database
func UpdateTask(taskID uint64, rspParam []byte) (err error) {
	task := TaskStore{}
	db := instance()

	if db.Where("task_id = ? and status = ?", taskID, statusDispatch).First(&task).RecordNotFound() {
		return
	}

	return db.Model(&task).Update(TaskStore{
		RspParam: rspParam,
		Status:   statusDone,
	}).Error
}

//GetTaskRspData return resp data by msgid
func GetTaskRspData(msgID int32) (rspData []TaskStore, err error) {

	db := instance()
	query := db.Where("msg_id = ? and status = ?", msgID, statusDone).Find(&rspData)
	if query.Error != nil {
		err = query.Error
		return
	}
	return
}
