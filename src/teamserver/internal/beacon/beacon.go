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

package beacon

import (
	"github.com/panjf2000/gnet"
	pb "teamserver/internal/proto/protobuf"
)

// Beacon is a struct to save beacon client connected information
type Beacon struct {
	BeaconId   string       //unique identifier for beacon, created at beacon client by mac addr and connect type
	SessionKey []byte       //communication key between beacon and teamserver
	ConnType   pb.CONN_TYPE //connection type, tcp udp https
	Conn       gnet.Conn    //unique identifier for gnet
}
