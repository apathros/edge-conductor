/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package kinddeployer

import (
	"fmt"
	"testing"

	"github.com/undefinedlabs/go-mpatch"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKindDeployer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KindDeployer Suite")
}

//region ----- Ginkgo-mPatch-v2.0.2-stable -----

// patch method used
func (p *patchDatabase) patchMethod(target, redirection interface{}) (*mpatch.Patch, error) {
	return mpatch.PatchMethod(target, redirection)
}

type PatchRequestContext struct {
	requests []*patchRequest
}

func (prl *PatchRequestContext) printDebugSummary() {
	println("===== PATCH LIST =====")
	for _, v := range prl.requests {
		println(v.patchID)
	}
	println("===== END PATCH LIST =====")
}

func (prl *PatchRequestContext) append(i *patchRequest) {
	prl.requests = append(prl.requests, i)
}

func newPatchRequestList() *PatchRequestContext {
	return &PatchRequestContext{make([]*patchRequest, 0)}
}

func (prl *PatchRequestContext) deepcopy() *PatchRequestContext {
	patchRequest := newPatchRequestList()

	for _, v := range prl.requests {
		patchRequest.append(newPatchRequest(v.patchID, v.args))
	}

	return patchRequest
}

type patchRequest struct {
	patchID string
	args    []interface{}
}

func newPatchRequest(patchID string, args []interface{}) *patchRequest {
	return &patchRequest{
		patchID: patchID,
		args:    args,
	}
}

type patchProperty int

const (
	PatchFunc patchProperty = iota
	PatchID
	Counter
	PatchGuard
)

type patchDatabase struct {
	registry    *patchRegistry
	requestList *PatchRequestContext
}

type patchRegistry struct {
	patches map[string]map[patchProperty]interface{}
}

func newPatchRegistry() *patchRegistry {
	pr := &patchRegistry{make(map[string]map[patchProperty]interface{})}
	return pr
}

func (p *patchDatabase) AddPatchRequest(patchID string, args ...interface{}) string {
	p.requestList.append(newPatchRequest(patchID, args))
	return patchID
}

func (p *patchDatabase) getContext() *PatchRequestContext {
	return p.requestList.deepcopy()
}

func (p *patchDatabase) _resolveConflictsInRequestList() *PatchRequestContext {
	newRequestList := newPatchRequestList()

OuterLoop:
	for _, i := range p.requestList.requests {
		for _, j := range newRequestList.requests {
			if j.patchID == i.patchID {
				print("[mPatch error->_resolveConflictsInRequestList]: PatchID: `" + j.patchID + "` already exists, hence ignored")
				continue OuterLoop
			}
		}
		newRequestList.append(i)
	}

	return newRequestList
}

func (p *patchDatabase) DeleteByRequestList(patchRequestList *PatchRequestContext) *patchDatabase {
	for _, v := range patchRequestList.requests {
		p.DeletePatchRequest(v.patchID)
	}

	return p
}

func (p *patchDatabase) DeletePatchRequest(patchID string) *patchDatabase {
	newPatchRequestList := newPatchRequestList()
	for _, v := range p.requestList.requests {
		if v.patchID == patchID {
			continue
		} else {
			newPatchRequestList.append(newPatchRequest(v.patchID, v.args))
		}
	}
	p.requestList = newPatchRequestList
	return p
}

func (p *patchDatabase) CreateNewPatch(patchID string, target, redirection interface{}) *patchDatabase {
	p.registry.patches[patchID] = make(map[patchProperty]interface{})
	p.registry.patches[patchID][PatchFunc] = func(args ...interface{}) int {
		patchguard, err := p.patchMethod(target, redirection)
		p.registry.patches[patchID][PatchGuard] = patchguard
		if err != nil {
			fmt.Print("[mPatch error->CreateNewPatch]: " + patchID + " " + err.Error())
		}
		return len(args)
	}
	return p
}

func (p *patchDatabase) _patchFuncDoesExistInRegistry(patchID string, patchRegistry map[string]map[patchProperty]interface{}) bool {
	for k := range patchRegistry {
		if k == patchID {
			return true
		}
	}
	return false
}

func (p *patchDatabase) UnpatchRequested(prl *PatchRequestContext) *patchDatabase {
	for _, i := range prl.requests {
		if p._patchFuncDoesExistInRegistry(i.patchID, p.registry.patches) == true {
			err := p.registry.patches[i.patchID][PatchGuard].(*mpatch.Patch).Unpatch()
			if err != nil {
				fmt.Print("[" + i.patchID + "]:" + "[mPatch error->unpatchAfterSpecs]: " + err.Error() + "\n")
			}
		} else {
			errorMsg := "========== Warning \n[Warning]: Spec is trying to unpatch PatchID: `" + i.patchID + "`, which the implementation does not exist. \n[Tip]: Check if you invoked with the correct monkeypatch id. \n========== WARNING \n"
			fmt.Print(errorMsg)
		}
	}
	return p
}

func (p *patchDatabase) PatchRequested(prl *PatchRequestContext) *patchDatabase {
	p._resolveConflictsInRequestList()
	for _, i := range prl.requests {
		if p._patchFuncDoesExistInRegistry(i.patchID, p.registry.patches) == true {
			p.registry.patches[i.patchID][Counter] = 0
			p.registry.patches[i.patchID][PatchID] = i.patchID
			p.registry.patches[i.patchID][PatchGuard] = nil
			// This below will patch the requested function
			patchFunc := p.registry.patches[i.patchID][PatchFunc].(func(args ...interface{}) int)
			_ = patchFunc(p.registry, i.args)
		} else {
			errorMsg := "========== PRE CHECK \n[Fatal]: Spec is trying to call PatchID: `" + i.patchID + "`, which the implementation does not exist. \n[Tip]: Check if you invoked with the correct monkeypatch id. \n========== PRE CHECK \n"
			fmt.Print(errorMsg)
			Fail(errorMsg)
		}
	}
	return p
}

func NewPatchService() *patchDatabase {
	return &patchDatabase{
		newPatchRegistry(),
		newPatchRequestList(),
	}
}

//endregion ----- Ginkgo-mPatch -----
