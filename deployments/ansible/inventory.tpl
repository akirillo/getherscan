%{ for inventory in inventories ~}
${trimspace(inventory)}

%{ endfor ~}
### All ###

[all:vars]
ansible_db_connection_string="${db_connection_string}"
