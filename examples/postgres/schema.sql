CREATE TABLE author (
  id BIGINT NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL
);

CREATE TABLE author_profile (
  id BIGINT NOT NULL PRIMARY KEY,
  author_id BIGINT NOT NULL UNIQUE,
  bio TEXT,
  avatar_url VARCHAR(255),
  created_at TIMESTAMP NOT NULL,
  FOREIGN KEY (author_id) REFERENCES author(id)
);

CREATE TABLE video (
  id BIGINT NOT NULL PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  slug VARCHAR(255) NOT NULL,
  author_id BIGINT NOT NULL,
  description TEXT,
  created_at TIMESTAMP NOT NULL,
  UNIQUE (slug),
  FOREIGN KEY (author_id) REFERENCES author(id)
);

CREATE TABLE comment (
  id BIGINT NOT NULL PRIMARY KEY,
  video_id BIGINT NOT NULL,
  author_id BIGINT NOT NULL,
  body TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  FOREIGN KEY (video_id) REFERENCES video(id),
  FOREIGN KEY (author_id) REFERENCES author(id)
);

CREATE TABLE tag (
  id BIGINT NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE video_tag (
  video_id BIGINT NOT NULL,
  tag_id BIGINT NOT NULL,
  PRIMARY KEY (video_id, tag_id),
  FOREIGN KEY (video_id) REFERENCES video(id),
  FOREIGN KEY (tag_id) REFERENCES tag(id)
);
