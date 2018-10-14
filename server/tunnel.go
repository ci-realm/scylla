package server

import (
	"github.com/jackc/pgx"
	"github.com/rgzr/sshtun"
)

const localDBPort = 7777

func setupDatabaseTunnel(started chan bool) {
	pgxcfg, err := pgx.ParseURI(config.DatabaseURL)
	if err != nil {
		logger.Fatalln(err)
	}

	tun := sshtun.New(localDBPort, "3.120.166.103", int(pgxcfg.Port))
	tun.SetDebug(true)
	tun.SetPort(443)
	tun.SetKeyFile(config.PrivateSSHKeyPath)
	tun.SetRemoteHost(pgxcfg.Host)
	tun.SetConnState(func(t *sshtun.SSHTun, state sshtun.ConnState) {
		switch state {
		case sshtun.StateStarting:
			logger.Println("SSH DB tunnel starting")
		case sshtun.StateStarted:
			logger.Println("SSH DB tunnel started")
			started <- true
		case sshtun.StateStopped:
			logger.Println("SSH DB tunnel stopped")
		}
	})

	err = tun.Start()
	if err != nil {
		logger.Fatalln("failed establishing SSH DB tunnel:", err)
	}
}
