// Package mocks holds hand-maintained testify mocks for the access
// module-local interfaces (PolicyDecider, UsageMeter). The AGENTS.md
// exception covers module-local interfaces: these mirror the shape mockery
// emits (EXPECT + typed Call wrappers) but are owned here, because Task 11
// (voice) and Task 12 (AI review) consume the same two shapes and there
// must be one definition, one mocks dir.
package mocks

import (
	"context"
	"time"

	"engflex-api/internal/common/enums"
	accessservices "engflex-api/internal/modules/access/services"

	mock "github.com/stretchr/testify/mock"
)

// NewMockPolicyDecider creates a new instance of MockPolicyDecider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPolicyDecider(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPolicyDecider {
	if helper, ok := t.(interface{ Helper() }); ok {
		helper.Helper()
	}

	mock := &MockPolicyDecider{}
	mock.Mock.Test(t)

	t.Cleanup(func() {
		if helper, ok := t.(interface{ Helper() }); ok {
			helper.Helper()
		}
		mock.AssertExpectations(t)
	})

	return mock
}

// MockPolicyDecider is a hand-maintained mock type for the PolicyDecider type
type MockPolicyDecider struct {
	mock.Mock
}

type MockPolicyDecider_Expecter struct {
	mock *mock.Mock
}

func (_m *MockPolicyDecider) EXPECT() *MockPolicyDecider_Expecter {
	return &MockPolicyDecider_Expecter{mock: &_m.Mock}
}

// Decide provides a mock function for the type MockPolicyDecider
func (_mock *MockPolicyDecider) Decide(ctx context.Context, userID string, isAdmin bool, action accessservices.Action) (accessservices.Decision, error) {
	ret := _mock.Called(ctx, userID, isAdmin, action)

	if len(ret) == 0 {
		panic("no return value specified for Decide")
	}

	var r0 accessservices.Decision
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, bool, accessservices.Action) (accessservices.Decision, error)); ok {
		return returnFunc(ctx, userID, isAdmin, action)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, bool, accessservices.Action) accessservices.Decision); ok {
		r0 = returnFunc(ctx, userID, isAdmin, action)
	} else {
		r0 = ret.Get(0).(accessservices.Decision)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, bool, accessservices.Action) error); ok {
		r1 = returnFunc(ctx, userID, isAdmin, action)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockPolicyDecider_Decide_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Decide'
type MockPolicyDecider_Decide_Call struct {
	*mock.Call
}

// Decide is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
//   - isAdmin bool
//   - action accessservices.Action
func (_e *MockPolicyDecider_Expecter) Decide(ctx any, userID any, isAdmin any, action any) *MockPolicyDecider_Decide_Call {
	return &MockPolicyDecider_Decide_Call{Call: _e.mock.On("Decide", ctx, userID, isAdmin, action)}
}

func (_c *MockPolicyDecider_Decide_Call) Run(run func(ctx context.Context, userID string, isAdmin bool, action accessservices.Action)) *MockPolicyDecider_Decide_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		var arg2 bool
		if args[2] != nil {
			arg2 = args[2].(bool)
		}
		var arg3 accessservices.Action
		if args[3] != nil {
			arg3 = args[3].(accessservices.Action)
		}
		run(
			arg0,
			arg1,
			arg2,
			arg3,
		)
	})
	return _c
}

func (_c *MockPolicyDecider_Decide_Call) Return(decision accessservices.Decision, err error) *MockPolicyDecider_Decide_Call {
	_c.Call.Return(decision, err)
	return _c
}

func (_c *MockPolicyDecider_Decide_Call) RunAndReturn(run func(ctx context.Context, userID string, isAdmin bool, action accessservices.Action) (accessservices.Decision, error)) *MockPolicyDecider_Decide_Call {
	_c.Call.Return(run)
	return _c
}

// Location provides a mock function for the type MockPolicyDecider
func (_mock *MockPolicyDecider) Location() *time.Location {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Location")
	}

	var r0 *time.Location
	if returnFunc, ok := ret.Get(0).(func() *time.Location); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*time.Location)
		}
	}
	return r0
}

// MockPolicyDecider_Location_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Location'
type MockPolicyDecider_Location_Call struct {
	*mock.Call
}

// Location is a helper method to define mock.On call
func (_e *MockPolicyDecider_Expecter) Location() *MockPolicyDecider_Location_Call {
	return &MockPolicyDecider_Location_Call{Call: _e.mock.On("Location")}
}

func (_c *MockPolicyDecider_Location_Call) Run(run func()) *MockPolicyDecider_Location_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockPolicyDecider_Location_Call) Return(location *time.Location) *MockPolicyDecider_Location_Call {
	_c.Call.Return(location)
	return _c
}

func (_c *MockPolicyDecider_Location_Call) RunAndReturn(run func() *time.Location) *MockPolicyDecider_Location_Call {
	_c.Call.Return(run)
	return _c
}

