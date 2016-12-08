/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"fmt"

	"github.com/cloudflare/cfssl/certdb"
	"github.com/hyperledger/fabric-cop/cli/server/dbutil"
	"github.com/hyperledger/fabric-cop/cli/server/ldap"
	"github.com/hyperledger/fabric-cop/cli/server/spi"
	"github.com/jmoiron/sqlx"
)

var userRegistry spi.UserRegistry
var certDBAccessor *CertDBAccessor

// InitUserRegistry is the factory method for the user registry.
// If LDAP is configured, then LDAP is used for the user registry;
// otherwise, the CFSSL DB which is used for the certificates table is used.
func InitUserRegistry(cfg *Config) error {

	var err error

	if cfg.LDAP != nil {
		// LDAP is being used for the user registry
		userRegistry, err = ldap.NewClient(cfg.LDAP)
		if err != nil {
			return err
		}
	} else {
		// The database is being used for the user registry
		var exists bool

		switch cfg.DBdriver {
		case "sqlite3":
			db, exists, err = dbutil.NewUserRegistrySQLLite3(cfg.DataSource)
			if err != nil {
				return err
			}

		case "postgres":
			db, exists, err = dbutil.NewUserRegistryPostgres(cfg.DataSource, &cfg.TLSConf.DBClient)
			if err != nil {
				return err
			}

		case "mysql":
			db, exists, err = dbutil.NewUserRegistryMySQL(cfg.DataSource, &cfg.TLSConf.DBClient)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("invalid 'DBDriver' in config file: %s", cfg.DBdriver)
		}

		dbAccessor := new(Accessor)
		dbAccessor.SetDB(db)

		userRegistry = dbAccessor

		// If the DB doesn't exist, bootstrap the DB
		if !exists {
			err := bootstrapDB()
			if err != nil {
				return err
			}
		}

	}

	return nil

}

// CertificateAccessor extends CFSSL database APIs for Certificates table
func CertificateAccessor(db *sqlx.DB) certdb.Accessor {
	certDBAccess := NewCertDBAccessor(db)
	certDBAccessor = certDBAccess
	return certDBAccess
}