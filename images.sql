CREATE TABLE images (
	id SERIAL PRIMARY KEY,
	filename TEXT NOT NULL,
	image BYTEA,
	created TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX filename_idx ON images (filename);
