package models

import (
	"database/sql/driver"
	"engflex-api/internal/common/enums"
	"github.com/lib/pq"
)

type GoalArray []enums.Goal

func (a GoalArray) Value() (driver.Value, error) { return pq.Array([]enums.Goal(a)).Value() }
func (a *GoalArray) Scan(value any) error {
	var values pq.StringArray
	if err := values.Scan(value); err != nil {
		return err
	}
	if values == nil {
		*a = nil
		return nil
	}
	*a = make(GoalArray, len(values))
	for i, value := range values {
		(*a)[i] = enums.Goal(value)
	}
	return nil
}

type InterestTopicArray []enums.InterestTopic

func (a InterestTopicArray) Value() (driver.Value, error) {
	return pq.Array([]enums.InterestTopic(a)).Value()
}
func (a *InterestTopicArray) Scan(value any) error {
	var values pq.StringArray
	if err := values.Scan(value); err != nil {
		return err
	}
	if values == nil {
		*a = nil
		return nil
	}
	*a = make(InterestTopicArray, len(values))
	for i, value := range values {
		(*a)[i] = enums.InterestTopic(value)
	}
	return nil
}
