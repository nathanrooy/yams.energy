-- schemas
CREATE SCHEMA IF NOT EXISTS events;
CREATE SCHEMA IF NOT EXISTS subscribers;

-- strava subscribers    
CREATE TABLE IF NOT EXISTS subscribers.strava (
    id            integer    NOT NULL PRIMARY KEY,
    access_token  text       NOT NULL,
    refresh_token text       NOT NULL,
    expires_at    integer    NOT NULL);

CREATE INDEX idx_subscribers_strava ON subscribers.strava (id);

-- events
CREATE TABLE IF NOT EXISTS events.sink (
    event_time 	  int8      NOT NULL,
    event         jsonb     NOT NULL);
