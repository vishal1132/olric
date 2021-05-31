// Copyright 2018-2021 Burak Sezer
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

package journal

import (
	"os"
	"testing"
	"time"

	"github.com/buraksezer/olric/internal/testutil"
	"github.com/cespare/xxhash"
)

func TestJournal_Append(t *testing.T) {
	f, err := testutil.CreateTmpfile(t, nil)
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}

	c := &Config{
		Path: f.Name(),
	}
	j, err := New(c)
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}
	defer func() {
		err = j.Close()
		if err != nil {
			t.Fatalf("Expected nil. Got: %v", err)
		}
	}()

	for i := 0; i < 100; i++ {
		e := NewMockEntry()
		e.SetKey(testutil.ToKey(i))
		e.SetValue(testutil.ToVal(i))
		e.SetTimestamp(time.Now().UnixNano())
		e.SetTTL(18071988)
		hkey := xxhash.Sum64String(testutil.ToKey(i))

		err = j.Append(OpPut, hkey, e)
		if err != nil {
			t.Fatalf("Expected nil. Got: %v", err)
		}
		err = j.Append(OpUpdateTTL, hkey, e)
		if err != nil {
			t.Fatalf("Expected nil. Got: %v", err)
		}
		err = j.Append(OpDelete, hkey, e)
		if err != nil {
			t.Fatalf("Expected nil. Got: %v", err)
		}
	}

	s := j.Stats()
	if s.QueueLen == 300 {
		t.Fatalf("Expected s.QueueLen: 300. Got: %d", s.QueueLen)
	}

	// Start consuming
	err = j.Start()
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}

	// Relatively long for 100 entries
	<-time.After(250 * time.Millisecond)

	s = j.Stats()
	if s.Put != 100 {
		t.Fatalf("Expected s.Put: 100. Got: %d", s.Put)
	}
	if s.UpdateTTL != 100 {
		t.Fatalf("Expected s.Put: 100. Got: %d", s.UpdateTTL)
	}
	if s.Delete != 100 {
		t.Fatalf("Expected s.Put: 100. Got: %d", s.Delete)
	}
	if s.QueueLen != 0 {
		t.Fatalf("Expected s.QueueLen: 0. Got: %d", s.QueueLen)
	}
}

func TestJournal_Destroy(t *testing.T) {
	f, err := testutil.CreateTmpfile(t, nil)
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}

	c := &Config{
		Path: f.Name(),
	}
	j, err := New(c)
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}
	defer func() {
		err = j.Close()
		if err != nil {
			t.Fatalf("Expected nil. Got: %v", err)
		}
	}()

	e := NewMockEntry()
	e.SetKey("mykey")
	e.SetValue([]byte("myvalue"))
	e.SetTimestamp(time.Now().UnixNano())
	e.SetTTL(18071988)
	hkey := xxhash.Sum64String("mykey")
	err = j.Append(OpPut, hkey, e)
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}
	err = j.Destroy()
	if err != nil {
		t.Fatalf("Expected nil. Got: %v", err)
	}
	_, err = os.Stat(j.file.Name())
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("Journal file still exists: %s", err)
	}
}