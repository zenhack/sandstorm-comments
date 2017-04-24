
CREATE TABLE IF NOT EXISTS comments (
	id INTEGER PRIMARY KEY,
	article VARCHAR(255) NOT NULL,
	author VARCHAR(80) NOT NULL,
	body TEXT NOT NULL,
	needsMod BOOL NOT NULL
);

-- Misc. key/value store -- sometimes we have data that's not really
-- relational in nature (e.g. simple config options), but it's nice
-- to still piggy-back on SQLite for the ACID properties.
CREATE TABLE IF NOT EXISTS key_val (
	key TEXT UNIQUE NOT NULL,
	value TEXT NOT NULL
);
