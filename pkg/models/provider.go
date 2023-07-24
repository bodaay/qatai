package models

import "io"

type Provider interface {
	GetInfo(uuid string, host string, port int16, useSSL bool, skipVerify bool) (*ModelEndPoint, error)
	Generate(prompt string, model *ModelEndPoint, params *ModelParameters) (string, error)
	GenerateStream(prompt string, model *ModelEndPoint, params *ModelParameters) (io.ReadCloser, error)
}
