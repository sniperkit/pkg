/*
Sniperkit-Bot
- Status: analyzed
*/

// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
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

package binlogsync_test

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/alecthomas/assert"

	"github.com/sniperkit/snk.fork.corestoreio-pkg/sql/binlogsync"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/sql/ddl"
	"github.com/sniperkit/snk.fork.corestoreio-pkg/sql/dml"
)

// TODO(CyS): Add more tests

var (
	runIntegration = flag.Bool("integration", false, "Enables MySQL/MariaDB integration tests, env var CS_DSN must be set")
)

func TestIntegrationNewCanal(t *testing.T) {
	if !*runIntegration {
		t.Skip("Skipping integration tests. You can enable them with via CLI option `-integration`")
	}

	dsn := os.Getenv(dml.EnvDSN)
	if dsn == "" {
		t.Skipf("Skipping integration test because environment variable %q not set.", dml.EnvDSN)
	}

	c, err := binlogsync.NewCanal(dsn, binlogsync.WithMySQL())
	if err != nil {
		t.Fatalf("%+v", err)
	}

	c.RegisterRowsEventHandler(catalogProductEvent{idx: 1001, t: t})
	// c.RegisterRowsEventHandler(catalogProductEvent{idx: 1002, t: t})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	cancel()
	assert.NoError(t, c.Start(ctx), "Start error")

	select {
	case <-ctx.Done():
		println("Closing", ctx.Err().Error())
		assert.NoError(t, c.Close(), "c.Close()")
	}

}

type catalogProductEvent struct {
	idx int
	t   *testing.T
}

func (cpe catalogProductEvent) Do(_ context.Context, action string, table ddl.Table, rows [][]interface{}) error {
	sl := time.Duration(rand.Intn(100)) * time.Millisecond
	time.Sleep(sl)

	cpe.t.Logf("%d Sleep: %s => %q.%q", cpe.idx, sl, table.Schema, table.Name)
	for _, r := range rows {
		cpe.t.Logf("%#v", r)
	}
	cpe.t.Logf("\n")
	return nil
}
func (cpe catalogProductEvent) Complete(_ context.Context) error {
	return nil // errors.NewFatalf("[test] What is incomplete?")
}
func (cpe catalogProductEvent) String() string {
	return "WTF? catalogProductEvent"
}
