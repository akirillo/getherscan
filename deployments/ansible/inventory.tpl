%{ for inventory in inventories ~}
${trimspace(inventory)}

%{ endfor ~}
### All ###

[all:vars]
ansible_db_host="${db_host}"
ansible_db_port="${db_port}"
ansible_db_user="${db_user}"
ansible_db_password="${db_password}"
ansible_db_name="${db_name}"
