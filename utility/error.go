package utility

import (
	"errors"
	"fmt"
)

// ErrorCollector is a function that collects errors and returns the collected errors.
type ErrorCollector []*error

// Collect  adds an error to the collector.
//  @receiver ec
//  @param err
func (ec ErrorCollector) Collect(err error) {
	ec = append(ec, &err)
}

// Error is a wrapper for the error type that implements the ErrorCollector interface.
//  @receiver ec
//  @return string
func (ec ErrorCollector) Error() string {
	var errStr string
	for _, err := range ec {
		errStr += fmt.Sprintf("%s\n", (*err).Error())
	}
	return errStr
}

// String is a wrapper for the string type that implements the ErrorCollector interface.
//  @receiver ec
//  @return string
func (ec ErrorCollector) String() string {
	return ec.Error()
}

// NoError is a wrapper for the error type that implements the ErrorCollector interface.
//  @receiver ec
//  @return bool
func (ec ErrorCollector) NoError() bool {
	return len(ec) == 0
}

// GenErr is a wrapper for the error type that implements the ErrorCollector interface.
//  @receiver ec
//  @return error
func (ec ErrorCollector) GenErr() error {
	return errors.New(ec.Error())
}

// Errorf is a wrapper for the error type that implements the ErrorCollector interface.
//  @receiver ec
//  @param format
//  @param args
func (ec ErrorCollector) Errorf(format string, args ...interface{}) {
	ec.Collect(fmt.Errorf(format, args...))
}
