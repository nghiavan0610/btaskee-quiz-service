package repositories

import (
	"github.com/google/wire"
)

var RepositoryProviderSet = wire.NewSet(
	ProvideAuthRepository,
	ProvideUserRepository,
	ProvideQuizRepository,
	ProvideQuestionRepository,
	ProvideSessionRepository,
)
