/*
 * Minio Cloud Storage, (C) 2015, 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

var (
	readWriteBucketActions = []string{
		"s3:GetBucketLocation",
		"s3:ListBucket",
		"s3:ListBucketMultipartUploads",
		// Add more bucket level read-write actions here.
	}
	readWriteObjectActions = []string{
		"s3:AbortMultipartUpload",
		"s3:DeleteObject",
		"s3:GetObject",
		"s3:ListMultipartUploadParts",
		"s3:PutObject",
		// Add more object level read-write actions here.
	}
)

// Write only actions.
var (
	writeOnlyBucketActions = []string{
		"s3:GetBucketLocation",
		"s3:ListBucketMultipartUploads",
		// Add more bucket level write actions here.
	}
	writeOnlyObjectActions = []string{
		"s3:AbortMultipartUpload",
		"s3:DeleteObject",
		"s3:ListMultipartUploadParts",
		"s3:PutObject",
		// Add more object level write actions here.
	}
)

// Read only actions.
var (
	readOnlyBucketActions = []string{
		"s3:GetBucketLocation",
		"s3:ListBucket",
		// Add more bucket level read actions here.
	}
	readOnlyObjectActions = []string{
		"s3:GetObject",
		// Add more object level read actions here.
	}
)

// Obtain bucket statement for read-write bucketPolicy.
func getReadWriteObjectStatement(bucketName, objectPrefix string) policyStatement {
	objectResourceStatement := policyStatement{}
	objectResourceStatement.Effect = "Allow"
	objectResourceStatement.Principal.AWS = []string{"*"}
	objectResourceStatement.Resources = []string{fmt.Sprintf("%s%s", AWSResourcePrefix, bucketName+"/"+objectPrefix+"*")}
	objectResourceStatement.Actions = readWriteObjectActions
	return objectResourceStatement
}

// Obtain object statement for read-write bucketPolicy.
func getReadWriteBucketStatement(bucketName, objectPrefix string) policyStatement {
	bucketResourceStatement := policyStatement{}
	bucketResourceStatement.Effect = "Allow"
	bucketResourceStatement.Principal.AWS = []string{"*"}
	bucketResourceStatement.Resources = []string{fmt.Sprintf("%s%s", AWSResourcePrefix, bucketName)}
	bucketResourceStatement.Actions = readWriteBucketActions
	return bucketResourceStatement
}

// Obtain statements for read-write bucketPolicy.
func getReadWriteStatement(bucketName, objectPrefix string) []policyStatement {
	statements := []policyStatement{}
	// Save the read write policy.
	statements = append(statements, getReadWriteBucketStatement(bucketName, objectPrefix), getReadWriteObjectStatement(bucketName, objectPrefix))
	return statements
}

// Obtain bucket statement for read only bucketPolicy.
func getReadOnlyBucketStatement(bucketName, objectPrefix string) policyStatement {
	bucketResourceStatement := policyStatement{}
	bucketResourceStatement.Effect = "Allow"
	bucketResourceStatement.Principal.AWS = []string{"*"}
	bucketResourceStatement.Resources = []string{fmt.Sprintf("%s%s", AWSResourcePrefix, bucketName)}
	bucketResourceStatement.Actions = readOnlyBucketActions
	return bucketResourceStatement
}

// Obtain object statement for read only bucketPolicy.
func getReadOnlyObjectStatement(bucketName, objectPrefix string) policyStatement {
	objectResourceStatement := policyStatement{}
	objectResourceStatement.Effect = "Allow"
	objectResourceStatement.Principal.AWS = []string{"*"}
	objectResourceStatement.Resources = []string{fmt.Sprintf("%s%s", AWSResourcePrefix, bucketName+"/"+objectPrefix+"*")}
	objectResourceStatement.Actions = readOnlyObjectActions
	return objectResourceStatement
}

// Obtain statements for read only bucketPolicy.
func getReadOnlyStatement(bucketName, objectPrefix string) []policyStatement {
	statements := []policyStatement{}
	// Save the read only policy.
	statements = append(statements, getReadOnlyBucketStatement(bucketName, objectPrefix), getReadOnlyObjectStatement(bucketName, objectPrefix))
	return statements
}

// Obtain bucket statements for write only bucketPolicy.
func getWriteOnlyBucketStatement(bucketName, objectPrefix string) policyStatement {

	bucketResourceStatement := policyStatement{}
	bucketResourceStatement.Effect = "Allow"
	bucketResourceStatement.Principal.AWS = []string{"*"}
	bucketResourceStatement.Resources = []string{fmt.Sprintf("%s%s", AWSResourcePrefix, bucketName)}
	bucketResourceStatement.Actions = writeOnlyBucketActions
	return bucketResourceStatement
}

// Obtain object statements for write only bucketPolicy.
func getWriteOnlyObjectStatement(bucketName, objectPrefix string) policyStatement {
	objectResourceStatement := policyStatement{}
	objectResourceStatement.Effect = "Allow"
	objectResourceStatement.Principal.AWS = []string{"*"}
	objectResourceStatement.Resources = []string{fmt.Sprintf("%s%s", AWSResourcePrefix, bucketName+"/"+objectPrefix+"*")}
	objectResourceStatement.Actions = writeOnlyObjectActions
	return objectResourceStatement
}

// Obtain statements for write only bucketPolicy.
func getWriteOnlyStatement(bucketName, objectPrefix string) []policyStatement {
	statements := []policyStatement{}
	// Write only policy.
	// Save the write only policy.
	statements = append(statements, getWriteOnlyBucketStatement(bucketName, objectPrefix), getWriteOnlyBucketStatement(bucketName, objectPrefix))
	return statements
}

// Tests validate Action validator.
func TestIsValidActions(t *testing.T) {
	testCases := []struct {
		// input.
		actions []string
		// expected output.
		err error
		// flag indicating whether the test should pass.
		shouldPass bool
	}{
		// Inputs with unsupported Action.
		// Test case - 1.
		// "s3:ListObject" is an invalid Action.
		{[]string{"s3:GetObject", "s3:ListObject", "s3:RemoveObject"}, errors.New("Unsupported action found: ‘s3:ListObject’, please validate your policy document."), false},
		// Test case - 2.
		// Empty Actions.
		{[]string{}, errors.New("Action list cannot be empty."), false},
		// Test case - 3.
		// "s3:DeleteEverything"" is an invalid Action.
		{[]string{"s3:GetObject", "s3:ListBucket", "s3:PutObject", "s3:DeleteEverything"},
			errors.New("Unsupported action found: ‘s3:DeleteEverything’, please validate your policy document."), false},

		// Inputs with valid Action.
		// Test Case - 4.
		{[]string{"s3:*", "*", "s3:GetObject", "s3:ListBucket",
			"s3:PutObject", "s3:GetBucketLocation", "s3:DeleteObject", "s3:AbortMultipartUpload", "s3:ListBucketMultipartUploads", "s3:ListMultipartUploadParts"}, nil, true},
	}
	for i, testCase := range testCases {
		err := isValidActions(testCase.actions)
		if err != nil && testCase.shouldPass {
			t.Errorf("Test %d: Expected to pass, but failed with: <ERROR> %s", i+1, err.Error())
		}
		if err == nil && !testCase.shouldPass {
			t.Errorf("Test %d: Expected to fail with <ERROR> \"%s\", but passed instead", i+1, testCase.err.Error())
		}
		// Failed as expected, but does it fail for the expected reason.
		if err != nil && !testCase.shouldPass {
			if err.Error() != testCase.err.Error() {
				t.Errorf("Test %d: Expected to fail with error \"%s\", but instead failed with error \"%s\"", i+1, testCase.err.Error(), err.Error())
			}
		}

	}
}

// Tests validate Effect validator.
func TestIsValidEffect(t *testing.T) {
	testCases := []struct {
		// input.
		effect string
		// expected output.
		err error
		// flag indicating whether the test should pass.
		shouldPass bool
	}{
		// Inputs with unsupported Effect.
		// Test case - 1.
		{"DontAllow", errors.New("Unsupported Effect found: ‘DontAllow’, please validate your policy document."), false},
		// Test case - 2.
		{"NeverAllow", errors.New("Unsupported Effect found: ‘NeverAllow’, please validate your policy document."), false},
		// Test case - 3.
		{"AllowAlways", errors.New("Unsupported Effect found: ‘AllowAlways’, please validate your policy document."), false},

		// Inputs with valid Effect.
		// Test Case - 4.
		{"Allow", nil, true},
		// Test Case - 5.
		{"Deny", nil, true},
	}
	for i, testCase := range testCases {
		err := isValidEffect(testCase.effect)
		if err != nil && testCase.shouldPass {
			t.Errorf("Test %d: Expected to pass, but failed with: <ERROR> %s", i+1, err.Error())
		}
		if err == nil && !testCase.shouldPass {
			t.Errorf("Test %d: Expected to fail with <ERROR> \"%s\", but passed instead", i+1, testCase.err.Error())
		}
		// Failed as expected, but does it fail for the expected reason.
		if err != nil && !testCase.shouldPass {
			if err.Error() != testCase.err.Error() {
				t.Errorf("Test %d: Expected to fail with error \"%s\", but instead failed with error \"%s\"", i+1, testCase.err.Error(), err.Error())
			}
		}
	}
}

// Tests validate Resources validator.
func TestIsValidResources(t *testing.T) {
	testCases := []struct {
		// input.
		resources []string
		// expected output.
		err error
		// flag indicating whether the test should pass.
		shouldPass bool
	}{
		// Inputs with unsupported Action.
		// Test case - 1.
		// Empty Resources.
		{[]string{}, errors.New("Resource list cannot be empty."), false},
		// Test case - 2.
		// A valid resource should have prefix "arn:aws:s3:::".
		{[]string{"my-resource"}, errors.New("Unsupported resource style found: ‘my-resource’, please validate your policy document."), false},
		// Test case - 3.
		// A Valid resource should have bucket name followed by "arn:aws:s3:::".
		{[]string{"arn:aws:s3:::"}, errors.New("Invalid resource style found: ‘arn:aws:s3:::’, please validate your policy document."), false},
		// Test Case - 4.
		// Valid resource shouldn't have slash('/') followed by "arn:aws:s3:::".
		{[]string{"arn:aws:s3:::/"}, errors.New("Invalid resource style found: ‘arn:aws:s3:::/’, please validate your policy document."), false},

		// Test cases with valid Resources.
		{[]string{"arn:aws:s3:::my-bucket"}, nil, true},
		{[]string{"arn:aws:s3:::my-bucket/Asia/*"}, nil, true},
		{[]string{"arn:aws:s3:::my-bucket/Asia/India/*"}, nil, true},
	}
	for i, testCase := range testCases {
		err := isValidResources(testCase.resources)
		if err != nil && testCase.shouldPass {
			t.Errorf("Test %d: Expected to pass, but failed with: <ERROR> %s", i+1, err.Error())
		}
		if err == nil && !testCase.shouldPass {
			t.Errorf("Test %d: Expected to fail with <ERROR> \"%s\", but passed instead", i+1, testCase.err.Error())
		}
		// Failed as expected, but does it fail for the expected reason.
		if err != nil && !testCase.shouldPass {
			if err.Error() != testCase.err.Error() {
				t.Errorf("Test %d: Expected to fail with error \"%s\", but instead failed with error \"%s\"", i+1, testCase.err.Error(), err.Error())
			}
		}
	}
}

// Tests validate principals validator.
func TestIsValidPrincipals(t *testing.T) {
	testCases := []struct {
		// input.
		principals []string
		// expected output.
		err error
		// flag indicating whether the test should pass.
		shouldPass bool
	}{
		// Inputs with unsupported Principals.
		// Test case - 1.
		// Empty Principals list.
		{[]string{}, errors.New("Principal cannot be empty."), false},
		// Test case - 2.
		// "*" is the only valid principal.
		{[]string{"my-principal"}, errors.New("Unsupported principal style found: ‘my-principal’, please validate your policy document."), false},
		// Test case - 3.
		{[]string{"*", "111122233"}, errors.New("Unsupported principal style found: ‘111122233’, please validate your policy document."), false},

		// Test case - 4.
		// Test case with valid principal value.
		{[]string{"*"}, nil, true},
	}
	for i, testCase := range testCases {
		err := isValidPrincipals(testCase.principals)
		if err != nil && testCase.shouldPass {
			t.Errorf("Test %d: Expected to pass, but failed with: <ERROR> %s", i+1, err.Error())
		}
		if err == nil && !testCase.shouldPass {
			t.Errorf("Test %d: Expected to fail with <ERROR> \"%s\", but passed instead", i+1, testCase.err.Error())
		}
		// Failed as expected, but does it fail for the expected reason.
		if err != nil && !testCase.shouldPass {
			if err.Error() != testCase.err.Error() {
				t.Errorf("Test %d: Expected to fail with error \"%s\", but instead failed with error \"%s\"", i+1, testCase.err.Error(), err.Error())
			}
		}
	}
}

// Tests validate policyStatement condition validator.
func TestIsValidConditions(t *testing.T) {
	// returns empty conditions map.
	setEmptyConditions := func() map[string]map[string]string {
		return make(map[string]map[string]string)
	}

	// returns map with the "StringEquals" set to empty map.
	setEmptyStringEquals := func() map[string]map[string]string {
		emptyMap := make(map[string]string)
		conditions := make(map[string]map[string]string)
		conditions["StringEquals"] = emptyMap
		return conditions

	}

	// returns map with the "StringNotEquals" set to empty map.
	setEmptyStringNotEquals := func() map[string]map[string]string {
		emptyMap := make(map[string]string)
		conditions := make(map[string]map[string]string)
		conditions["StringNotEquals"] = emptyMap
		return conditions

	}
	// Generate conditions.
	generateConditions := func(key1, key2, value string) map[string]map[string]string {
		innerMap := make(map[string]string)
		innerMap[key2] = value
		conditions := make(map[string]map[string]string)
		conditions[key1] = innerMap
		return conditions
	}

	// generate ambigious conditions.
	generateAmbigiousConditions := func() map[string]map[string]string {
		innerMap := make(map[string]string)
		innerMap["s3:prefix"] = "Asia/"
		conditions := make(map[string]map[string]string)
		conditions["StringEquals"] = innerMap
		conditions["StringNotEquals"] = innerMap
		return conditions
	}

	// generate valid and non valid type in the condition map.
	generateValidInvalidConditions := func() map[string]map[string]string {
		innerMap := make(map[string]string)
		innerMap["s3:prefix"] = "Asia/"
		conditions := make(map[string]map[string]string)
		conditions["StringEquals"] = innerMap
		conditions["InvalidType"] = innerMap
		return conditions
	}

	// generate valid and invalid keys for valid types in the same condition map.
	generateValidInvalidConditionKeys := func() map[string]map[string]string {
		innerMapValid := make(map[string]string)
		innerMapValid["s3:prefix"] = "Asia/"
		innerMapInValid := make(map[string]string)
		innerMapInValid["s3:invalid"] = "Asia/"
		conditions := make(map[string]map[string]string)
		conditions["StringEquals"] = innerMapValid
		conditions["StringEquals"] = innerMapInValid
		return conditions
	}

	// List of Conditions used for test cases.
	testConditions := []map[string]map[string]string{
		generateConditions("StringValues", "s3:max-keys", "100"),
		generateConditions("StringEquals", "s3:Object", "100"),
		generateAmbigiousConditions(),
		generateValidInvalidConditions(),
		generateValidInvalidConditionKeys(),
		setEmptyConditions(),
		setEmptyStringEquals(),
		setEmptyStringNotEquals(),
		generateConditions("StringEquals", "s3:prefix", "Asia/"),
		generateConditions("StringEquals", "s3:max-keys", "100"),
		generateConditions("StringNotEquals", "s3:prefix", "Asia/"),
		generateConditions("StringNotEquals", "s3:max-keys", "100"),
	}

	testCases := []struct {
		inputCondition map[string]map[string]string
		// expected result.
		expectedErr error
		// flag indicating whether test should pass.
		shouldPass bool
	}{
		// Malformed conditions.
		// Test case - 1.
		// "StringValues" is an invalid type.
		{testConditions[0], fmt.Errorf("Unsupported condition type 'StringValues', " +
			"please validate your policy document."), false},
		// Test case - 2.
		// "s3:Object" is an invalid key.
		{testConditions[1], fmt.Errorf("Unsupported condition key " +
			"'StringEquals', please validate your policy document."), false},
		// Test case - 3.
		// Test case with Ambigious conditions set.
		{testConditions[2], fmt.Errorf("Ambigious condition values for key 's3:prefix', " +
			"please validate your policy document."), false},
		// Test case - 4.
		// Test case with valid and invalid condition types.
		{testConditions[3], fmt.Errorf("Unsupported condition type 'InvalidType', " +
			"please validate your policy document."), false},
		// Test case - 5.
		// Test case with valid and invalid condition keys.
		{testConditions[4], fmt.Errorf("Unsupported condition key 'StringEquals', " +
			"please validate your policy document."), false},
		// Test cases with valid conditions.
		// Test case - 6.
		{testConditions[5], nil, true},
		// Test case - 7.
		{testConditions[6], nil, true},
		// Test case - 8.
		{testConditions[7], nil, true},
		// Test case - 9.
		{testConditions[8], nil, true},
		// Test case - 10.
		{testConditions[9], nil, true},
		// Test case - 11.
		{testConditions[10], nil, true},
		// Test case 10.
		{testConditions[11], nil, true},
	}
	for i, testCase := range testCases {
		actualErr := isValidConditions(testCase.inputCondition)
		if actualErr != nil && testCase.shouldPass {
			t.Errorf("Test %d: Expected to pass, but failed with: <ERROR> %s", i+1, actualErr.Error())
		}
		if actualErr == nil && !testCase.shouldPass {
			t.Errorf("Test %d: Expected to fail with <ERROR> \"%s\", but passed instead", i+1, testCase.expectedErr.Error())
		}
		// Failed as expected, but does it fail for the expected reason.
		if actualErr != nil && !testCase.shouldPass {
			if actualErr.Error() != testCase.expectedErr.Error() {
				t.Errorf("Test %d: Expected to fail with error \"%s\", but instead failed with error \"%s\"", i+1, testCase.expectedErr.Error(), actualErr.Error())
			}
		}
	}
}

// Tests validate Policy Action and Resource fields.
func TestCheckbucketPolicyResources(t *testing.T) {
	// constructing policy statement without invalidPrefixActions (check bucket-policy-parser.go).
	setValidPrefixActions := func(statements []policyStatement) []policyStatement {
		statements[0].Actions = []string{"s3:DeleteObject", "s3:PutObject"}
		return statements
	}
	// contructing policy statement with recursive resources.
	// should result in ErrMalformedPolicy
	setRecurseResource := func(statements []policyStatement) []policyStatement {
		statements[0].Resources = []string{"arn:aws:s3:::minio-bucket/Asia/*", "arn:aws:s3:::minio-bucket/Asia/India/*"}
		return statements
	}

	// constructing policy statement with lexically close characters.
	// should not result in ErrMalformedPolicy
	setResourceLexical := func(statements []policyStatement) []policyStatement {
		statements[0].Resources = []string{"arn:aws:s3:::minio-bucket/op*", "arn:aws:s3:::minio-bucket/oo*"}
		return statements
	}

	// List of bucketPolicy used for tests.
	bucketAccessPolicies := []bucketPolicy{
		// bucketPolicy - 1.
		// Contains valid read only policy statement.
		{Version: "1.0", Statements: getReadOnlyStatement("minio-bucket", "")},
		// bucketPolicy - 2.
		// Contains valid read-write only policy statement.
		{Version: "1.0", Statements: getReadWriteStatement("minio-bucket", "Asia/")},
		// bucketPolicy - 3.
		// Contains valid write only policy statement.
		{Version: "1.0", Statements: getWriteOnlyStatement("minio-bucket", "Asia/India/")},
		// bucketPolicy - 4.
		// Contains invalidPrefixActions.
		// Since resourcePrefix is not to the bucket-name, it return ErrMalformedPolicy.
		{Version: "1.0", Statements: getReadOnlyStatement("minio-bucket-fail", "Asia/India/")},
		// bucketPolicy - 5.
		// constructing policy statement without invalidPrefixActions (check bucket-policy-parser.go).
		// but bucket part of the resource is not equal to the bucket name.
		// this results in return of ErrMalformedPolicy.
		{Version: "1.0", Statements: setValidPrefixActions(getWriteOnlyStatement("minio-bucket-fail", "Asia/India/"))},
		// bucketPolicy - 6.
		// contructing policy statement with recursive resources.
		// should result in ErrMalformedPolicy
		{Version: "1.0", Statements: setRecurseResource(setValidPrefixActions(getWriteOnlyStatement("minio-bucket", "")))},
		// BucketPolciy - 7.
		// constructing policy statement with non recursive but
		// lexically close resources.
		// should result in ErrNone.
		{Version: "1.0", Statements: setResourceLexical(setValidPrefixActions(getWriteOnlyStatement("minio-bucket", "oo")))},
	}

	testCases := []struct {
		inputPolicy bucketPolicy
		// expected results.
		apiErrCode APIErrorCode
		// Flag indicating whether the test should pass.
		shouldPass bool
	}{
		// Test case - 1.
		{bucketAccessPolicies[0], ErrNone, true},
		// Test case - 2.
		{bucketAccessPolicies[1], ErrNone, true},
		// Test case - 3.
		{bucketAccessPolicies[2], ErrNone, true},
		// Test case - 4.
		// contains invalidPrefixActions (check bucket-policy-parser.go).
		// Resource prefix will not be equal to the bucket name in this case.
		{bucketAccessPolicies[3], ErrMalformedPolicy, false},
		// Test case - 5.
		// actions contain invalidPrefixActions (check bucket-policy-parser.go).
		// Resource prefix bucket part is not equal to the bucket name in this case.
		{bucketAccessPolicies[4], ErrMalformedPolicy, false},
		// Test case - 6.
		// contructing policy statement with recursive resources.
		// should result in ErrPolicyNesting.
		{bucketAccessPolicies[5], ErrPolicyNesting, false},
		// Test case - 7.
		// constructing policy statement with lexically close
		// characters.
		// should result in ErrNone.
		{bucketAccessPolicies[6], ErrNone, true},
	}
	for i, testCase := range testCases {
		apiErrCode := checkBucketPolicyResources("minio-bucket", &testCase.inputPolicy)
		if apiErrCode != ErrNone && testCase.shouldPass {
			t.Errorf("Test %d: Expected to pass, but failed with Errocode %v", i+1, apiErrCode)
		}
		if apiErrCode == ErrNone && !testCase.shouldPass {
			t.Errorf("Test %d: Expected to fail with ErrCode %v, but passed instead", i+1, testCase.apiErrCode)
		}
		// Failed as expected, but does it fail for the expected reason.
		if apiErrCode != ErrNone && !testCase.shouldPass {
			if testCase.apiErrCode != apiErrCode {
				t.Errorf("Test %d: Expected to fail with error code %v, but instead failed with error code %v", i+1, testCase.apiErrCode, apiErrCode)
			}
		}
	}
}

// Tests validate parsing of BucketAccessPolicy.
func TestParseBucketPolicy(t *testing.T) {
	// set Unsupported Actions.
	setUnsupportedActions := func(statements []policyStatement) []policyStatement {
		// "s3:DeleteEverything"" is an Unsupported Action.
		statements[0].Actions = []string{"s3:GetObject", "s3:ListBucket", "s3:PutObject", "s3:DeleteEverything"}
		return statements
	}
	// set unsupported Effect.
	setUnsupportedEffect := func(statements []policyStatement) []policyStatement {
		// Effect "Don't allow" is Unsupported.
		statements[0].Effect = "DontAllow"
		return statements
	}
	// set unsupported principals.
	setUnsupportedPrincipals := func(statements []policyStatement) []policyStatement {
		// "User1111"" is an Unsupported Principal.
		statements[0].Principal.AWS = []string{"*", "User1111"}
		return statements
	}
	// set unsupported Resources.
	setUnsupportedResources := func(statements []policyStatement) []policyStatement {
		// "s3:DeleteEverything"" is an Unsupported Action.
		statements[0].Resources = []string{"my-resource"}
		return statements
	}
	// List of bucketPolicy used for test cases.
	bucketAccesPolicies := []bucketPolicy{
		// bucketPolicy - 0.
		// bucketPolicy statement empty.
		{Version: "1.0"},
		// bucketPolicy - 1.
		// bucketPolicy version empty.
		{Version: "", Statements: []policyStatement{}},
		// bucketPolicy - 2.
		// Readonly bucketPolicy.
		{Version: "1.0", Statements: getReadOnlyStatement("minio-bucket", "")},
		// bucketPolicy - 3.
		// Read-Write bucket policy.
		{Version: "1.0", Statements: getReadWriteStatement("minio-bucket", "Asia/")},
		// bucketPolicy - 4.
		// Write only bucket policy.
		{Version: "1.0", Statements: getWriteOnlyStatement("minio-bucket", "Asia/India/")},
		// bucketPolicy - 5.
		// bucketPolicy statement contains unsupported action.
		{Version: "1.0", Statements: setUnsupportedActions(getReadOnlyStatement("minio-bucket", ""))},
		// bucketPolicy - 6.
		// bucketPolicy statement contains unsupported Effect.
		{Version: "1.0", Statements: setUnsupportedEffect(getReadWriteStatement("minio-bucket", "Asia/"))},
		// bucketPolicy - 7.
		// bucketPolicy statement contains unsupported Principal.
		{Version: "1.0", Statements: setUnsupportedPrincipals(getWriteOnlyStatement("minio-bucket", "Asia/India/"))},
		// bucketPolicy - 8.
		// bucketPolicy statement contains unsupported Resource.
		{Version: "1.0", Statements: setUnsupportedResources(getWriteOnlyStatement("minio-bucket", "Asia/India/"))},
	}

	testCases := []struct {
		inputPolicy bucketPolicy
		// expected results.
		expectedPolicy bucketPolicy
		err            error
		// Flag indicating whether the test should pass.
		shouldPass bool
	}{
		// Test case - 1.
		// bucketPolicy statement empty.
		{bucketAccesPolicies[0], bucketPolicy{}, errors.New("Policy statement cannot be empty."), false},
		// Test case - 2.
		// bucketPolicy version empty.
		{bucketAccesPolicies[1], bucketPolicy{}, errors.New("Policy version cannot be empty."), false},
		// Test case - 3.
		// Readonly bucketPolicy.
		{bucketAccesPolicies[2], bucketAccesPolicies[2], nil, true},
		// Test case - 4.
		// Read-Write bucket policy.
		{bucketAccesPolicies[3], bucketAccesPolicies[3], nil, true},
		// Test case - 5.
		// Write only bucket policy.
		{bucketAccesPolicies[4], bucketAccesPolicies[4], nil, true},
		// Test case - 6.
		// bucketPolicy statement contains unsupported action.
		{bucketAccesPolicies[5], bucketAccesPolicies[5], fmt.Errorf("Unsupported action found: ‘s3:DeleteEverything’, please validate your policy document."), false},
		// Test case - 7.
		// bucketPolicy statement contains unsupported Effect.
		{bucketAccesPolicies[6], bucketAccesPolicies[6], fmt.Errorf("Unsupported Effect found: ‘DontAllow’, please validate your policy document."), false},
		// Test case - 8.
		// bucketPolicy statement contains unsupported Principal.
		{bucketAccesPolicies[7], bucketAccesPolicies[7], fmt.Errorf("Unsupported principal style found: ‘User1111’, please validate your policy document."), false},
		// Test case - 9.
		// bucketPolicy statement contains unsupported Resource.
		{bucketAccesPolicies[8], bucketAccesPolicies[8], fmt.Errorf("Unsupported resource style found: ‘my-resource’, please validate your policy document."), false},
	}
	for i, testCase := range testCases {
		var buffer bytes.Buffer
		encoder := json.NewEncoder(&buffer)
		err := encoder.Encode(testCase.inputPolicy)
		if err != nil {
			t.Fatalf("Test %d: Couldn't Marshal bucket policy %s", i+1, err)
		}

		var actualAccessPolicy = &bucketPolicy{}
		err = parseBucketPolicy(&buffer, actualAccessPolicy)
		if err != nil && testCase.shouldPass {
			t.Errorf("Test %d: Expected to pass, but failed with: <ERROR> %s", i+1, err.Error())
		}
		if err == nil && !testCase.shouldPass {
			t.Errorf("Test %d: Expected to fail with <ERROR> \"%s\", but passed instead", i+1, testCase.err.Error())
		}
		// Failed as expected, but does it fail for the expected reason.
		if err != nil && !testCase.shouldPass {
			if err.Error() != testCase.err.Error() {
				t.Errorf("Test %d: Expected to fail with error \"%s\", but instead failed with error \"%s\"", i+1, testCase.err.Error(), err.Error())
			}
		}
		// Test passes as expected, but the output values are verified for correctness here.
		if err == nil && testCase.shouldPass {
			if testCase.expectedPolicy.String() != actualAccessPolicy.String() {
				t.Errorf("Test %d: The expected statements from resource statement generator doesn't match the actual statements", i+1)
			}
		}
	}
}
