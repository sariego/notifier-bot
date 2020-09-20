/* INSERT initial data here */

insert into "channel_info" ("channel_id", "name", "users")
values
  ('channelx', 'test channel  name', '{user1,user2}');

insert into "feedback" ("user_id", "channel_id", "tag", "content")
values
  ('user1', 'channel1', 'feedback', 'Lorem Ipsum Dolor Sit Amet Consectetur Adipisicing Elit Sed Do Eiusmod Tempor'),
  ('user2', 'channel2', 'feedback', 'Lorem Ipsum Dolor Sit'),
  ('user2', 'channelx', 'bug', 'Lorem Ipsum Dolor Sit Amet Consectetur Ad');
  
insert into "identity" ("username", "user_id", "channel_id")
values 
  ('test1', 'user1', 'channel1'),
  ('test2', 'user2', 'channel2');
