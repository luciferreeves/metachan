package meta

func (f facade) MustHave() required {
	return required{req: f.req, ctx: f.ctx}
}

func (f facade) Default(def string) withDefault {
	return withDefault{req: f.req, ctx: f.ctx, def: def}
}
