/*
 * Copyright (c) Facebook, Inc. and its affiliates.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 */

package obsidian

import (
	"net/http"

	"magma/orc8r/cloud/go/obsidian/handlers"
	"magma/orc8r/cloud/go/orc8r"
	"magma/orc8r/cloud/go/services/configurator"
	"magma/orc8r/cloud/go/storage"

	"github.com/labstack/echo"
)

// Gateway is a special type of network entity. Since magmad supports multiple
// config types per gateway, each config is modeled as an entity with an
// association to the gateway entity. If the configType is "magmad_gateway" the
// config will simply be stored inside the gateway entity

func configuratorCreateGatewayConfig(c echo.Context, networkID string, configType string, configKey string, iConfig interface{}) error {
	// If the type is magmad_gateway add the config to the gw entity
	if configType == orc8r.MagmadGatewayType {
		return configuratorCreateEntityConfig(c, networkID, configType, configKey, iConfig)
	}

	config, nerr := GetConfigAndValidate(c, iConfig)
	if nerr != nil {
		return nerr
	}

	// the migrated handler will create the entity and associate the magmad
	// access gateway to it
	// note that this operation is not atomic, so there is a very slim but
	// nonzero chance that the entity is created without the proper assoc
	_, err := configurator.CreateEntity(networkID, configurator.NetworkEntity{
		Type:   configType,
		Key:    configKey,
		Config: config,
	})
	if err != nil {
		return handlers.HttpError(err, http.StatusInternalServerError)
	}

	_, err = configurator.UpdateEntity(networkID, configurator.EntityUpdateCriteria{
		// hardcoded type to prevent import cycle
		Type:              orc8r.MagmadGatewayType,
		Key:               configKey,
		AssociationsToAdd: []storage.TypeAndKey{{Type: configType, Key: configKey}},
	})
	if err != nil {
		return handlers.HttpError(err, http.StatusInternalServerError)
	}
	return c.JSON(http.StatusCreated, configKey)
}

func configuratorDeleteGatewayConfig(c echo.Context, networkID string, configType string, configKey string) error {
	// If the type is magmad_gateway remove config from the gateway entity
	if configType == orc8r.MagmadGatewayType {
		return configuratorDeleteEntityConfig(c, networkID, configType, configKey)
	}

	err := configurator.DeleteEntity(networkID, configType, configKey)
	if err != nil {
		return handlers.HttpError(err, http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}
