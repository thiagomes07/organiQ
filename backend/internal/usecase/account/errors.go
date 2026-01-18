package account

import "errors"

var (
	ErrInvalidUserID             = errors.New("invalid_user_id")
	ErrUserNotFound              = errors.New("user_not_found")
	ErrPlanNotFound              = errors.New("plan_not_found")
	ErrInvalidName               = errors.New("invalid_name")
	ErrInvalidEmail              = errors.New("invalid_email")
	ErrEmailAlreadyExists        = errors.New("email_already_exists")
	ErrNoIntegrationPayload      = errors.New("no_integration_payload")
	ErrWordPressConfigIncomplete = errors.New("wordpress_config_incomplete")
	ErrAnalyticsConfigIncomplete = errors.New("analytics_config_incomplete")
	ErrInvalidPassword           = errors.New("invalid_password")
	ErrIncorrectPassword         = errors.New("incorrect_password")
)
