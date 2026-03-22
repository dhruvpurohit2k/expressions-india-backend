CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS enquiry(
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	name VARCHAR(50) NOT NULL,
	designation VARCHAR(100) NOT NULL,
	email_id VARCHAR(225) NOT NULL UNIQUE,
	contact_number TEXT NOT NULL,
	body TEXT NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- CREATE TABLE IF NOT EXISTS activity(
-- 	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
-- 	title VARCHAR(255) NOT NULL,
-- 	start_date DATE NOT NULL,
-- 	end_date DATE,
-- 	is_active BOOLEAN DEFAULT true,
-- 	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
-- 	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );

CREATE TABLE IF NOT EXISTS event(
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	title VARCHAR(255) NOT NULL,
	description TEXT,
	start_date DATE NOT NULL,
	end_date DATE,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workshop_type(
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name VARCHAR(100) NOT NULL UNIQUE
);

INSERT INTO workshop_type (name) VALUES 
	('Half Day Certificate Programs'),
	('National Summit And Conference'),
	('Online And Offline Pysychology Internship'),
	('School Students Events And Competitions'),
	('Webinar Based Orientation Course')
	ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS workshop(
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	title VARCHAR(300) NOT NULL,
	start_date DATE NOT NULL,
	end_date DATE,
	description TEXT,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	workshop_type INT REFERENCES workshop_type(id) ON DELETE CASCADE,
	registration_link TEXT
);

CREATE TABLE IF NOT EXISTS user_type(
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name VARCHAR (30) NOT NULL
);

INSERT INTO user_type (name) VALUES 
	('STUDENT'),
	('TEACHER'),
	('COUNCELLOR'),
	('PARENT'),
	('HEAD OF SCHOOL')
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS user_workshop(
	user_type_id INT REFERENCES user_type(id) ON DELETE CASCADE,
	workshop_id UUID REFERENCES workshop(id) ON DELETE CASCADE,
	PRIMARY KEY (user_type_id,workshop_id)
);


CREATE TABLE IF NOT EXISTS journal(
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	title TEXT NOT NULL,
	start_date DATE NOT NULL,
	end_date DATE NOT NULL,
	volume_number INT NOT NULL,
	issue_number INT NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS journal_chapter(
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	title TEXT NOT NULL,
	journal_id UUID REFERENCES journal(id) ON DELETE CASCADE,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS author(
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS journal_chapter_author(
	journal_chapter_id UUID REFERENCES journal_chapter(id) ON DELETE CASCADE,
	author_id INT REFERENCES author(id) ON DELETE CASCADE,
	PRIMARY KEY (journal_chapter_id,author_id)
);

CREATE TABLE IF NOT EXISTS media(
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	title TEXT,
	media_type VARCHAR(20) NOT NULL,
	s3_key VARCHAR(255) NOT NULL,
	description TEXT,
	url TEXT NOT NULL UNIQUE,
	thumbnail_url TEXT,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS album(
	id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	name TEXT NOT NULL UNIQUE,
	description TEXT,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS album_media(
	album_id UUID REFERENCES album(id) ON DELETE CASCADE,
	media_id UUID REFERENCES media(id) ON DELETE CASCADE,
	PRIMARY KEY (album_id,media_id)
);

-- CREATE TABLE IF NOT EXISTS activity_media (
--     activity_id UUID REFERENCES activity(id) ON DELETE CASCADE,
--     media_id UUID REFERENCES media(id) ON DELETE CASCADE,
--     PRIMARY KEY (activity_id, media_id)
-- );

CREATE TABLE IF NOT EXISTS journal_media (
    journal_id UUID REFERENCES journal(id) ON DELETE CASCADE,
    media_id UUID REFERENCES media(id) ON DELETE CASCADE,
    PRIMARY KEY (journal_id, media_id)
);

CREATE TABLE IF NOT EXISTS journal_chapter_media (
    journal_chapter_id UUID REFERENCES journal_chapter(id) ON DELETE CASCADE,
    media_id UUID REFERENCES media(id) ON DELETE CASCADE,
    PRIMARY KEY (journal_chapter_id, media_id)
);

CREATE TABLE IF NOT EXISTS workshop_media(
    workshop_id UUID REFERENCES workshop(id) ON DELETE CASCADE,
    media_id UUID REFERENCES media(id) ON DELETE CASCADE,
    PRIMARY KEY (workshop_id, media_id)
);

CREATE TABLE IF NOT EXISTS event_media(
	event_id UUID REFERENCES event(id) ON DELETE CASCADE,
	media_id UUID REFERENCES media(id) ON DELETE CASCADE,
	PRIMARY KEY (event_id,media_id)
);


--Trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$

BEGIN
	NEW.updated_at = CURRENT_TIMESTAMP;
	RETURN NEW;
END;

$$ language 'plpgsql';


-- TRIGGERS FOR updated_at
CREATE TRIGGER update_workshop_updated_at 
	BEFORE UPDATE ON workshop
	FOR EACH ROW
	EXECUTE FUNCTION update_updated_at_column();

-- CREATE TRIGGER update_activity_updated_at 
-- 	BEFORE UPDATE ON activity
-- 	FOR EACH ROW
-- 	EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_journal_updated_at
	BEFORE UPDATE ON journal
	FOR EACH ROW 
	EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_journal_chapter_updated_at
	BEFORE UPDATE ON journal_chapter
	FOR EACH ROW 
	EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_event_updated_at
	BEFORE UPDATE ON event
	FOR EACH ROW
	EXECUTE FUNCTION update_updated_at_column();
