// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: crudapp/v1/service.proto

package crudappv1

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on CreateRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *CreateRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on CreateRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in CreateRequestMultiError, or
// nil if none found.
func (m *CreateRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *CreateRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := utf8.RuneCountInString(m.GetUserId()); l < 1 || l > 100 {
		err := CreateRequestValidationError{
			field:  "UserId",
			reason: "value length must be between 1 and 100 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if l := utf8.RuneCountInString(m.GetData()); l < 1 || l > 5000 {
		err := CreateRequestValidationError{
			field:  "Data",
			reason: "value length must be between 1 and 5000 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return CreateRequestMultiError(errors)
	}

	return nil
}

// CreateRequestMultiError is an error wrapping multiple validation errors
// returned by CreateRequest.ValidateAll() if the designated constraints
// aren't met.
type CreateRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CreateRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CreateRequestMultiError) AllErrors() []error { return m }

// CreateRequestValidationError is the validation error returned by
// CreateRequest.Validate if the designated constraints aren't met.
type CreateRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CreateRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CreateRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CreateRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CreateRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CreateRequestValidationError) ErrorName() string { return "CreateRequestValidationError" }

// Error satisfies the builtin error interface
func (e CreateRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCreateRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CreateRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CreateRequestValidationError{}

// Validate checks the field values on CreateResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *CreateResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on CreateResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in CreateResponseMultiError,
// or nil if none found.
func (m *CreateResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *CreateResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetPost()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, CreateResponseValidationError{
					field:  "Post",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, CreateResponseValidationError{
					field:  "Post",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetPost()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return CreateResponseValidationError{
				field:  "Post",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return CreateResponseMultiError(errors)
	}

	return nil
}

// CreateResponseMultiError is an error wrapping multiple validation errors
// returned by CreateResponse.ValidateAll() if the designated constraints
// aren't met.
type CreateResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m CreateResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m CreateResponseMultiError) AllErrors() []error { return m }

// CreateResponseValidationError is the validation error returned by
// CreateResponse.Validate if the designated constraints aren't met.
type CreateResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e CreateResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e CreateResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e CreateResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e CreateResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e CreateResponseValidationError) ErrorName() string { return "CreateResponseValidationError" }

// Error satisfies the builtin error interface
func (e CreateResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sCreateResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = CreateResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = CreateResponseValidationError{}

// Validate checks the field values on ReadRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ReadRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ReadRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ReadRequestMultiError, or
// nil if none found.
func (m *ReadRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ReadRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := utf8.RuneCountInString(m.GetUserId()); l < 1 || l > 100 {
		err := ReadRequestValidationError{
			field:  "UserId",
			reason: "value length must be between 1 and 100 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if l := utf8.RuneCountInString(m.GetPostId()); l < 1 || l > 100 {
		err := ReadRequestValidationError{
			field:  "PostId",
			reason: "value length must be between 1 and 100 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return ReadRequestMultiError(errors)
	}

	return nil
}

// ReadRequestMultiError is an error wrapping multiple validation errors
// returned by ReadRequest.ValidateAll() if the designated constraints aren't met.
type ReadRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ReadRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ReadRequestMultiError) AllErrors() []error { return m }

// ReadRequestValidationError is the validation error returned by
// ReadRequest.Validate if the designated constraints aren't met.
type ReadRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ReadRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ReadRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ReadRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ReadRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ReadRequestValidationError) ErrorName() string { return "ReadRequestValidationError" }

// Error satisfies the builtin error interface
func (e ReadRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sReadRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ReadRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ReadRequestValidationError{}

// Validate checks the field values on ReadResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ReadResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ReadResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ReadResponseMultiError, or
// nil if none found.
func (m *ReadResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *ReadResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetPost()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, ReadResponseValidationError{
					field:  "Post",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, ReadResponseValidationError{
					field:  "Post",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetPost()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ReadResponseValidationError{
				field:  "Post",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return ReadResponseMultiError(errors)
	}

	return nil
}

// ReadResponseMultiError is an error wrapping multiple validation errors
// returned by ReadResponse.ValidateAll() if the designated constraints aren't met.
type ReadResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ReadResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ReadResponseMultiError) AllErrors() []error { return m }

// ReadResponseValidationError is the validation error returned by
// ReadResponse.Validate if the designated constraints aren't met.
type ReadResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ReadResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ReadResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ReadResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ReadResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ReadResponseValidationError) ErrorName() string { return "ReadResponseValidationError" }

// Error satisfies the builtin error interface
func (e ReadResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sReadResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ReadResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ReadResponseValidationError{}

// Validate checks the field values on ReadAllRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *ReadAllRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ReadAllRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in ReadAllRequestMultiError,
// or nil if none found.
func (m *ReadAllRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *ReadAllRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := utf8.RuneCountInString(m.GetUserId()); l < 1 || l > 100 {
		err := ReadAllRequestValidationError{
			field:  "UserId",
			reason: "value length must be between 1 and 100 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return ReadAllRequestMultiError(errors)
	}

	return nil
}

// ReadAllRequestMultiError is an error wrapping multiple validation errors
// returned by ReadAllRequest.ValidateAll() if the designated constraints
// aren't met.
type ReadAllRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ReadAllRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ReadAllRequestMultiError) AllErrors() []error { return m }

// ReadAllRequestValidationError is the validation error returned by
// ReadAllRequest.Validate if the designated constraints aren't met.
type ReadAllRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ReadAllRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ReadAllRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ReadAllRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ReadAllRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ReadAllRequestValidationError) ErrorName() string { return "ReadAllRequestValidationError" }

// Error satisfies the builtin error interface
func (e ReadAllRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sReadAllRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ReadAllRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ReadAllRequestValidationError{}

// Validate checks the field values on ReadAllResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *ReadAllResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on ReadAllResponse with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// ReadAllResponseMultiError, or nil if none found.
func (m *ReadAllResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *ReadAllResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetPosts() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, ReadAllResponseValidationError{
						field:  fmt.Sprintf("Posts[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, ReadAllResponseValidationError{
						field:  fmt.Sprintf("Posts[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return ReadAllResponseValidationError{
					field:  fmt.Sprintf("Posts[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	// no validation rules for LastIndex

	if len(errors) > 0 {
		return ReadAllResponseMultiError(errors)
	}

	return nil
}

// ReadAllResponseMultiError is an error wrapping multiple validation errors
// returned by ReadAllResponse.ValidateAll() if the designated constraints
// aren't met.
type ReadAllResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ReadAllResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ReadAllResponseMultiError) AllErrors() []error { return m }

// ReadAllResponseValidationError is the validation error returned by
// ReadAllResponse.Validate if the designated constraints aren't met.
type ReadAllResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ReadAllResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ReadAllResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ReadAllResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ReadAllResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ReadAllResponseValidationError) ErrorName() string { return "ReadAllResponseValidationError" }

// Error satisfies the builtin error interface
func (e ReadAllResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sReadAllResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ReadAllResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ReadAllResponseValidationError{}

// Validate checks the field values on UpsertRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *UpsertRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on UpsertRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in UpsertRequestMultiError, or
// nil if none found.
func (m *UpsertRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *UpsertRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := utf8.RuneCountInString(m.GetUserId()); l < 1 || l > 100 {
		err := UpsertRequestValidationError{
			field:  "UserId",
			reason: "value length must be between 1 and 100 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if l := utf8.RuneCountInString(m.GetPostId()); l < 1 || l > 100 {
		err := UpsertRequestValidationError{
			field:  "PostId",
			reason: "value length must be between 1 and 100 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if l := utf8.RuneCountInString(m.GetData()); l < 1 || l > 5000 {
		err := UpsertRequestValidationError{
			field:  "Data",
			reason: "value length must be between 1 and 5000 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return UpsertRequestMultiError(errors)
	}

	return nil
}

// UpsertRequestMultiError is an error wrapping multiple validation errors
// returned by UpsertRequest.ValidateAll() if the designated constraints
// aren't met.
type UpsertRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m UpsertRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m UpsertRequestMultiError) AllErrors() []error { return m }

// UpsertRequestValidationError is the validation error returned by
// UpsertRequest.Validate if the designated constraints aren't met.
type UpsertRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e UpsertRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e UpsertRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e UpsertRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e UpsertRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e UpsertRequestValidationError) ErrorName() string { return "UpsertRequestValidationError" }

// Error satisfies the builtin error interface
func (e UpsertRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sUpsertRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = UpsertRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = UpsertRequestValidationError{}

// Validate checks the field values on UpsertResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *UpsertResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on UpsertResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in UpsertResponseMultiError,
// or nil if none found.
func (m *UpsertResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *UpsertResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetPost()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, UpsertResponseValidationError{
					field:  "Post",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, UpsertResponseValidationError{
					field:  "Post",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetPost()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return UpsertResponseValidationError{
				field:  "Post",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return UpsertResponseMultiError(errors)
	}

	return nil
}

// UpsertResponseMultiError is an error wrapping multiple validation errors
// returned by UpsertResponse.ValidateAll() if the designated constraints
// aren't met.
type UpsertResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m UpsertResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m UpsertResponseMultiError) AllErrors() []error { return m }

// UpsertResponseValidationError is the validation error returned by
// UpsertResponse.Validate if the designated constraints aren't met.
type UpsertResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e UpsertResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e UpsertResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e UpsertResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e UpsertResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e UpsertResponseValidationError) ErrorName() string { return "UpsertResponseValidationError" }

// Error satisfies the builtin error interface
func (e UpsertResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sUpsertResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = UpsertResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = UpsertResponseValidationError{}

// Validate checks the field values on DeleteRequest with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *DeleteRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeleteRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in DeleteRequestMultiError, or
// nil if none found.
func (m *DeleteRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *DeleteRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if l := utf8.RuneCountInString(m.GetUserId()); l < 1 || l > 100 {
		err := DeleteRequestValidationError{
			field:  "UserId",
			reason: "value length must be between 1 and 100 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if l := utf8.RuneCountInString(m.GetPostId()); l < 1 || l > 100 {
		err := DeleteRequestValidationError{
			field:  "PostId",
			reason: "value length must be between 1 and 100 runes, inclusive",
		}
		if !all {
			return err
		}
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return DeleteRequestMultiError(errors)
	}

	return nil
}

// DeleteRequestMultiError is an error wrapping multiple validation errors
// returned by DeleteRequest.ValidateAll() if the designated constraints
// aren't met.
type DeleteRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeleteRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeleteRequestMultiError) AllErrors() []error { return m }

// DeleteRequestValidationError is the validation error returned by
// DeleteRequest.Validate if the designated constraints aren't met.
type DeleteRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteRequestValidationError) ErrorName() string { return "DeleteRequestValidationError" }

// Error satisfies the builtin error interface
func (e DeleteRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteRequestValidationError{}

// Validate checks the field values on DeleteResponse with the rules defined in
// the proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *DeleteResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on DeleteResponse with the rules defined
// in the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in DeleteResponseMultiError,
// or nil if none found.
func (m *DeleteResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *DeleteResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return DeleteResponseMultiError(errors)
	}

	return nil
}

// DeleteResponseMultiError is an error wrapping multiple validation errors
// returned by DeleteResponse.ValidateAll() if the designated constraints
// aren't met.
type DeleteResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m DeleteResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m DeleteResponseMultiError) AllErrors() []error { return m }

// DeleteResponseValidationError is the validation error returned by
// DeleteResponse.Validate if the designated constraints aren't met.
type DeleteResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DeleteResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DeleteResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DeleteResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DeleteResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DeleteResponseValidationError) ErrorName() string { return "DeleteResponseValidationError" }

// Error satisfies the builtin error interface
func (e DeleteResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDeleteResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DeleteResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DeleteResponseValidationError{}

// Validate checks the field values on Post with the rules defined in the proto
// definition for this message. If any rules are violated, the first error
// encountered is returned, or nil if there are no violations.
func (m *Post) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Post with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in PostMultiError, or nil if none found.
func (m *Post) ValidateAll() error {
	return m.validate(true)
}

func (m *Post) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for UserId

	// no validation rules for PostId

	// no validation rules for Data

	if all {
		switch v := interface{}(m.GetCreatedAt()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, PostValidationError{
					field:  "CreatedAt",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, PostValidationError{
					field:  "CreatedAt",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetCreatedAt()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return PostValidationError{
				field:  "CreatedAt",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if all {
		switch v := interface{}(m.GetUpdatedAt()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, PostValidationError{
					field:  "UpdatedAt",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, PostValidationError{
					field:  "UpdatedAt",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetUpdatedAt()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return PostValidationError{
				field:  "UpdatedAt",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return PostMultiError(errors)
	}

	return nil
}

// PostMultiError is an error wrapping multiple validation errors returned by
// Post.ValidateAll() if the designated constraints aren't met.
type PostMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m PostMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m PostMultiError) AllErrors() []error { return m }

// PostValidationError is the validation error returned by Post.Validate if the
// designated constraints aren't met.
type PostValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e PostValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e PostValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e PostValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e PostValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e PostValidationError) ErrorName() string { return "PostValidationError" }

// Error satisfies the builtin error interface
func (e PostValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sPost.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = PostValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = PostValidationError{}
