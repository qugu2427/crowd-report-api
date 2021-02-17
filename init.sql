CREATE TABLE articles (
    id BIGSERIAL NOT NULL PRIMARY KEY,
    author VARCHAR(50) NOT NULL,
    author_google_id VARCHAR(25) NOT NULL,
    image_url VARCHAR(75) NOT NULL,
    title VARCHAR(200) NOT NULL,
    body VARCHAR(10000) NOT NULL,
    tags VARCHAR(200) NOT NULL,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
    vector tsvector
);

CREATE TABLE tags (
    tag VARCHAR(25) NOT NULL PRIMARY KEY
);

INSERT INTO tags (tag) VALUES ('science');
INSERT INTO tags (tag) VALUES ('sports');
INSERT INTO tags (tag) VALUES ('entertainment');
INSERT INTO tags (tag) VALUES ('education');
INSERT INTO tags (tag) VALUES ('politics');
INSERT INTO tags (tag) VALUES ('opinion');
INSERT INTO tags (tag) VALUES ('buisness');