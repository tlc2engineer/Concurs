select id,sname,fname,email,city,country,t4.ms from  accounts inner join 
(select max(s) as ms,pid from
 (select likes.id,likes.pid,s from likes inner join  
(select id,similarity(1441,id) as s 
from (select distinct id from likes where pid in 
(select pid from likes where id=1441) and not id=1441) as t
---where id  in (select id from accounts where city='Варанск')
order by s desc) as t1
on likes.id=t1.id
where pid not in (select pid from likes where id=1441))as t3
group by pid 
)as t4
on accounts.id=t4.pid
order by ms desc,id desc limit 18