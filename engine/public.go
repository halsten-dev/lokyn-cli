package engine

// This file only contains the public API of the engine package.

// DiscoverProject is the function that allows to explore the current project root,
// find the .lokyn folder if existing and load the project.
func DiscoverProject() (Project, error) {
	var project Project

	err := discoverProject(&project, "./")

	return project, err
}

// DiscoverKeys is the function that allows to explore the current project root,
// find every go files and analyse them to find all the Lokyn keys.
func DiscoverKeys() (Keys, error) {
	var keys Keys

	err := discoverDirectory(&keys, "./")

	return keys, err
}
