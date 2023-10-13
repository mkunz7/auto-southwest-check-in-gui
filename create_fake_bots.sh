tmux neww -t cgui -n AAA123
tmux neww -t cgui -n BBB123
tmux neww -t cgui -n CCC123
tmux neww -t cgui -n DDD123
sleep 1
tmux send-keys -t cgui:AAA123 "C-m"
tmux send-keys -t cgui:BBB123 "C-m"
tmux send-keys -t cgui:CCC123 "C-m"
tmux send-keys -t cgui:DDD123 "C-m"