// Principal provides a mock function for the type MockPolicyDecider
func (_mock *MockPolicyDecider) Principal(ctx context.Context, userID string, isAdmin bool) (accessservices.Principal, accessservices.TrialInfo, error) {
	ret := _mock.Called(ctx, userID, isAdmin)

	if len(ret) == 0 {
		panic("no return value specified for Principal")
	}

	var r0 accessservices.Principal
	var r1 accessservices.TrialInfo
	var r2 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, bool) (accessservices.Principal, accessservices.TrialInfo, error)); ok {
		return returnFunc(ctx, userID, isAdmin)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, bool) accessservices.Principal); ok {
		r0 = returnFunc(ctx, userID, isAdmin)
	} else {
		r0 = ret.Get(0).(accessservices.Principal)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, bool) accessservices.TrialInfo); ok {
		r1 = returnFunc(ctx, userID, isAdmin)
	} else {
		r1 = ret.Get(1).(accessservices.TrialInfo)
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, string, bool) error); ok {
		r2 = returnFunc(ctx, userID, isAdmin)
	} else {
		r2 = ret.Error(2)
	}
	return r0, r1, r2
}

// MockPolicyDecider_Principal_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Principal'
type MockPolicyDecider_Principal_Call struct {
	*mock.Call
}

// Principal is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
//   - isAdmin bool
func (_e *MockPolicyDecider_Expecter) Principal(ctx any, userID any, isAdmin any) *MockPolicyDecider_Principal_Call {
	return &MockPolicyDecider_Principal_Call{Call: _e.mock.On("Principal", ctx, userID, isAdmin)}
}

func (_c *MockPolicyDecider_Principal_Call) Run(run func(ctx context.Context, userID string, isAdmin bool)) *MockPolicyDecider_Principal_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		var arg2 bool
		if args[2] != nil {
			arg2 = args[2].(bool)
		}
		run(
			arg0,
			arg1,
			arg2,
		)
	})
	return _c
}

func (_c *MockPolicyDecider_Principal_Call) Return(principal accessservices.Principal, trialInfo accessservices.TrialInfo, err error) *MockPolicyDecider_Principal_Call {
	_c.Call.Return(principal, trialInfo, err)
	return _c
}

func (_c *MockPolicyDecider_Principal_Call) RunAndReturn(run func(ctx context.Context, userID string, isAdmin bool) (accessservices.Principal, accessservices.TrialInfo, error)) *MockPolicyDecider_Principal_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockUsageMeter creates a new instance of MockUsageMeter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUsageMeter(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUsageMeter {
	if helper, ok := t.(interface{ Helper() }); ok {
		helper.Helper()
	}

	mock := &MockUsageMeter{}
	mock.Mock.Test(t)

	t.Cleanup(func() {
		if helper, ok := t.(interface{ Helper() }); ok {
			helper.Helper()
		}
		mock.AssertExpectations(t)
	})

	return mock
}

// MockUsageMeter is a hand-maintained mock type for the UsageMeter type
type MockUsageMeter struct {
	mock.Mock
}

type MockUsageMeter_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUsageMeter) EXPECT() *MockUsageMeter_Expecter {
	return &MockUsageMeter_Expecter{mock: &_m.Mock}
}

// Commit provides a mock function for the type MockUsageMeter
func (_mock *MockUsageMeter) Commit(ctx context.Context, reservationID string) error {
	ret := _mock.Called(ctx, reservationID)

	if len(ret) == 0 {
		panic("no return value specified for Commit")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = returnFunc(ctx, reservationID)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockUsageMeter_Commit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Commit'
type MockUsageMeter_Commit_Call struct {
	*mock.Call
}

// Commit is a helper method to define mock.On call
//   - ctx context.Context
//   - reservationID string
func (_e *MockUsageMeter_Expecter) Commit(ctx any, reservationID any) *MockUsageMeter_Commit_Call {
	return &MockUsageMeter_Commit_Call{Call: _e.mock.On("Commit", ctx, reservationID)}
}

func (_c *MockUsageMeter_Commit_Call) Run(run func(ctx context.Context, reservationID string)) *MockUsageMeter_Commit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *MockUsageMeter_Commit_Call) Return(err error) *MockUsageMeter_Commit_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockUsageMeter_Commit_Call) RunAndReturn(run func(ctx context.Context, reservationID string) error) *MockUsageMeter_Commit_Call {
	_c.Call.Return(run)
	return _c
}

