package models

import (
	"fmt"
	"log/slog"
)

type opCode int

const (
	_ opCode = iota
	OpCodeCreate
	OpCodeDelete

	opLabelNotSet = "NOTSET"
	opLabelCreate = "CREATE"
	opLabelDelete = "DELETE"
)

type PodOperation struct {
	PodID string
	Code  opCode
}

func (po *PodOperation) LogValue() slog.Value {
	if po == nil {
		return slog.StringValue("<NONE>")
	}
	var a string
	switch po.Code {
	case OpCodeCreate:
		a = opLabelCreate
	case OpCodeDelete:
		a = opLabelDelete
	default:
		a = opLabelNotSet
	}
	return slog.StringValue(fmt.Sprintf("<%s> %s", a, po.PodID))
}

func OpCreate(podID string) *PodOperation {
	return &PodOperation{
		PodID: podID,
		Code:  OpCodeCreate,
	}
}

func OpDelete(podID string) *PodOperation {
	return &PodOperation{
		PodID: podID,
		Code:  OpCodeDelete,
	}
}

func UpdateOperations(s, sBefore *Status) []*PodOperation {
	if s == nil || sBefore == nil || s.ID != sBefore.ID {
		return nil
	}
	ops := make([]*PodOperation, 0)
	if s.VWAP != sBefore.VWAP {
		podID := fmt.Sprintf("VWAP-%d", s.ID)
		if s.VWAP {
			ops = append(ops, OpCreate(podID))
		} else {
			ops = append(ops, OpDelete(podID))
		}
	}
	if s.TWAP != sBefore.TWAP {
		podID := fmt.Sprintf("TWAP-%d", s.ID)
		if s.TWAP {
			ops = append(ops, OpCreate(podID))
		} else {
			ops = append(ops, OpDelete(podID))
		}
	}
	if s.HFT != sBefore.HFT {
		podID := fmt.Sprintf("HFT-%d", s.ID)
		if s.HFT {
			ops = append(ops, OpCreate(podID))
		} else {
			ops = append(ops, OpDelete(podID))
		}
	}
	return ops
}

func RestartOperations(s *Status) []*PodOperation {
	if s == nil {
		return nil
	}
	ops := make([]*PodOperation, 0)
	if s.VWAP {
		podID := fmt.Sprintf("VWAP-%d", s.ID)
		ops = append(ops, OpDelete(podID), OpCreate(podID))
	}
	if s.TWAP {
		podID := fmt.Sprintf("TWAP-%d", s.ID)
		ops = append(ops, OpDelete(podID), OpCreate(podID))
	}
	if s.HFT {
		podID := fmt.Sprintf("HFT-%d", s.ID)
		ops = append(ops, OpDelete(podID), OpCreate(podID))
	}
	return ops
}

func DeleteOperations(s *Status) []*PodOperation {
	if s == nil {
		return nil
	}
	ops := make([]*PodOperation, 0)
	if s.VWAP {
		podID := fmt.Sprintf("VWAP-%d", s.ID)
		ops = append(ops, OpDelete(podID))
	}
	if s.TWAP {
		podID := fmt.Sprintf("TWAP-%d", s.ID)
		ops = append(ops, OpDelete(podID))
	}
	if s.HFT {
		podID := fmt.Sprintf("HFT-%d", s.ID)
		ops = append(ops, OpDelete(podID))
	}
	return ops
}
