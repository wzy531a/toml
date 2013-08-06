package toml

import (
	"fmt"
	gs "github.com/rafrombrc/gospec/src/gospec"
	"reflect"
	"time"
)

func CheckTypeSpec(c gs.Context) {
	var err error

	var tomlBlob = `
ranking = ["Springsteen", "J Geils"]

[bands.Springsteen]
type = "ignore_this"
started = 1973
albums = ["Greetings", "WIESS", "Born to Run", "Darkness"]
not_albums = ["Greetings", "WIESS", "Born to Run", "Darkness"]

[bands.J Geils]
started = 1970
albums = ["The J. Geils Band", "Full House", "Blow Your Face Out"]
`

	type classics struct {
		Ranking []string
		Bands   map[string]Primitive
	}

	c.Specify("check mapping", func() {
		// Do the initial decode. Reflection is delayed on Primitive values.
		var music classics
		var md MetaData
		md, err = Decode(tomlBlob, &music)
		c.Assume(err, gs.IsNil)
		fmt.Printf("md.mapping kind(): %s\n", reflect.TypeOf(md.mapping))
		// TODO: do the strict comparison in a separate function
		err = CheckType(md.mapping, music)
		c.Assume(err, gs.IsNil)
	})
}

func DecodeStrictSpec(c gs.Context) {

	// This blob when used with an empty_ignore blob
	var testSimple = `
age = 250
andrew = "gallant"
kait = "brady"
now = 1987-07-05T05:45:00Z 
yesOrNo = true
pi = 3.14
colors = [
	["red", "green", "blue"],
	["cyan", "magenta", "yellow", "black"],
]

[Annoying.Cats]
plato = "smelly"
cauchy = "stupido"
`

	var tomlBlob = `
# Some comments.
[alpha]
ip = "10.0.0.1"

	[alpha.config]
	Ports = [8001, 8002]
	Location = "Toronto"
	Created = 1987-07-05T05:45:00Z
`

	type serverConfig struct {
		Ports    []int
		Location string
		Created  time.Time
	}

	type server struct {
		IPAddress string       `toml:"ip"`
		Config    serverConfig `toml:"config"`
	}

	type kitties struct {
		Plato  string
		Cauchy string
	}

	type simple struct {
		Age      int
		Colors   [][]string
		Pi       float64
		YesOrNo  bool
		Now      time.Time
		Andrew   string
		Kait     string
		Annoying map[string]kitties
	}

	type servers map[string]server

	var config servers
	var val simple
	var err error
	var md MetaData

	md, err = Decode(tomlBlob, &config) //, empty_ignore)
	c.Assume(err, gs.IsNil)
	err = CheckType(md.mapping, config) // this should pass with no errors
	c.Assume(err, gs.IsNil)

	//empty_ignore := map[string]interface{}{}
	md, err = Decode(testSimple, &val)
	c.Assume(err, gs.IsNil)
	err = CheckType(md.mapping, val) // this should pass with no errors
	c.Assume(err, gs.IsNil)

	// TODO: convert this to use Decode and CheckType
	//_, err = DecodeStrict(testBadArg, &val, empty_ignore)
	//c.Assume(err.Error(), gs.Equals, "Configuration contains key [not_andrew] which doesn't exist in struct")

}
