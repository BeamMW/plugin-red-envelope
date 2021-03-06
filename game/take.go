package game

func (game *Game) Take(uid UID) (uint64, error) {
	game.lock()
	defer game.unlock()

	var amount uint64
	var err error

	if amount, err = game.envelope.take(uid); err != nil {
		return 0, err
	}
	game.updateUser(uid, func(user *User) bool {
		user.Taken += amount
		return true
	})

	return amount, err
}
