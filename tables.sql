CREATE TABLE categories (
	id INTEGER,
	language TEXT,
	parent_id INTEGER NOT NULL,
	image_url TEXT NOT NULL,
	title TEXT NOT NULL,
	last_updated BIGINT NOT NULL,
	PRIMARY KEY (id, language)
);

--CREATE TYPE state AS (
--	code TEXT,
--	name TEXT
--);

--CREATE TABLE countries (
--	code TEXT PRIMARY KEY,
--	name TEXT NOT NULL,
--	region TEXT NOT NULL,
--	states state[],
--	last_updated BIGINT NOT NULL
--);

CREATE TABLE countries (
	code TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	region TEXT NOT NULL,
	states JSONB,
	last_updated BIGINT NOT NULL
);

--CREATE TYPE mockup AS (
--	id INTEGER,
--	category_name TEXT,
--	view_name TEXT,
--	restricted_to_variants INTEGER[]
--);

--CREATE TYPE mockup_style AS (
--	placement TEXT,
--	display_name TEXT,
--	technique TEXT,
--	print_area_width REAL,
--	print_area_height REAL,
--	print_area_type TEXT,
--	dpi INTEGER,
--	mockup_styles mockup[]
--);

CREATE TABLE mockup_styles (
	_id SERIAL,
	product_id INTEGER PRIMARY KEY,
	mockup_styles mockup_style[],
	last_updated BIGINT NOT NULL
);

--CREATE TYPE mockup_template AS (
--	catalog_variant_ids INTEGER[],
--	placement TEXT,
--	technique TEXT,
--	image_url TEXT,
--	background_url TEXT,
--	background_color INTEGER,
--	template_width INTEGER ,
--	template_height INTEGER ,
--	print_area_width INTEGER ,
--	print_area_height INTEGER ,
--	print_area_top INTEGER ,
--	print_area_left INTEGER ,
--	template_positioning TEXT ,
--	orientation TEXT,
--	template_type TEXT,
--	role TEXT
--);

--CREATE TABLE mockup_templates (
--	_id SERIAL,
--	product_id INTEGER PRIMARY KEY,
--	mockup_templates mockup_template[],
--	last_updated BIGINT NOT NULL
--);

--CREATE TYPE color AS (
--	name TEXT,
--	value TEXT
--);

--CREATE TYPE technique AS (
--	key TEXT,
--	display_name TEXT,
--	is_default BOOLEAN
--);

--CREATE TYPE catalog_option AS (
--	name TEXT,
--	techniques TEXT[],
--	type TEXT,
--	values JSONB[]
--);

--CREATE TYPE file_layer AS (
--	type TEXT,
--	layer_options catalog_option[]
--);

--CREATE TYPE placement AS (
--	placement TEXT,
--	technique TEXT,
--	layers file_layer[],
--	placement_options catalog_option[],
--	conflicting_placements TEXT[]
--);

--CREATE TABLE products (
--	id INTEGER,
--	language TEXT,
--	main_category_id INTEGER NOT NULL,
--	type TEXT NOT NULL,
--	name TEXT NOT NULL,
--	brand TEXT,
--	model TEXT,
--	image TEXT NOT NULL,
--	variant_count INTEGER NOT NULL,
--	is_discontinued BOOLEAN NOT NULL,
--	description TEXT NOT NULL,
--	sizes TEXT[] NOT NULL,
--	colors color[] NOT NULL,
--	techniques technique[] NOT NULL,
--	placements placement[] NOT NULL,
--	last_updated BIGINT NOT NULL,
--	PRIMARY KEY (id, language)
--);

CREATE TABLE products (
	id INTEGER PRIMARY KEY,
	main_category_id INTEGER NOT NULL,
	type TEXT NOT NULL,
	name TEXT NOT NULL,
	brand TEXT,
	model TEXT,
	image TEXT NOT NULL,
	variant_count INTEGER NOT NULL,
	catalog_variant_ids INTEGER[] NOT NULL,
	is_discontinued BOOLEAN NOT NULL,
	description TEXT NOT NULL,
	sizes TEXT[] NOT NULL,
	colors JSONB NOT NULL,
	techniques JSONB NOT NULL,
	placements JSONB NOT NULL,
	product_options JSONB NOT NULL,
	last_updated BIGINT NOT NULL
);

CREATE TABLE product_translations (
	product_id INTEGER,
	language TEXT,
	name TEXT NOT NULL,
	description TEXT NOT NULL,
	last_updated BIGINT NOT NULL,
	PRIMARY KEY (product_id, language)
);


CREATE TYPE file_option_prices AS (
	name TEXT,
	type TEXT,
	values JSONB[],
	description TEXT,
	price JSONB
);

CREATE TYPE layer_option_prices AS (
	type TEXT,
	additional_price TEXT,
	layer_options JSONB[]
);

CREATE TYPE layer AS (
	type TEXT,
	additional_price TEXT,
	layer_options layer_option_prices[]
);

--CREATE TYPE additional_placement AS (
--	id TEXT,
--	title TEXT,
--	type TEXT,
--	technique_key TEXT,
--	price TEXT,
--	discounted_price TEXT,
--	placement_options file_option_prices[],
--	layers layer[]
--);

--CREATE TYPE variant_technique_price AS (
--	technique_key TEXT,
--	technique_display_name TEXT,
--	price TEXT,
--	discounted_price TEXT
--);

--CREATE TYPE variants_price_data AS (
--	id INTEGER,
--	techniques variant_technique_price[]
--);

CREATE TABLE products_prices (
	product_id INTEGER NOT NULL,
	currency TEXT NOT NULL,
	product_prices JSONB NOT NULL,
	last_updated BIGINT NOT NULL,
	PRIMARY KEY (product_id, currency)
);

CREATE TABLE mockup_templates (
	product_id INTEGER PRIMARY KEY,
	mockup_templates JSONB NOT NULL,
	last_updated BIGINT NOT NULL
);

CREATE TABLE mockup_styles (
	product_id INTEGER PRIMARY KEY,
	mockup_styles JSONB NOT NULL,
	last_updated BIGINT NOT NULL
);

CREATE TABLE variants (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	catalog_product_id INTEGER NOT NULL,
	color TEXT,
	color_code TEXT,
	color_code2 TEXT,
	image TEXT NOT NULL,
	size TEXT NOT NULL,
	availability JSONB NOT NULL,
	last_updated BIGINT NOT NULL
);
