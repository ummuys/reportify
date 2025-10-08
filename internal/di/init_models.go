package di

import (
	"sq/internal/cache"
	"sq/internal/config"
	"sq/internal/convert"
	"sq/internal/repository"
	"sq/internal/secure"
	"sq/internal/service"
	"sq/internal/web/handlers"
)

type Services struct {
	ReportService service.ReportService
}

type Repositorys struct {
	ReportDB    repository.ReportDB
	ReportCache cache.ReportCache
}

type Tools struct {
	Logger        *config.Loggers
	ReportConvert convert.ReportConvert
}

type Handlers struct {
	ReportHandler handlers.ReportHandler
	AuthHandler   handlers.AuthHandler
}

type Secure struct {
	TokenManager   secure.TokenManager
	PasswordHasher secure.PasswordHasher
}
