package filepath

import "path/filepath"

func Resolve(makeAbs bool, paths ...string) (resolved string) {

	for _, path := range paths {
		if filepath.IsAbs(path) {
			resolved = path
		} else {
			resolved = filepath.Join(resolved, path)
		}
	}

	if makeAbs {
		var err error
		resolved, err = filepath.Abs(resolved)

		if err != nil {
			panic(err)
		}
	} else {
		resolved = filepath.Clean(resolved)
	}

	return resolved
}
