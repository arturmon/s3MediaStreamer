package rabbitmq

import (
	"context"
	"reflect"
	"unicode"
)

func (s *Service) HandleMessage(ctx context.Context, queueName string, messageBody map[string]interface{}) {
	// example: s3QueueEvent -> S3QueueEvent
	// S3QueueEvent or HandleOtherEvent
	queueName = capitalizeFirstLetter(queueName)

	method := reflect.ValueOf(s).MethodByName(queueName)
	if method.IsValid() {
		// Prepare arguments for method call
		args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(messageBody)}
		// Check that the number of arguments matches the expected one
		if method.Type().NumIn() == len(args) {
			// Call the method with arguments
			method.Call(args)
		} else {
			s.logger.Errorf("Method %s requires %d arguments, but got %d", queueName, method.Type().NumIn(), len(args))
		}
	} else {
		s.logger.Errorf("No handler defined for queue: %s. Message processing skipped.", queueName)
	}
}

func capitalizeFirstLetter(input string) string {
	if len(input) == 0 {
		return input
	}
	runes := []rune(input)

	if unicode.IsUpper(runes[0]) {
		return input
	}

	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
