package controller

import (
	"fmt"
)

func (crtl *Controller) Standup() error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto;`,

		`CREATE SCHEMA IF NOT EXISTS oauth2_server;`,

		// `DROP TABLE IF EXISTS oauth2_server.applications;`,
		`CREATE TABLE IF NOT EXISTS oauth2_server.applications (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			website TEXT NOT NULL,
			redirect_url TEXT NOT NULL,
			client_id TEXT NOT NULL,
			client_secret TEXT NOT NULL
		);`,
		// `INSERT INTO oauth2_server.applications (name, website, redirect_url, client_id, client_secret)
		// VALUES (
		// 	'demo',
		// 	'http://demo.com',
		// 	'http://demo.com/auth_redirect',
		// 	'CLIENT_ID_DEMO',
		// 	crypt('CLIENT_SECRET_DEMO', gen_salt('bf'))
		//   ) ON CONFLICT DO NOTHING;`,

		// `DROP TABLE IF EXISTS oauth2_server.users;`,
		`CREATE TABLE IF NOT EXISTS oauth2_server.users (
			id SERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			roles TEXT[] NOT NULL,
			application_id INT NOT NULL
		);`,
		`ALTER TABLE oauth2_server.users DROP CONSTRAINT IF EXISTS fk_users_applications;`,
		`ALTER TABLE oauth2_server.users ADD CONSTRAINT fk_users_applications 
		FOREIGN KEY (application_id) REFERENCES oauth2_server.applications (id);`,
		// `INSERT INTO oauth2_server.users (username, password, roles, application_id) VALUES (
		// 	'johndoe',
		// 	crypt('johnspassword', gen_salt('bf')),
		// 	'{read, write}'::TEXT[],
		// 	1
		//   ) ON CONFLICT DO NOTHING;`,
	}

	for _, q := range queries {
		_, err := crtl.db.Exec(q)
		if err != nil {
			return err
		}
	}

	fmt.Println("standup complete")

	return nil
}
