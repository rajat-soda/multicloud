// Copyright (c) 2018 Huawei Technologies Co., Ltd. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mongo

import (
	"errors"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/micro/go-log"
	. "github.com/opensds/multi-cloud/dataflow/pkg/model"
)

var adap = &adapter{}

var DataBaseName = "multi-cloud"
var CollJob = "job"

func Init(host string) *adapter {
	session, err := mgo.Dial(host)
	if err != nil {
		panic(err)
	}

	session.SetMode(mgo.Monotonic, true)
	adap.s = session
	adap.userID = "unknown"

	return adap
}

func Exit() {
	adap.s.Close()
}

type adapter struct {
	s      *mgo.Session
	userID string
}

func (ad *adapter) GetJobStatus(jobID string) string {
	job := Job{}
	ss := ad.s.Copy()
	defer ss.Close()
	c := ss.DB(DataBaseName).C(CollJob)

	err := c.Find(bson.M{"_id": bson.ObjectIdHex(jobID)}).One(&job)
	if err != nil {
		log.Logf("Get job[ID#%s] failed:%v.\n", jobID, err)
		return ""
	}

	return job.Status
}

func (ad *adapter) UpdateJob(job *Job) error {
	ss := ad.s.Copy()
	defer ss.Close()

	c := ss.DB(DataBaseName).C(CollJob)
	j := Job{}
	err := c.Find(bson.M{"_id": job.Id}).One(&j)
	if err != nil {
		log.Logf("Get job[id:%v] failed before update it, err:%v\n", job.Id, err)

		return errors.New("Get job failed before update it.")
	}

	if !job.StartTime.IsZero() {
		j.StartTime = job.StartTime
	}
	if !job.EndTime.IsZero() {
		j.EndTime = job.EndTime
	}
	if job.TotalCapacity != 0 {
		j.TotalCapacity = job.TotalCapacity
	}
	if job.TotalCount != 0 {
		j.TotalCount = job.TotalCount
	}
	if job.PassedCount != 0 {
		j.PassedCount = job.PassedCount
	}
	if job.PassedCapacity != 0 {
		j.PassedCapacity = job.PassedCapacity
	}
	if job.Status != "" {
		j.Status = job.Status
	}

	err = c.Update(bson.M{"_id": j.Id}, &j)
	if err != nil {
		log.Logf("Update job in database failed, err:%v\n", err)
		return errors.New("Update job in database failed.")
	}

	log.Log("Update job in database successfully.")
	return nil
}
