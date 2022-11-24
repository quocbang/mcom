package mock

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"gitlab.kenda.com.tw/kenda/mcom"
)

// Script definition.
type Script struct {
	Name   FuncName
	Input  Input
	Output Output
}

// Input definition.
type Input struct {
	Request interface{}
	Options []interface{}
}

// Output definition.
type Output struct {
	Response interface{}
	Error    error
}

// FuncName as DataManager method's name.
type FuncName string

// dataManager for mock service.
// our mock service is based on scripting expected behavior compare with actual calling behavior.
type dataManager struct {
	mutex   sync.Mutex
	step    uint
	scripts []Script
}

// New Mock DataManager with mock scripts as input parameter.
func New(scripts []Script) (mcom.DataManager, error) {
	return &dataManager{
		step:    0,
		scripts: scripts,
	}, nil
}

func (dm *dataManager) Close() error {
	// check if all scripts have been executed before Close()
	// and should not have any scripts after Close()
	if len(dm.scripts) != int(dm.step) {
		return fmt.Errorf("should not close dataManager in step-%d "+
			"before all scripts have been executed. total scripts: %d",
			dm.step,
			len(dm.scripts),
		)
	}
	return nil
}

func (dm *dataManager) run(
	ctx context.Context,
	name FuncName,
	req interface{},
	parseOptions func(expectedOpts []interface{}) (*parsedOptions, error),
	checkReply func(willReturnReply interface{}) bool,
) (reply interface{}, err error) {
	script, step, err := dm.nextScript(name)
	if err != nil {
		return nil, err
	}

	o, err := parseOptions(script.Input.Options)
	if err != nil {
		return nil, fmt.Errorf("got error in step-%d: %v", step, err)
	}
	if !reflect.DeepEqual(o.expected, o.actual) {
		return nil, newBadInputOptionError(o.expected, o.actual).ErrorWithStep(step)
	}

	expectedReq := script.Input.Request
	if !reflect.DeepEqual(expectedReq, req) {
		return nil, newBadRequestError(expectedReq, req).ErrorWithStep(step)
	}

	if err = script.Output.Error; err != nil {
		return nil, err
	}

	reply = script.Output.Response
	if !checkReply(reply) {
		return nil, newBadResponseError(step)
	}
	return reply, nil
}

// nextScript will return the cursor's next script (current script) and step.
func (dm *dataManager) nextScript(funcName FuncName) (script Script, step uint, err error) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	step = dm.step

	if int(step) == len(dm.scripts) {
		return script, step, fmt.Errorf("missing script in step-%d", step)
	}

	script = dm.scripts[step]
	dm.step++

	if script.Name != funcName {
		return script, step, fmt.Errorf("execute the wrong script in step-%d, expected method: %s", step, script.Name)
	}
	return
}

func (dm *dataManager) SignInStation(ctx context.Context, req mcom.SignInStationRequest, opts ...mcom.SignInStationOption) error {
	_, err := dm.run(ctx, FuncSignInStation, req, func(expectedOpts []interface{}) (*parsedOptions, error) {
		if len(opts) != len(expectedOpts) {
			return nil, newMismatchInputOptionLengthError(len(expectedOpts), len(opts))
		}
		expectedOptions := make([]mcom.SignInStationOption, len(expectedOpts))
		for i, inputOpt := range expectedOpts {
			o, ok := inputOpt.(mcom.SignInStationOption)
			if !ok {
				return nil, badOptionType("mcom.SignInStationOption")
			}
			expectedOptions[i] = o
		}
		expectedOpt := mcom.ParseSignInStationOptions(expectedOptions)
		actualOpt := mcom.ParseSignInStationOptions(opts)
		return &parsedOptions{
			expected: expectedOpt.VerifyWorkDate(req.WorkDate), actual: actualOpt.VerifyWorkDate(req.WorkDate),
		}, nil
	}, noReply)
	return err
}

type parsedOptions struct{ expected, actual interface{} }

func noReply(willReturnReply interface{}) bool { return willReturnReply == nil }

func noOptions(expectedOpts []interface{}) (*parsedOptions, error) {
	if len(expectedOpts) > 0 {
		return nil, fmt.Errorf("bad script, it is expected no option")
	}
	return new(parsedOptions), nil
}
