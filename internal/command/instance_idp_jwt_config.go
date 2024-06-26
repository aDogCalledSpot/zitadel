package command

import (
	"context"

	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/zerrors"
)

func (c *Commands) ChangeDefaultIDPJWTConfig(ctx context.Context, config *domain.JWTIDPConfig) (*domain.JWTIDPConfig, error) {
	if config.IDPConfigID == "" {
		return nil, zerrors.ThrowInvalidArgument(nil, "INSTANCE-m9322", "Errors.IDMissing")
	}
	existingConfig := NewInstanceIDPJWTConfigWriteModel(ctx, config.IDPConfigID)
	err := c.eventstore.FilterToQueryReducer(ctx, existingConfig)
	if err != nil {
		return nil, err
	}

	if existingConfig.State == domain.IDPConfigStateRemoved || existingConfig.State == domain.IDPConfigStateUnspecified {
		return nil, zerrors.ThrowNotFound(nil, "INSTANCE-2m00d", "Errors.IAM.IDPConfig.AlreadyExists")
	}

	instanceAgg := InstanceAggregateFromWriteModel(&existingConfig.WriteModel)
	changedEvent, hasChanged, err := existingConfig.NewChangedEvent(
		ctx,
		instanceAgg,
		config.IDPConfigID,
		config.JWTEndpoint,
		config.Issuer,
		config.KeysEndpoint,
		config.HeaderName)
	if err != nil {
		return nil, err
	}
	if !hasChanged {
		return nil, zerrors.ThrowPreconditionFailed(nil, "INSTANCE-3n9gg", "Errors.IAM.IDPConfig.NotChanged")
	}

	pushedEvents, err := c.eventstore.Push(ctx, changedEvent)
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(existingConfig, pushedEvents...)
	if err != nil {
		return nil, err
	}
	return writeModelToIDPJWTConfig(&existingConfig.JWTConfigWriteModel), nil
}
