// Copyright (C) 2018 The Android Open Source Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apex

import (
	"fmt"
	"io"

	"android/soong/android"
	"github.com/google/blueprint/proptools"
)

var String = proptools.String

func init() {
	android.RegisterModuleType("apex_key", apexKeyFactory)
}

type apexKey struct {
	android.ModuleBase

	properties apexKeyProperties

	public_key_file  android.Path
	private_key_file android.Path

	keyName string
}

type apexKeyProperties struct {
	// Path to the public key file in avbpubkey format. Installed to the device.
	// Base name of the file is used as the ID for the key.
	Public_key *string
	// Path to the private key file in pem format. Used to sign APEXs.
	Private_key *string
}

func apexKeyFactory() android.Module {
	module := &apexKey{}
	module.AddProperties(&module.properties)
	android.InitAndroidModule(module)
	return module
}

func (m *apexKey) DepsMutator(ctx android.BottomUpMutatorContext) {
}

func (m *apexKey) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	m.public_key_file = android.PathForModuleSrc(ctx, String(m.properties.Public_key))
	m.private_key_file = android.PathForModuleSrc(ctx, String(m.properties.Private_key))

	pubKeyName := m.public_key_file.Base()[0 : len(m.public_key_file.Base())-len(m.public_key_file.Ext())]
	privKeyName := m.private_key_file.Base()[0 : len(m.private_key_file.Base())-len(m.private_key_file.Ext())]

	if pubKeyName != privKeyName {
		ctx.ModuleErrorf("public_key %q (keyname:%q) and private_key %q (keyname:%q) do not have same keyname",
			m.public_key_file.String(), pubKeyName, m.private_key_file, privKeyName)
		return
	}
	m.keyName = pubKeyName

	ctx.InstallFile(android.PathForModuleInstall(ctx, "etc/security/apex"), m.keyName, m.public_key_file)
}

func (m *apexKey) AndroidMk() android.AndroidMkData {
	return android.AndroidMkData{
		Class:      "ETC",
		OutputFile: android.OptionalPathForPath(m.public_key_file),
		Extra: []android.AndroidMkExtraFunc{
			func(w io.Writer, outputFile android.Path) {
				fmt.Fprintln(w, "LOCAL_MODULE_PATH :=", "$(TARGET_OUT)/etc/security/apex")
				fmt.Fprintln(w, "LOCAL_INSTALLED_MODULE_STEM :=", m.keyName)
			},
		},
	}
}
