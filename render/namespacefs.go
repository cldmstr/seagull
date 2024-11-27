package render

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// NamespaceFS is an implementation of the `fs.FS`, `fs.GlobFS` and `fs.ReadFileFS` interfaces that makes a map of
// namespaced filesystems look as if they were a single filesystem. This is used to keep the templates from different
// sub-paths separate, while still being able to access them from a single component.
type NamespaceFS map[string]fs.FS

var (
	_ fs.FS         = &NamespaceFS{}
	_ fs.GlobFS     = &NamespaceFS{}
	_ fs.ReadFileFS = &NamespaceFS{}
)

// AddNamespace adds the subdir of the passed files under the namespace. Subdir must exist.
func (t NamespaceFS) AddNamespace(namespace, subdir string, files fs.FS) error {
	sub, err := fs.Sub(files, subdir)
	if err != nil {
		return fmt.Errorf("add filesystem %q: %w", namespace, err)
	}
	t[namespace] = sub

	return nil
}

// ReadFile implements the fs.ReadFileFS interface for the namespaced subdirectories.
func (t NamespaceFS) ReadFile(path string) ([]byte, error) {
	ns, internalName, err := t.internalName(path)
	if err != nil {
		return nil, fmt.Errorf("get namespace from path %q: %w", path, err)
	}

	fsys, err := t.getFS(ns)
	if err != nil {
		return nil, fmt.Errorf("get sub-filesystem for namespace %q: %w", ns, err)
	}

	f, err := fs.ReadFile(fsys, internalName)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", path, err)
	}

	return f, nil
}

// Open implements the fs.FS interface for the namespaced subdirectories.
func (t NamespaceFS) Open(path string) (fs.File, error) {
	ns, internalName, err := t.internalName(path)
	if err != nil {
		return nil, fmt.Errorf("get namespace from path %q: %w", path, err)
	}

	fsys, err := t.getFS(ns)
	if err != nil {
		return nil, fmt.Errorf("get sub-filesystem for namespace %q: %w", path, err)
	}

	f, err := fsys.Open(internalName)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", path, err)
	}

	return f, nil
}

// Glob implements the fs.GlobFS interface for the namespaced subdirectories.
func (t NamespaceFS) Glob(pattern string) ([]string, error) {
	ns, internalPattern, err := t.internalName(pattern)
	if err != nil {
		return nil, fmt.Errorf("split namespace from pattern %q: %w", pattern, err)
	}

	fsys, err := t.getFS(ns)
	if err != nil {
		return nil, fmt.Errorf("get sub-filesystem for namespace %q: %w", ns, err)
	}

	matches, err := fs.Glob(fsys, internalPattern)
	if err != nil {
		return nil, fmt.Errorf("call Glob on sub-filesystem for namespace %q with pattern %q: %w", ns, internalPattern, err)
	}

	paths := make([]string, 0, len(matches))
	for _, n := range matches {
		paths = append(paths, filepath.Join(ns, n))
	}

	return paths, nil
}

// internalName splits the passed name into the namespace before the first / and the rest of the file path
func (t NamespaceFS) internalName(name string) (namespace string, internalName string, err error) {
	nameSegments := strings.Split(name, string(filepath.Separator))
	if len(nameSegments) < 2 {
		return "", "", fmt.Errorf("path %q requires at least two levels: namespace/filepath: %w", name, err)
	}
	namespace = nameSegments[0]
	internalName = filepath.Join(nameSegments[1:]...)

	return namespace, internalName, nil
}

// getFS retrieves the actual filesystem for the namespace
func (t NamespaceFS) getFS(namespace string) (fs.FS, error) {
	if fsys, ok := t[namespace]; ok {
		return fsys, nil
	}

	return nil, fmt.Errorf("namespace %q does not exist", namespace)
}
