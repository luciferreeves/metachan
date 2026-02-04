package env

func Parse(config any) error {
	elem, t, err := validateConfigInput(config)
	if err != nil {
		return err
	}

	for i := range elem.NumField() {
		field := elem.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		envKey := fieldType.Tag.Get("env")
		defaultVal := fieldType.Tag.Get("default")

		if envKey == "" {
			continue
		}

		setFieldFromEnv(field, envKey, defaultVal)
	}

	return nil
}
