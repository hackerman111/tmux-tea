set -gF @tmux_tea_dir "#{d:current_file}"

bind-key t run-shell -b "#{@tmux_tea_dir}/scripts/tmux-tea confirm"
bind-key T display-popup -E -w 72 -h 28 "#{@tmux_tea_dir}/scripts/tmux-tea menu"

if -F '#{!=:#{@tmux_tea_status_loaded},1}' {
	set -gF status-right "#{status-right} ##(#{@tmux_tea_dir}/scripts/tmux-tea status)"
	set -g @tmux_tea_status_loaded 1
}