// Release provides a mock function for the type MockUsageMeter
func (_mock *MockUsageMeter) Release(ctx context.Context, reservationID string) error {
	ret := _mock.Called(ctx, reservationID)

	if len(ret) == 0 {
		panic("no return value specified for Release")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = returnFunc(ctx, reservationID)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockUsageMeter_Release_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Release'
type MockUsageMeter_Release_Call struct {
	*mock.Call
}

// Release is a helper method to define mock.On call
//   - ctx context.Context
//   - reservationID string
func (_e *MockUsageMeter_Expecter) Release(ctx any, reservationID any) *MockUsageMeter_Release_Call {
	return &MockUsageMeter_Release_Call{Call: _e.mock.On("Release", ctx, reservationID)}
}

func (_c *MockUsageMeter_Release_Call) Run(run func(ctx context.Context, reservationID string)) *MockUsageMeter_Release_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *MockUsageMeter_Release_Call) Return(err error) *MockUsageMeter_Release_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockUsageMeter_Release_Call) RunAndReturn(run func(ctx context.Context, reservationID string) error) *MockUsageMeter_Release_Call {
	_c.Call.Return(run)
	return _c
}

// Reserve provides a mock function for the type MockUsageMeter
func (_mock *MockUsageMeter) Reserve(ctx context.Context, userID string, plan enums.PlanKey, charges []accessservices.Charge, idempotencyBase string, refType enums.UsageRefType, refID string, lease time.Duration, now time.Time, loc *time.Location) (*accessservices.Reservation, error) {
	ret := _mock.Called(ctx, userID, plan, charges, idempotencyBase, refType, refID, lease, now, loc)

	if len(ret) == 0 {
		panic("no return value specified for Reserve")
	}

	var r0 *accessservices.Reservation
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, enums.PlanKey, []accessservices.Charge, string, enums.UsageRefType, string, time.Duration, time.Time, *time.Location) (*accessservices.Reservation, error)); ok {
		return returnFunc(ctx, userID, plan, charges, idempotencyBase, refType, refID, lease, now, loc)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, enums.PlanKey, []accessservices.Charge, string, enums.UsageRefType, string, time.Duration, time.Time, *time.Location) *accessservices.Reservation); ok {
		r0 = returnFunc(ctx, userID, plan, charges, idempotencyBase, refType, refID, lease, now, loc)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*accessservices.Reservation)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, enums.PlanKey, []accessservices.Charge, string, enums.UsageRefType, string, time.Duration, time.Time, *time.Location) error); ok {
		r1 = returnFunc(ctx, userID, plan, charges, idempotencyBase, refType, refID, lease, now, loc)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUsageMeter_Reserve_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Reserve'
type MockUsageMeter_Reserve_Call struct {
	*mock.Call
}

// Reserve is a helper method to define mock.On call
//   - ctx context.Context
//   - userID string
//   - plan enums.PlanKey
//   - charges []accessservices.Charge
//   - idempotencyBase string
//   - refType enums.UsageRefType
//   - refID string
//   - lease time.Duration
//   - now time.Time
//   - loc *time.Location
func (_e *MockUsageMeter_Expecter) Reserve(ctx any, userID any, plan any, charges any, idempotencyBase any, refType any, refID any, lease any, now any, loc any) *MockUsageMeter_Reserve_Call {
	return &MockUsageMeter_Reserve_Call{Call: _e.mock.On("Reserve", ctx, userID, plan, charges, idempotencyBase, refType, refID, lease, now, loc)}
}

func (_c *MockUsageMeter_Reserve_Call) Run(run func(ctx context.Context, userID string, plan enums.PlanKey, charges []accessservices.Charge, idempotencyBase string, refType enums.UsageRefType, refID string, lease time.Duration, now time.Time, loc *time.Location)) *MockUsageMeter_Reserve_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		var arg2 enums.PlanKey
		if args[2] != nil {
			arg2 = args[2].(enums.PlanKey)
		}
		var arg3 []accessservices.Charge
		if args[3] != nil {
			arg3 = args[3].([]accessservices.Charge)
		}
		var arg4 string
		if args[4] != nil {
			arg4 = args[4].(string)
		}
		var arg5 enums.UsageRefType
		if args[5] != nil {
			arg5 = args[5].(enums.UsageRefType)
		}
		var arg6 string
		if args[6] != nil {
			arg6 = args[6].(string)
		}
		var arg7 time.Duration
		if args[7] != nil {
			arg7 = args[7].(time.Duration)
		}
		var arg8 time.Time
		if args[8] != nil {
			arg8 = args[8].(time.Time)
		}
		var arg9 *time.Location
		if args[9] != nil {
			arg9 = args[9].(*time.Location)
		}
		run(
			arg0,
			arg1,
			arg2,
			arg3,
			arg4,
			arg5,
			arg6,
			arg7,
			arg8,
			arg9,
		)
	})
	return _c
}

func (_c *MockUsageMeter_Reserve_Call) Return(reservation *accessservices.Reservation, err error) *MockUsageMeter_Reserve_Call {
	_c.Call.Return(reservation, err)
	return _c
}

func (_c *MockUsageMeter_Reserve_Call) RunAndReturn(run func(ctx context.Context, userID string, plan enums.PlanKey, charges []accessservices.Charge, idempotencyBase string, refType enums.UsageRefType, refID string, lease time.Duration, now time.Time, loc *time.Location) (*accessservices.Reservation, error)) *MockUsageMeter_Reserve_Call {
	_c.Call.Return(run)
	return _c
}
