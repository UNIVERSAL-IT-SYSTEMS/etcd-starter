// Copyright 2015 CoreOS, Inc.
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

package security

import (
	"encoding/json"
	"log"
	"path"

	etcderr "github.com/coreos/etcd-starter/Godeps/_workspace/src/github.com/coreos/etcd/error"
	"github.com/coreos/etcd-starter/Godeps/_workspace/src/github.com/coreos/etcd/etcdserver"
	"github.com/coreos/etcd-starter/Godeps/_workspace/src/github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/etcd-starter/Godeps/_workspace/src/golang.org/x/net/context"
)

func (s *Store) ensureSecurityDirectories() error {
	for _, res := range []string{StorePermsPrefix, StorePermsPrefix + "/users/", StorePermsPrefix + "/roles/"} {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()
		pe := false
		rr := etcdserverpb.Request{
			Method:    "PUT",
			Path:      res,
			Dir:       true,
			PrevExist: &pe,
		}
		_, err := s.server.Do(ctx, rr)
		if err != nil {
			if e, ok := err.(*etcderr.Error); ok {
				if e.ErrorCode == etcderr.EcodeNodeExist {
					continue
				}
			}
			log.Println("security: Trying to create security directories in the store; failed:", err)
			return err
		}
	}
	return nil
}

func (s *Store) detectSecurity() bool {
	if s.server == nil {
		return false
	}
	_, err := s.requestResource("/users/root", false)
	if err == nil {
		return true
	}
	return false
}

func (s *Store) requestResource(res string, dir bool) (etcdserver.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	p := path.Join(StorePermsPrefix, res)
	rr := etcdserverpb.Request{
		Method: "GET",
		Path:   p,
		Dir:    dir,
	}
	return s.server.Do(ctx, rr)
}

func (s *Store) updateResource(res string, value interface{}) (etcdserver.Response, error) {
	return s.setResource(res, value, true)
}
func (s *Store) createResource(res string, value interface{}) (etcdserver.Response, error) {
	return s.setResource(res, value, false)
}
func (s *Store) setResource(res string, value interface{}, prevexist bool) (etcdserver.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	data, err := json.Marshal(value)
	if err != nil {
		return etcdserver.Response{}, err
	}
	p := path.Join(StorePermsPrefix, res)
	rr := etcdserverpb.Request{
		Method:    "PUT",
		Path:      p,
		Val:       string(data),
		PrevExist: &prevexist,
	}
	return s.server.Do(ctx, rr)
}

func (s *Store) deleteResource(res string) (etcdserver.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	pex := true
	p := path.Join(StorePermsPrefix, res)
	rr := etcdserverpb.Request{
		Method:    "DELETE",
		Path:      p,
		PrevExist: &pex,
	}
	return s.server.Do(ctx, rr)
}
