#TODO: use pkgs.writeShellScript
echo "Setting up PostgreSQL"
alias pg_start="pg_ctl -D $PGDATA -l $PGDATA/logfile start"
alias pg_stop="pg_ctl -D $PGDATA stop"

pg_setup() {
	pg_stop;
	rm -rf $PG;
	initdb -D $PGDATA &&
	echo "unix_socket_directories = '$PGDATA'" >> $PGDATA/postgresql.conf &&
	pg_start &&
	createdb
}
