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
	"github.com/golang/protobuf/proto"
	pb "teamserver/internal/proto/protobuf"
	"time"
)

//beaconStore struct represents a beacon info to save beacon in database
type beaconStore struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	BeaconId  string `gorm:"primary_key"`
	IpAddr    string
	Detail    []byte
}

//UpdateBeacon update beacon info in database
func UpdateBeacon(beaconId string, ipAddr string, detail []byte) (err error) {
	beacon := beaconStore{
		BeaconId: beaconId,
		IpAddr:   ipAddr,
		Detail:   detail,
	}

	db := instance()
	db.AutoMigrate(&beaconStore{})
	if err = db.Create(&beacon).Error; err == nil {
		return
	}

	beaconOld := beaconStore{}

	if db.First(&beaconOld, "beacon_id = ?", beaconId).RecordNotFound() {
		return errors.New("beacon id not found")
	}

	return db.Model(&beaconOld).Update(beaconStore{
		IpAddr: ipAddr,
		Detail: detail,
	}).Error
}

//DeleteBeacon delete beacon info by beaconid, only modify the delete flag not delete
func DeleteBeacon(beaconId string) (err error) {

	db := instance()
	return db.Where("beacon_id = ?", beaconId).Delete(&beaconStore{}).Error
}

//GetBeacons return all beacons info in database
func GetBeacons() (beaconsData []byte, err error) {
	var beacon []beaconStore

	db := instance()
	query := db.Find(&beacon)
	if query.Error != nil {
		err = query.Error
		return
	}

	rsp := &pb.BeaconsRsp{}
	for _, v := range beacon {
		value := &pb.MapValueData{}
		proto.Unmarshal(v.Detail, value)
		var detail string
		for k, v := range value.DictValue {
			detail = detail + k + ":" + v + ", "
		}

		b := &pb.BeaconInfo{
			CreateTm:   v.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdateTm:   v.UpdatedAt.Format("2006-01-02 15:04:05"),
			Ipaddr:     v.IpAddr,
			BeaconId:   v.BeaconId,
			DetailInfo: detail,
		}
		rsp.Beacon = append(rsp.Beacon, b)
	}
	beaconsData, err = proto.Marshal(rsp)
	return
}
