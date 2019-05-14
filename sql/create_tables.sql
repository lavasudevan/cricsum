create table if not exists player (
id integer primary key autoincrement,
name text not null ,
team text
);

create table if not exists game (
  date text,
  id integer primary key autoincrement,
  innings1_id integer ,
  innings2_id integer 
);

create table if not exists 
innings_index (
  type text primary key,
  id integer 
);

create table if not exists 
innings (
  id integer,
  player_id integer,
  runs_scored integer,
  how_out text,
  fielder_id integer,
  bowler_id integer,
  primary key(id, player_id)
);


create table if not exists 
bowl_innings (
  id int,
  player_id integer,
  overs_bowled real,
  maiden integer,
  runs integer,
  wickets integer,
  primary key(id, player_id)
   
);
