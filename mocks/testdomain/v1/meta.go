package testdomain

import (
//"github.com/ugorji/go/codec"
//. "github.com/xzeus/cqrs"
)

//var domain = MustCompile(&Meta{})

//type __ struct{ MessageBase }

//func (_ __) Domain() Domain { return domain }

//func (__ DomainMessage) Clone() MessageDefiner { return __.Domain().MessageTypes().ByInstance(__) } //.Commands().All()[0] } // ByType(__)

/*
https://github.com/ugorji/go/tree/master/codec

func (__ __) Encode() error {
	codec.
}
*/

//type MessageBase struct {
//	DomainLinker
//}

// TODO: return message definer for this instance
//func (__ *MessageBase) Def() Message { return __.Domain().Commands()[0] } //(ByType(__)) }

// Opts into execution scope bounded context registration
//var domain = BoundedContext.Register(&Meta{})

//var domain = MustCompile(&Meta{})
/*
var domain Domain

type __ struct{ MessageBase }

func (_ __) Domain() Domain { return domain }

type Meta struct {
	Aggregate *TestDomain
	Commands  struct {
		_ *SetEmpty `id:"1000"`
		_ *SetValue `id:"2000"`
	}
	Events struct {
		_ *ValueChanged      `id:"1000"`
		_ *ValueChangeFailed `id:"1010"`
	}
}
*/

/* Manually populated DomainDef example

var uri = "github.com/xzeus/cqrs/mocks/testdomain/v1"
var domain2 = &DomainDef{
	Uri: uri,
	Id: crypto.CrcHash64([]byte(uri)),
	Name: "testdomain",
	Version: 1,
	Proto: func() Aggregate {
		return &TestDomain{}
	},
	CommandDefs: []CommandTypeDef {
		CommandTypeDef{
			Domain: domain2,
			MessageType: MakeVersionedCommandType(1, 1000),
			Name: "SetEmpty",
			Proto: func() Command {
				return &SetEmpty{}
			},
		},
		CommandTypeDef{
			Domain: domain2,
			MessageType: MakeVersionedCommandType(1, 2000),
			Name: "SetValue",
			Proto: func() Command {
				return &SetValue{}
			},
		},
	},
}
*/
