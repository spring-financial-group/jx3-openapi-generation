package packageGenerator

import (
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/angular"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/csharp"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/java"
	"spring-financial-group/jx3-openapi-generation/pkg/packageGenerator/python"
)

type Factory struct {
	Version     string
	ServiceName string
	RepoOwner   string
	RepoName    string
	GitToken string
}

func NewFactory(version, serviceName, repoOwner, repoName, gitToken string) *Factory {
	return &Factory{Version: version, ServiceName: serviceName, RepoOwner: repoOwner, RepoName: repoName, GitToken: gitToken}
}

func (f *Factory) NewGenerator(language string) (domain.PackageGenerator, error) {
	var generator domain.PackageGenerator
	switch language {
	case domain.CSharp:
		generator = csharp.NewGenerator(f.Version, f.ServiceName, f.RepoOwner, f.RepoName)
	case domain.Java:
		generator = java.NewGenerator(f.Version, f.ServiceName, f.RepoOwner, f.RepoName)
	case domain.Angular:
		generator = angular.NewGenerator(f.Version, f.ServiceName, f.RepoOwner, f.RepoName)
	case domain.Python:
		generator = python.NewGenerator(f.Version, f.ServiceName, f.RepoOwner, f.RepoName, f.GitToken)
	default:
		return nil, &domain.ErrUnsupportedLanguage{Language: language}
	}
	return generator, nil
}
