alter table `task__output` add `stage_id` int null references `task__stage`(`id`);

drop index if exists task__stage__start_output_id;
drop index if exists task__stage__end_output_id;

alter table `task__stage` drop column
    `start_output_id`;
alter table `task__stage` drop column
    `end_output_id`;