package services

import "github.com/google/wire"

var ServiceProviderSet = wire.NewSet(
	ProvideTokenService,
	ProvideAuthService,
	ProvideUserService,
	ProvideValidationService,
	ProvideQuizService,
	ProvideQuestionService,
	ProvideSessionService,
)
