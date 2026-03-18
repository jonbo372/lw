package namegen

import (
	"math/rand"
	"time"
)

var adjectives = []string{
	"blazing", "cosmic", "dancing", "electric", "fuzzy", "gentle", "happy", "jolly", "keen", "lazy",
	"mighty", "noble", "odd", "peaceful", "quirky", "rapid", "silent", "turbo", "vivid", "witty",
	"amber", "bronze", "crimson", "dusty", "emerald", "frosty", "golden", "hollow", "icy", "jade",
}

var nouns = []string{
	"badger", "cobra", "dolphin", "eagle", "falcon", "gecko", "hawk", "iguana", "jackal", "koala",
	"lemur", "moose", "narwhal", "otter", "parrot", "quail", "raven", "squid", "toucan", "urchin",
	"anvil", "beacon", "comet", "dagger", "ember", "flute", "geyser", "harpoon", "iris", "jetpack",
}

// Generate returns a random "adjective_noun" name.
func Generate() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return adjectives[r.Intn(len(adjectives))] + "_" + nouns[r.Intn(len(nouns))]
}
