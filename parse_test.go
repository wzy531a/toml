package toml

import (
	"testing"
)

var testParseSmall = `
# This is a TOML document. Boom.

wat = "chipper"

[owner.andrew.gallant]
hmm = "hi"

[owner] # Whoa there.
andreW = "gallant # poopy" # weeeee
predicate = false
num = -5192
f = -0.5192
zulu = 1979-05-27T07:32:00Z
whoop = "poop"
rawstring = 'this is \d+ and \user'
string = "C:\\user\\balabla\\"
tests = [ [1, 2, 3], ["abc", "xyz"] ]
arrs = [ # hmm
		 # more comments are awesome.
	1987-07-05T05:45:00Z,
	# say wat?
	1987-07-05T05:45:00Z,
	1987-07-05T05:45:00Z,
	# sweetness
] # more comments
# hehe
`

var testParseSmall2 = `
[a]
better = 43

[a.b.c]
answer = 42
`

func TestParse(t *testing.T) {
	_, err := parse(testParseSmall)
	if err != nil {
		t.Fatal(err)
	}
}
