package utils

// generate unique code
// Limited to uint8 for now
// Easier to work encode in byte code idk
type Code_Generator struct {
	Start_ID uint8
}

// Just increments the default value
func (idGen *Code_Generator) NewCode() uint8 {
	ret_id := idGen.Start_ID
	idGen.Start_ID += 1

	return ret_id
}
