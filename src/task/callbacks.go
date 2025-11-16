package task

import "fmt"

type CallBack struct {
	taskCallback func(result any)
}

func NewCallBack(callback func(result any)) *CallBack {
	return &CallBack{
		taskCallback: callback,
	}
}

func (cb *CallBack) OnComplete(result any) {
	if cb.taskCallback != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Callback panic recovered: %v\n", r)
				}
			}()
			cb.taskCallback(result)
		}()
	}
}

func (cb *CallBack) OnError(err error) {
	if cb.taskCallback != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Error callback panic recovered: %v\n", r)
				}
			}()
			result := map[string]any{
				"error":  err.Error(),
				"status": "failed",
			}
			cb.taskCallback(result)
		}()
	}
}
