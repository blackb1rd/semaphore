alter table `role` add column `project_id` int;
alter table `role` add foreign key (`project_id`) references `project` (`id`) on delete cascade;
