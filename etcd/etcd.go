/* Copyright 2022 Zinc Labs Inc. and Contributors
*
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

package etcd

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	client "go.etcd.io/etcd/client/v3"

	"test-cluster/config"
)

var timeout = 30 * time.Second

type EtcdStorage struct {
	prefix string
	cli    *client.Client
}

func New(prefix string) *EtcdStorage {
	cli, err := client.New(client.Config{
		Endpoints:   config.Global.Etcd.Endpoints,
		DialTimeout: 5 * time.Second,
		Username:    config.Global.Etcd.Username,
		Password:    config.Global.Etcd.Password,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("open etcd for cluster metadata failed")
	}

	t := &EtcdStorage{
		prefix: prefix,
		cli:    cli,
	}

	go func() {
		eventChan := t.Watch()
		for e := range eventChan {
			for _, ev := range e.Events {
				log.Debug().Str("etcd", "event").Str("type", ev.Type.String()).Str("kv", ev.Kv.String()).Msg("")
			}
		}
	}()

	return t
}

func (t *EtcdStorage) Join(nodeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	lease, _ := t.cli.Lease.Grant(ctx, 10)
	t.cli.Lease.KeepAlive(context.Background(), lease.ID)
	_, err := t.cli.Put(ctx, t.prefix+"/nodes/"+nodeName, "ok", client.WithLease(lease.ID))
	return err
}

func (t *EtcdStorage) Leave(nodeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err := t.cli.Delete(ctx, t.prefix+"/nodes/"+nodeName)
	return err
}

func (t *EtcdStorage) Watch() <-chan client.WatchResponse {
	return t.cli.Watch(context.Background(), t.prefix+"/nodes/", client.WithPrefix())
}
