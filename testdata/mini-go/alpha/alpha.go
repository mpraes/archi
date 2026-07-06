package alpha

import "github.com/example/mini-go/beta"

const SharedMagic = "shared-secret-value"

func Alpha() {
	beta.Beta()
}
