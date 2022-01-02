package util

type Configuration struct {
	ID uint

	// Raw mesh key. Used to authenticate mesh messages & derive all other keys.
	MeshKey string

	// Derived symmetric key for ChaCha20-Poly1305 operations.
	ChaChaKey []byte `gorm:"-"`
}
