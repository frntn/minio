/*
 * Minio Cloud Storage, (C) 2016 Minio, Inc.
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

import "strings"

// List of valid event types.
var suppportedEventTypes = map[string]struct{}{
	// Object created event types.
	"s3:ObjectCreated:*":                       {},
	"s3:ObjectCreated:Put":                     {},
	"s3:ObjectCreated:Post":                    {},
	"s3:ObjectCreated:Copy":                    {},
	"s3:ObjectCreated:CompleteMultipartUpload": {},
	// Object removed event types.
	"s3:ObjectRemoved:*":      {},
	"s3:ObjectRemoved:Delete": {},
}

// checkEvent - checks if an event is supported.
func checkEvent(event string) APIErrorCode {
	_, ok := suppportedEventTypes[event]
	if !ok {
		return ErrEventNotification
	}
	return ErrNone
}

// checkEvents - checks given list of events if all of them are valid.
// given if one of them is invalid, this function returns an error.
func checkEvents(events []string) APIErrorCode {
	for _, event := range events {
		if s3Error := checkEvent(event); s3Error != ErrNone {
			return s3Error
		}
	}
	return ErrNone
}

// Valid if filterName is 'prefix'.
func isValidFilterNamePrefix(filterName string) bool {
	return "prefix" == filterName
}

// Valid if filterName is 'suffix'.
func isValidFilterNameSuffix(filterName string) bool {
	return "suffix" == filterName
}

// Is this a valid filterName? - returns true if valid.
func isValidFilterName(filterName string) bool {
	return isValidFilterNamePrefix(filterName) || isValidFilterNameSuffix(filterName)
}

// checkFilterRules - checks given list of filter rules if all of them are valid.
func checkFilterRules(filterRules []filterRule) APIErrorCode {
	ruleSetMap := make(map[string]string)
	// Validate all filter rules.
	for _, filterRule := range filterRules {
		// Unknown filter rule name found, returns an appropriate error.
		if !isValidFilterName(filterRule.Name) {
			return ErrFilterNameInvalid
		}

		// Filter names should not be set twice per notification service
		// configuration, if found return an appropriate error.
		if _, ok := ruleSetMap[filterRule.Name]; ok {
			if isValidFilterNamePrefix(filterRule.Name) {
				return ErrFilterNamePrefix
			} else if isValidFilterNameSuffix(filterRule.Name) {
				return ErrFilterNameSuffix
			} else {
				return ErrFilterNameInvalid
			}
		}

		// Maximum prefix length can be up to 1,024 characters, validate.
		if !IsValidObjectPrefix(filterRule.Value) {
			return ErrFilterPrefixValueInvalid
		}

		// Set the new rule name to keep track of duplicates.
		ruleSetMap[filterRule.Name] = filterRule.Value
	}
	// Success all prefixes validated.
	return ErrNone
}

// checkQueueARN - check if the queue arn is valid.
func checkQueueARN(queueARN string) APIErrorCode {
	if !strings.HasPrefix(queueARN, minioSqs) {
		return ErrARNNotification
	}
	if !strings.HasPrefix(queueARN, minioSqs+serverConfig.GetRegion()+":") {
		return ErrRegionNotification
	}
	return ErrNone
}

// checkLambdaARN - check if the lambda arn is valid.
func checkLambdaARN(lambdaARN string) APIErrorCode {
	if !strings.HasPrefix(lambdaARN, minioLambda) {
		return ErrARNNotification
	}
	if !strings.HasPrefix(lambdaARN, minioLambda+serverConfig.GetRegion()+":") {
		return ErrRegionNotification
	}
	return ErrNone
}

// Validate if we recognize the queue type.
func isValidQueue(sqsARN arnSQS) bool {
	amqpQ := isAMQPQueue(sqsARN)       // Is amqp queue?.
	elasticQ := isElasticQueue(sqsARN) // Is elastic queue?.
	redisQ := isRedisQueue(sqsARN)     // Is redis queue?.
	return amqpQ || elasticQ || redisQ
}

// Validate if we recognize the lambda type.
func isValidLambda(lambdaARN arnLambda) bool {
	return isMinL(lambdaARN) // Is minio lambda?.
}

// Validates account id for input queue ARN.
func isValidQueueID(queueARN string) bool {
	// Unmarshals QueueARN into structured object.
	sqsARN := unmarshalSqsARN(queueARN)
	// AMQP queue.
	if isAMQPQueue(sqsARN) {
		amqpN := serverConfig.GetAMQPNotifyByID(sqsARN.AccountID)
		return amqpN.Enable && amqpN.URL != ""
	} else if isElasticQueue(sqsARN) { // Elastic queue.
		elasticN := serverConfig.GetElasticSearchNotifyByID(sqsARN.AccountID)
		return elasticN.Enable && elasticN.URL != ""

	} else if isRedisQueue(sqsARN) { // Redis queue.
		redisN := serverConfig.GetRedisNotifyByID(sqsARN.AccountID)
		return redisN.Enable && redisN.Addr != ""
	}
	return false
}

// Check - validates queue configuration and returns error if any.
func checkQueueConfig(qConfig queueConfig) APIErrorCode {
	// Check queue arn is valid.
	if s3Error := checkQueueARN(qConfig.QueueARN); s3Error != ErrNone {
		return s3Error
	}

	// Unmarshals QueueARN into structured object.
	sqsARN := unmarshalSqsARN(qConfig.QueueARN)
	// Validate if sqsARN requested any of the known supported queues.
	if !isValidQueue(sqsARN) {
		return ErrARNNotification
	}

	// Validate if the account ID is correct.
	if !isValidQueueID(qConfig.QueueARN) {
		return ErrARNNotification
	}

	// Check if valid events are set in queue config.
	if s3Error := checkEvents(qConfig.Events); s3Error != ErrNone {
		return s3Error
	}

	// Check if valid filters are set in queue config.
	if s3Error := checkFilterRules(qConfig.Filter.Key.FilterRules); s3Error != ErrNone {
		return s3Error
	}

	// Success.
	return ErrNone
}

// Check - validates queue configuration and returns error if any.
func checkLambdaConfig(lConfig lambdaConfig) APIErrorCode {
	// Check queue arn is valid.
	if s3Error := checkLambdaARN(lConfig.LambdaARN); s3Error != ErrNone {
		return s3Error
	}

	// Unmarshals QueueARN into structured object.
	lambdaARN := unmarshalLambdaARN(lConfig.LambdaARN)
	// Validate if lambdaARN requested any of the known supported queues.
	if !isValidLambda(lambdaARN) {
		return ErrARNNotification
	}

	// Check if valid events are set in queue config.
	if s3Error := checkEvents(lConfig.Events); s3Error != ErrNone {
		return s3Error
	}

	// Check if valid filters are set in queue config.
	if s3Error := checkFilterRules(lConfig.Filter.Key.FilterRules); s3Error != ErrNone {
		return s3Error
	}

	// Success.
	return ErrNone
}

// Validates all incoming queue configs, checkQueueConfig validates if the
// input fields for each queues is not malformed and has valid configuration
// information.  If validation fails bucket notifications are not enabled.
func validateQueueConfigs(queueConfigs []queueConfig) APIErrorCode {
	for _, qConfig := range queueConfigs {
		if s3Error := checkQueueConfig(qConfig); s3Error != ErrNone {
			return s3Error
		}
	}
	// Success.
	return ErrNone
}

// Validates all incoming lambda configs, checkLambdaConfig validates if the
// input fields for each queues is not malformed and has valid configuration
// information.  If validation fails bucket notifications are not enabled.
func validateLambdaConfigs(lambdaConfigs []lambdaConfig) APIErrorCode {
	for _, lConfig := range lambdaConfigs {
		if s3Error := checkLambdaConfig(lConfig); s3Error != ErrNone {
			return s3Error
		}
	}
	// Success.
	return ErrNone
}

// Validates all the bucket notification configuration for their validity,
// if one of the config is malformed or has invalid data it is rejected.
// Configuration is never applied partially.
func validateNotificationConfig(nConfig notificationConfig) APIErrorCode {
	if s3Error := validateQueueConfigs(nConfig.QueueConfigs); s3Error != ErrNone {
		return s3Error
	}
	if s3Error := validateLambdaConfigs(nConfig.LambdaConfigs); s3Error != ErrNone {
		return s3Error
	}

	// Add validation for other configurations.
	return ErrNone
}

// Unmarshals input value of AWS ARN format into minioLambda object.
// Returned value represents minio lambda type, currently supported are
// - minio
func unmarshalLambdaARN(lambdaARN string) arnLambda {
	lambda := arnLambda{}
	if !strings.HasPrefix(lambdaARN, minioLambda+serverConfig.GetRegion()+":") {
		return lambda
	}
	lambdaType := strings.TrimPrefix(lambdaARN, minioLambda+serverConfig.GetRegion()+":")
	switch {
	case strings.HasSuffix(lambdaType, lambdaTypeMinio):
		lambda.Type = lambdaTypeMinio
	} // Add more lambda here.
	lambda.AccountID = strings.TrimSuffix(lambdaType, ":"+lambda.Type)
	return lambda
}

// Unmarshals input value of AWS ARN format into minioSqs object.
// Returned value represents minio sqs types, currently supported are
// - amqp
// - elasticsearch
// - redis
func unmarshalSqsARN(queueARN string) (mSqs arnSQS) {
	mSqs = arnSQS{}
	if !strings.HasPrefix(queueARN, minioSqs+serverConfig.GetRegion()+":") {
		return mSqs
	}
	sqsType := strings.TrimPrefix(queueARN, minioSqs+serverConfig.GetRegion()+":")
	switch {
	case strings.HasSuffix(sqsType, queueTypeAMQP):
		mSqs.Type = queueTypeAMQP
	case strings.HasSuffix(sqsType, queueTypeElastic):
		mSqs.Type = queueTypeElastic
	case strings.HasSuffix(sqsType, queueTypeRedis):
		mSqs.Type = queueTypeRedis
	} // Add more queues here.
	mSqs.AccountID = strings.TrimSuffix(sqsType, ":"+mSqs.Type)
	return mSqs
}
