package shared

// ObjectSnapshot is a typed, immutable view used by transactional change
// logging. Domain snapshots implement this interface without entering a raw
// map or persistence representation at the application boundary.
type ObjectSnapshot interface {
	ObjectType() string
	CloneSnapshot() ObjectSnapshot
}
