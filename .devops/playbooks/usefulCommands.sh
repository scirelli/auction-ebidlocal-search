ansible-playbook installPyenv.yml --syntax-check
ansible-playbook --inventory inventory.ini installPyenv.yml --syntax-check
ansible-playbook --inventory inventory.ini installPyenv.yml --ask-become-pass
