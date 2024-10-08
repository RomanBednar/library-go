package manifestclient

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Action string

const (
	// this is really a subset of patch, but we treat it separately because it is useful to do so
	ActionApply        Action = "Apply"
	ActionApplyStatus  Action = "ApplyStatus"
	ActionUpdate       Action = "Update"
	ActionUpdateStatus Action = "UpdateStatus"
	ActionCreate       Action = "Create"
	ActionDelete       Action = "Delete"
)

var (
	AllActions = sets.New[Action](
		ActionApply,
		ActionApplyStatus,
		ActionUpdate,
		ActionUpdateStatus,
		ActionCreate,
		ActionDelete,
	)
)

type AllActionsTracker[T SerializedRequestish] struct {
	actionToTracker map[Action]*actionTracker[T]
}

type ActionMetadata struct {
	Action    Action
	GVR       schema.GroupVersionResource
	Namespace string
	Name      string
}

type actionTracker[T SerializedRequestish] struct {
	action Action

	requests []T
}

func NewAllActionsTracker[T SerializedRequestish]() *AllActionsTracker[T] {
	return &AllActionsTracker[T]{
		actionToTracker: make(map[Action]*actionTracker[T]),
	}
}

func (a *AllActionsTracker[T]) AddRequests(requests ...T) {
	for _, request := range requests {
		a.AddRequest(request)
	}
}

func (a *AllActionsTracker[T]) AddRequest(request T) {
	if a.actionToTracker == nil {
		a.actionToTracker = map[Action]*actionTracker[T]{}
	}
	action := request.GetSerializedRequest().Action
	if _, ok := a.actionToTracker[action]; !ok {
		a.actionToTracker[action] = &actionTracker[T]{action: action}
	}
	a.actionToTracker[action].AddRequest(request)
}

func (a *AllActionsTracker[T]) ListActions() []Action {
	return sets.List(sets.KeySet(a.actionToTracker))
}

func (a *AllActionsTracker[T]) RequestsForAction(action Action) []T {
	return a.actionToTracker[action].Mutations()
}

func (a *AllActionsTracker[T]) AllRequests() []T {
	ret := []T{}
	for _, currActionTracker := range a.actionToTracker {
		ret = append(ret, currActionTracker.Mutations()...)
	}
	return ret
}

func (a *AllActionsTracker[T]) DeepCopy() *AllActionsTracker[T] {
	ret := &AllActionsTracker[T]{
		actionToTracker: make(map[Action]*actionTracker[T]),
	}

	for k, v := range a.actionToTracker {
		ret.actionToTracker[k] = v.DeepCopy()
	}
	return ret
}

func (a *actionTracker[T]) AddRequest(request T) {
	if a.action != request.GetSerializedRequest().Action {
		panic("coding error")
	}
	a.requests = append(a.requests, request)
}

func (a *actionTracker[T]) Mutations() []T {
	return a.requests
}

func (a *actionTracker[T]) DeepCopy() *actionTracker[T] {
	ret := &actionTracker[T]{
		action:   a.action,
		requests: make([]T, 0),
	}

	for _, v := range a.requests {
		ret.requests = append(ret.requests, v.DeepCopy().(T))
	}
	return ret
}
