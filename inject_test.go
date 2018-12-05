// Copyright 2018 The logrange Authors
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

package linker

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

type (
	A struct {
		fldA       int
		IntVal     int    `inject:"intVal"`
		StrVal     string `inject:"strVal"`
		InfI       I      `inject:"infVal"`
		StructB    b      `inject:""`
		StructBPtr *b     `inject:""`
	}

	b struct {
		IntVal int `inject:"intVal,optional:33"`
	}

	C struct {
		InfI     I `inject:""`
		InfICert I `inject:"certain"`
	}

	I interface {
		FuncI()
	}
)

// b supports I
func (bb *b) FuncI() {

}

func TestInjectDifferentFields(t *testing.T) {
	inj := New()
	inj.SetLogger(stdLogger{})

	a := &A{}
	bp := &b{10}
	bb := b{22}
	inj.Register(Component{Name: "AA", Value: a},
		Component{Name: "infVal", Value: bp},
		Component{Name: "", Value: bb},
		Component{Name: "", Value: bp},
		Component{Name: "intVal", Value: int(33)},
		Component{Name: "strVal", Value: "test"},
	)

	inj.Init(nil)

	if a.IntVal != 33 || a.InfI != bp || a.StrVal != "test" || a.StructB != bb || a.StructBPtr != bp {
		t.Fatal("Struct A is not properly initialized: ", a)
	}
}

func TestInterface(t *testing.T) {
	inj := New()
	inj.SetLogger(stdLogger{})

	a := &A{}
	bp := &b{10}
	var inf I
	inf = bp
	inj.Register(Component{Name: "AA", Value: a},
		Component{Name: "infVal", Value: inf},
		Component{Name: "intVal", Value: int(123)},
		Component{Name: "someStruct", Value: *bp},
		Component{Name: "strVal", Value: "test"},
	)

	inj.Init(nil)

	if a.IntVal != 123 || a.InfI != bp || a.StructBPtr != bp {
		t.Fatal("Struct A is not properly initialized: ", a)
	}
	if bp.IntVal != 123 {
		t.Fatal("bp must be initialized by int value")
	}
}

func TestAmbiguousInit(t *testing.T) {
	inj := New()
	inj.SetLogger(stdLogger{})

	bp := &b{10}
	var inf I
	inf = bp
	bp1 := &b{10}

	inj.Register(
		Component{Name: "", Value: &C{}},
		Component{Name: "intVal", Value: int(33)},
		Component{Name: "certain", Value: inf},
		Component{Name: "", Value: bp1},
	)
	catchPanic(t, "Ambiguous component assignment for the field InfI with type", func() { inj.Init(nil) })
}

func TestOptional(t *testing.T) {
	inj := New()
	inj.SetLogger(stdLogger{})

	bp := &b{10}
	var inf I
	inf = bp

	inj.Register(
		Component{Name: "", Value: inf},
	)
	inj.Init(nil)
	if bp.IntVal != 33 {
		t.Fatal("must be initialized by 33, but it is ", bp)
	}
}

type WrongDedault struct {
	FloatVal   float32 `inject: "wrongDefault,optional:1..23"`
	FloatOkVal float32 `inject: "test,optional:1.23"`
}

func TestOptionalPanics(t *testing.T) {
	inj := New()
	inj.SetLogger(stdLogger{})

	inj.Register(
		Component{Name: "", Value: &WrongDedault{}},
	)

	catchPanic(t, "Could not assign the default value=\"1..23\" to the field FloatVal ", func() { inj.Init(nil) })
}

func TestOptionalFloatOk(t *testing.T) {
	inj := New()
	inj.SetLogger(stdLogger{})

	wd := &WrongDedault{}
	inj.Register(
		Component{Name: "wrongDefault", Value: float32(1.22)},
		Component{Name: "", Value: wd},
	)
	inj.Init(nil)

	if wd.FloatOkVal != float32(1.23) || wd.FloatVal != float32(1.22) {
		t.Fatal("Something wrong with ", wd)
	}
}

type StringStruct struct {
	StringVal string `inject: "some,optional:abc"`
}

func TestStringOptional(t *testing.T) {
	inj := New()
	inj.SetLogger(stdLogger{})

	so := &StringStruct{}

	inj.Register(
		Component{Name: "", Value: so},
	)
	inj.Init(nil)
	if so.StringVal != "abc" {
		t.Fatal("must be initialized by 33, but it is ", so)
	}
}

type SelfReffered struct {
	SRef *SelfReffered `inject:""`
}

func TestCycleSelfReffered(t *testing.T) {
	inj := New()
	inj.SetLogger(stdLogger{})

	inj.Register(
		Component{Name: "", Value: &SelfReffered{}},
	)
	catchPanic(t, "Found a loop in the object graph dependencies.", func() { inj.Init(nil) })
}

type Looper1 struct {
	LP1 *Looper2 `inject:""`
}

type Looper2 struct {
	LP2 *Looper3 `inject:""`
}

type Looper3 struct {
	LP3 *Looper1 `inject:""`
}

func TestLoops(t *testing.T) {
	inj := New()
	inj.SetLogger(stdLogger{})

	inj.Register(
		Component{Name: "", Value: &Looper1{}},
		Component{Name: "", Value: &Looper2{}},
		Component{Name: "", Value: &Looper3{}},
	)
	catchPanic(t, "Found a loop in the object graph dependencies.", func() { inj.Init(nil) })
}

