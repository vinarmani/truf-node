package utils

import "os"

// Warning: Used environment variables are not encrypted in the CloudFormation template,
// nor to who have access to the instance if it used on a service configuration file.
// Switch for encryption if necessary.
func GetEnvStringsForService(envDict map[string]string) string {
	envStrings := ""
	for k, v := range envDict {
		envStrings += "Environment=\"" + k + "=" + v + "\"\n"
	}
	return envStrings
}

// GetEnvVars returns a map of environment variables from the given list of
// environment variable names. If an environment variable is not set, it will
// be an empty string in the map.
func GetEnvVars(envNames ...string) map[string]string {
	envDict := make(map[string]string)
	for _, envName := range envNames {
		envDict[envName] = os.Getenv(envName)
	}
	return envDict
}
