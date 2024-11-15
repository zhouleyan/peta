/*
 *  This file is part of PETA.
 *  Copyright (C) 2024 The PETA Authors.
 *  PETA is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  PETA is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with PETA. If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"log"
	"os"
	"os/exec"
	"peta.io/peta/pkg/apis"
	configv1alpha2 "peta.io/peta/pkg/apis/config/v1alpha2"
	"peta.io/peta/pkg/apis/healthz"
	"peta.io/peta/pkg/apis/version"
	urlruntime "peta.io/peta/pkg/runtime"
)

var output string

func init() {
	flag.StringVar(&output, "output", "./api/peta-openapi-spec/swagger.json", "--output=./api.json")
}

func main() {
	flag.Parse()
	if err := validateSpec(generateSwaggerJSON()); err != nil {
		klog.Warningf("Swagger specification validation failed: %v", err)
	}
}

func validateSpec(apiSpec []byte) error {
	swaggerDoc, err := loads.Analyzed(apiSpec, "")
	if err != nil {
		return err
	}

	// Attempts to report about all errors
	validate.SetContinueOnErrors(false)

	v := validate.NewSpecValidator(swaggerDoc.Schema(), strfmt.Default)
	result, _ := v.Validate(swaggerDoc)

	if result.HasWarnings() {
		log.Printf("See warnings below:\n")
		for _, warning := range result.Warnings {
			log.Printf("- WARNING: %s\n", warning.Error())
		}
	}

	if result.HasErrors() {
		str := fmt.Sprintf("The swagger spec is invalid against swagger specification %s.\nSee errors below:\n", swaggerDoc.Version())
		for _, desc := range result.Errors {
			str += fmt.Sprintf("- %s\n", desc.Error())
		}
		log.Println(str)
		return errors.New(str)
	}

	return nil
}

func generateSwaggerJSON() []byte {
	container := apis.Container

	handlers := []apis.Handler{
		version.NewFakeHandler(),
		healthz.NewFakeHandler(),
		configv1alpha2.NewFakeHandler(),
	}

	for _, h := range handlers {
		urlruntime.Must(h.AddToContainer(container))
	}

	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}

	data, _ := json.MarshalIndent(restfulspec.BuildSwagger(config), "", "  ")
	if err := os.WriteFile(output, data, 0644); err != nil {
		log.Fatal(err)
	}
	log.Printf("successfully written to %s", output)
	return data
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "PETA API",
			Description: "PETA OpenAPI",
			Version:     gitVersion(),
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "PETA",
					URL:   "https://peta.io",
					Email: "support@peta.io",
				},
			},
		},
	}

	// setup security definitions
	swo.SecurityDefinitions = map[string]*spec.SecurityScheme{
		"BearerToken": {SecuritySchemeProps: spec.SecuritySchemeProps{
			Type:        "apiKey",
			Name:        "Authorization",
			In:          "header",
			Description: "Bearer Token Authentication",
		}},
	}
	swo.Security = []map[string][]string{{"BearerToken": []string{}}}

	swo.Tags = []spec.Tag{
		{
			TagProps: spec.TagProps{
				Name: apis.TagConfigurations,
			},
		},
		{
			TagProps: spec.TagProps{
				Name: apis.TagNonResourceAPI,
			},
		},
	}
}

func gitVersion() string {
	out, err := exec.Command("sh", "-c", "git tag --sort=committerdate | tail -1 | tr -d '\n'").Output()
	if err != nil {
		log.Printf("failed to get git version: %s", err)
		return "v0.0.0"
	}
	return string(out)
}
