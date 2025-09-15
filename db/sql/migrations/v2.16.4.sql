alter table `task__output` add `stage_id` int null references `task__stage`(`id`);
