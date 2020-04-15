package query

import (
	"fmt"
	"github.com/efigence/rodrev/common"
	"github.com/glycerine/zygomys/zygo"
)


type Engine struct {
	r *common.Runtime
}


func NewQueryEngine(r *common.Runtime) *Engine {
	if r == nil {
		panic("need runtime")
	}
	return &Engine{r: r}
}


// ParseBool parses query and returns true if return is true or nonempty string, or > 0 numeric value
func (e *Engine) ParseBool(q string) (bool,error) {
	vars := e.r.Cfg.NodeMeta

	zg := zygo.NewZlisp()
	varsLisp, err := zygo.GoToSexp(vars, zg)
	if err != nil {
		return false, err
	}
	zg.AddGlobal("node", varsLisp)
	err = zg.LoadString(q)
	if err != nil {
		return false, fmt.Errorf("error parsing query [%s]: %s", q, err)
	}
	iters := 0
	zg.AddPreHook(
		func(zg *zygo.Zlisp, s string, se []zygo.Sexp) {
			iters++
			if iters > 1000 {
				zg.Clear()
			}
		})
	expr, err := zg.Run()
	zg.Stop()
	if iters >= 1000 {
		return false, fmt.Errorf("query [%s]: iterations limit exceeded: %d", q, iters)
	}
	if err != nil {
		return false, fmt.Errorf("error running query [%s]: %s", q, err)
	}
	switch v := expr.(type) {
	case *zygo.SexpBool:
		return v.Val, nil
	case *zygo.SexpStr:
		if len(v.S) > 0 {
			return true, nil
		} else {
			return false, nil
		}
	case *zygo.SexpInt:
		if v.Val > 0 {
			return true, nil
		} else {
			return false, nil
		}
	default:
		return false, fmt.Errorf("query return type %+v[%T] not supported, make your query return bool (or string/int > 0)",expr,expr)
	}
}
