package dht

type Leader struct {
	Follower
}

func (leader *Leader) Handle(m *Message) (err error) {
	err = leader.Follower.Handle(m)
	if err != nil {
		return err
	}

	return nil
}
