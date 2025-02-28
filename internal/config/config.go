package config

import "defibotgo/internal/models"

// ChainToRpcUrlRead maps a Chain to its RPC URL for view functions
var ChainToRpcUrlRead = map[models.Chain]string{
	models.Optimism: GetSecret(RpcNodeReadKey),
}

// ChainToRpcUrlWrite maps a Chain to its RPC URL for write functions
var ChainToRpcUrlWrite = map[models.Chain]string{
	models.Optimism: GetSecret(RpcNodeWriteKey),
}
