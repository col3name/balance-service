package infrastructure

import (
	"github.com/col3name/balance-transfer/pkg/common/app/errors"
	"github.com/col3name/balance-transfer/pkg/common/app/logger"
)

func InternalError(logger logger.Logger, err error) error {
	if err != nil {
		logger.Error(err)
	}
	return errors.ErrInternal
}
