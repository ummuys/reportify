package di

import (
	"github.com/ummuys/reportify/internal/cache"
	"github.com/ummuys/reportify/internal/config"
	"github.com/ummuys/reportify/internal/convert"
	"github.com/ummuys/reportify/internal/repository"
	"github.com/ummuys/reportify/internal/secure"
	"github.com/ummuys/reportify/internal/service"
	"github.com/ummuys/reportify/internal/web/handlers"
)

type Services struct {
	ReportService   service.ReportService
	UserService     service.UserService
	MetadataService service.MetadataService
	AdminService    service.AdminService
}

type Repositories struct {
	ReportDB    repository.ReportDB
	ReportCache cache.ReportCache
	UserDB      repository.UserDB
	MetadataDB  repository.MetadataDB
}

type Tools struct {
	Logger        *config.Loggers
	ReportConvert convert.ReportConvert
}

type Handlers struct {
	ReportHandler   handlers.ReportHandler
	AuthHandler     handlers.AuthHandler
	MetadataHandler handlers.MetadataHandler
	AdminHandler    handlers.AdminHandler
}

type Secure struct {
	TokenManager   secure.TokenManager
	PasswordHasher secure.PasswordHasher
}
