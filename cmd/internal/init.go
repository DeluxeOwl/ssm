package internal

func init() {
	if ssmPath, err := getModulePath(ssmModulePath, ssmModuleVersion); err == nil {
		SSMStates, _ = LoadStates(ssmPath)
	}
}
