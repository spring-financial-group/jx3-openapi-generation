package domain

type UVClient interface {
	// Generate pyproject.toml file required for the build and publish steps
	GeneratePyProjectFile(dir, packageName, packageVersion string) error

	// Build the python project in the given directory
	BuildProject(dir string) error

	// Publish the python project in the given directory to the specified index (default: pyx)
	PublishProject(dir string, indexName string) error
}