type CompCycle1 struct {
	refCnt   *int
	pstConst int
	init     int
	shutdwn  int
	CC       *CompCycle2 `inject:""`
	waitCtx  bool
}

type CompCycle2 struct {
	refCnt   *int
	pstConst int
	init     int
	shutdwn  int
	err      error
	CC       *CompCycle3 `inject: ""`
}

type CompCycle3 struct {
	refCnt   *int
	pstConst int
	init     int
	shutdwn  int
}

func (cc *CompCycle1) PostConstruct() {
	cc.pstConst = *cc.refCnt
}

func (cc *CompCycle1) Init(ctx context.Context) error {
	if cc.waitCtx {
		select {
		case <-ctx.Done():

		}
		return ctx.Err()
	}
	cc.init = *cc.refCnt
	*cc.refCnt++
	return nil
}

func (cc *CompCycle1) Shutdown() {
	cc.shutdwn = *cc.refCnt
	*cc.refCnt--
}

func (cc *CompCycle2) PostConstruct() {
	cc.pstConst = *cc.refCnt
}

func (cc *CompCycle2) Init(ctx context.Context) error {
	cc.init = *cc.refCnt
	*cc.refCnt++
	return cc.err
}

func (cc *CompCycle2) Shutdown() {
	cc.shutdwn = *cc.refCnt
	*cc.refCnt--
}

func (cc *CompCycle3) PostConstruct() {
	cc.pstConst = *cc.refCnt
}

func (cc *CompCycle3) Init(ctx context.Context) error {
	cc.init = *cc.refCnt
	*cc.refCnt++
	return nil
}

func (cc *CompCycle3) Shutdown() {
	cc.shutdwn = *cc.refCnt
	*cc.refCnt--
}

func TestCompCycle(t *testing.T) {
	i := 1
	cc1 := &CompCycle1{refCnt: &i}
	cc2 := &CompCycle2{refCnt: &i}
	cc3 := &CompCycle3{refCnt: &i}

	inj := New()
	inj.SetLogger(stdLogger{})

	inj.Register(
		Component{Name: "", Value: cc2},
		Component{Name: "", Value: cc1},
		Component{Name: "", Value: cc3},
	)

	inj.Init(nil)
	if cc3.init != 1 || cc2.init != 2 || cc1.init != 3 {
		t.Fatal("Wrong init order ", cc1, cc2, cc3)
	}

	inj.Shutdown()
	if cc3.shutdwn != 2 || cc2.shutdwn != 3 || cc1.shutdwn != 4 {
		t.Fatal("Wrong shutdown order ", cc1, cc2, cc3)
	}

	if cc3.pstConst != 1 || cc2.pstConst != 1 || cc1.pstConst != 1 {
		t.Fatal("Wrong PostConstruct order ", cc1, cc2, cc3)
	}
}

func TestInitCycleFailed(t *testing.T) {
	i := 1
	cc1 := &CompCycle1{refCnt: &i}
	cc2 := &CompCycle2{refCnt: &i}
	cc3 := &CompCycle3{refCnt: &i}

	inj := New()
	inj.SetLogger(stdLogger{})

	inj.Register(
		Component{Name: "", Value: cc2},
		Component{Name: "", Value: cc1},
		Component{Name: "", Value: cc3},
	)

	cc2.err = fmt.Errorf("Tra ta ta")

	catchPanic(t, "", func() { inj.Init(nil) })
	if cc3.init != 1 || cc2.init != 2 || cc1.init != 0 {
		t.Fatal("Wrong init order ", cc1, cc2, cc3)
	}

	if cc3.shutdwn != 3 || cc2.shutdwn != 0 || cc1.shutdwn != 0 {
		t.Fatal("Wrong shutdown order ", cc1, cc2, cc3)
	}

	if cc3.pstConst != 1 || cc2.pstConst != 1 || cc1.pstConst != 1 {
		t.Fatal("Wrong PostConstruct order ", cc1, cc2, cc3)
	}
}

func TestInitCycleClosedCtx(t *testing.T) {
	i := 1
	cc1 := &CompCycle1{refCnt: &i}
	cc2 := &CompCycle2{refCnt: &i}
	cc3 := &CompCycle3{refCnt: &i}

	inj := New()
	inj.SetLogger(stdLogger{})

	inj.Register(
		Component{Name: "", Value: cc2},
		Component{Name: "", Value: cc1},
		Component{Name: "", Value: cc3},
	)

	start := time.Now()
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Millisecond)
	cc1.waitCtx = true
	catchPanic(t, "", func() { inj.Init(ctx) })
	if time.Now().Sub(start) < 5*time.Millisecond {
		t.Fatal("expecting at least 5ms")
	}

	if cc3.init != 1 || cc2.init != 2 || cc1.init != 0 {
		t.Fatal("Wrong init order ", cc1, cc2, cc3)
	}

	if cc3.shutdwn != 2 || cc2.shutdwn != 3 || cc1.shutdwn != 0 {
		t.Fatal("Wrong shutdown order ", cc1, cc2, cc3)
	}

	if cc3.pstConst != 1 || cc2.pstConst != 1 || cc1.pstConst != 1 {
		t.Fatal("Wrong PostConstruct order ", cc1, cc2, cc3)
	}
}

func catchPanic(t *testing.T, pfx string, f func()) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Panic is expected")
		}

		if !strings.HasPrefix(r.(string), pfx) {
			t.Fatal("Expecting panic which starts with ", pfx, ", but got ", r)
		}
	}()

	f()
}
