package models

import (
	"fmt"
	"log/slog"
)

type OpCode int

const (
	_ OpCode = iota
	OpCodeCreate
	OpCodeDelete

	opLabelCreate = "CREATE"
	opLabelDelete = "DELETE"
)

type PodOperation struct {
	PodID string
	Code  OpCode
}

func (po PodOperation) LogValue() slog.Value {
	var a string
	switch po.Code {
	case OpCodeCreate:
		a = opLabelCreate
	case OpCodeDelete:
		a = opLabelDelete
	default:
		a = "UNKNOWN"
	}
	return slog.StringValue(fmt.Sprintf("<%s> %s", a, po.PodID))
}

func OpCreate(podID string) PodOperation {
	return PodOperation{
		PodID: podID,
		Code:  OpCodeCreate,
	}
}

func OpDelete(podID string) PodOperation {
	return PodOperation{
		PodID: podID,
		Code:  OpCodeDelete,
	}
}

// UpdateOperations возвращает список операций соответствующих изменению статуса подов.
func UpdateOperations(s, sBefore *Status, needRestart bool) []PodOperation {
	if s == nil || sBefore == nil || s.ID != sBefore.ID {
		return nil
	}

	ops := make([]PodOperation, 0)

	updatePod := func(podName string, isOn, wasOn bool, needRestart bool) {
		podID := fmt.Sprintf("%s-%d", podName, s.ID)
		if isOn != wasOn {
			if isOn {
				ops = append(ops, OpCreate(podID))
			} else {
				ops = append(ops, OpDelete(podID))
			}
		} else if wasOn && needRestart {
			ops = append(ops, OpDelete(podID), OpCreate(podID))
		}
	}

	updatePod("X", s.X, sBefore.X, needRestart)
	updatePod("Y", s.Y, sBefore.Y, needRestart)
	updatePod("Z", s.Z, sBefore.Z, needRestart)

	return ops
}

func DeleteOperations(s *Status) []PodOperation {
	if s == nil {
		return nil
	}
	return UpdateOperations(&Status{ID: s.ID}, s, false)
}
