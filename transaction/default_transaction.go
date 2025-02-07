/*
 * Copyright (c) 2022, AcmeStack
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package transaction

import (
	"context"
	"database/sql"
	"github.com/acmestack/gobatis/common"
	"github.com/acmestack/gobatis/connection"
	"github.com/acmestack/gobatis/datasource"
	"github.com/acmestack/gobatis/errors"
	"github.com/acmestack/gobatis/reflection"
	"github.com/acmestack/gobatis/statement"
	"github.com/acmestack/gobatis/util"
)

type DefaultTransaction struct {
	ds datasource.DataSource
	db *sql.DB
	tx *sql.Tx
}

func NewDefaultTransaction(ds datasource.DataSource, db *sql.DB) *DefaultTransaction {
	ret := &DefaultTransaction{ds: ds, db: db}
	return ret
}

func (trans *DefaultTransaction) GetConnection() connection.Connection {
	if trans.tx == nil {
		return (*connection.DefaultConnection)(trans.db)
	} else {
		return &TransactionConnection{tx: trans.tx}
	}
}

func (trans *DefaultTransaction) Close() {

}

func (trans *DefaultTransaction) Begin() error {
	tx, err := trans.db.Begin()
	if err != nil {
		return err
	}
	trans.tx = tx
	return nil
}

func (trans *DefaultTransaction) Commit() error {
	if trans.tx == nil {
		return errors.TRANSACTION_WITHOUT_BEGIN
	}

	err := trans.tx.Commit()
	if err != nil {
		return errors.TRANSACTION_COMMIT_ERROR
	}
	return nil
}

func (trans *DefaultTransaction) Rollback() error {
	if trans.tx == nil {
		return errors.TRANSACTION_WITHOUT_BEGIN
	}

	err := trans.tx.Rollback()
	if err != nil {
		return errors.TRANSACTION_COMMIT_ERROR
	}
	return nil
}

type TransactionConnection struct {
	tx *sql.Tx
}

type TransactionStatement struct {
	tx  *sql.Tx
	sql string
}

func (c *TransactionConnection) Prepare(sqlStr string) (statement.Statement, error) {
	ret := &TransactionStatement{
		tx:  c.tx,
		sql: sqlStr,
	}
	return ret, nil
}

func (c *TransactionConnection) Query(ctx context.Context, result reflection.Object, sqlStr string, params ...interface{}) error {
	db := c.tx
	rows, err := db.QueryContext(ctx, sqlStr, params...)
	if err != nil {
		return errors.STATEMENT_QUERY_ERROR
	}
	defer rows.Close()

	util.ScanRows(rows, result)
	return nil
}

func (c *TransactionConnection) Exec(ctx context.Context, sqlStr string, params ...interface{}) (common.Result, error) {
	db := c.tx
	return db.ExecContext(ctx, sqlStr, params...)
}

func (s *TransactionStatement) Query(ctx context.Context, result reflection.Object, params ...interface{}) error {
	rows, err := s.tx.QueryContext(ctx, s.sql, params...)
	if err != nil {
		return errors.STATEMENT_QUERY_ERROR
	}
	defer rows.Close()

	util.ScanRows(rows, result)
	return nil
}

func (s *TransactionStatement) Exec(ctx context.Context, params ...interface{}) (common.Result, error) {
	return s.tx.ExecContext(ctx, s.sql, params...)
}

func (s *TransactionStatement) Close() {
	//Will be closed when commit or rollback
}
