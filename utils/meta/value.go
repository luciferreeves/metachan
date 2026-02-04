package meta

func (f facade) MustHave() required {
	return required{req: f.req}
}

func (f facade) Default(def string) withDefault {
	return withDefault{req: f.req, def: def}
}
