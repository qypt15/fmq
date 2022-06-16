package broker

import "github.com/eclipse/paho.mqtt.golang/packets"

func (b *Broker) getSession(cli *client, req *packets.ConnectPacket, resp *packets.ConnackPacket) error {

	var err error
	if len(req.ClientIdentifier) == 0{
		req.CleanSession = true

	}
	cid := req.ClientIdentifier

	// If CleanSession is NOT set, check the sessions store for existing sessions.
	// If found, return it.
	if !req.CleanSession {
		if cli.session, err = b.sessionMgr.Get(cid); err == nil {
			resp.SessionPresent = true

			if err := cli.session.Update(req); err != nil {
				return err
			}
		}
	}

	// If CleanSession, or no existing sessions found, then create a new one
	if cli.session == nil {
		if cli.session, err = b.sessionMgr.New(cid); err != nil {
			return err
		}

		resp.SessionPresent = false

		if err := cli.session.Init(req); err != nil {
			return err
		}
	}

	return nil
}