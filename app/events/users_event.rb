class UsersEvent
  def self.properties(action, user)
    {
      action: action,
      user_id: user.id,
      keen: { timestamp: user.created_at }
    }
  end

  def self.publish(action, user)
    Keen.publish(:users, properties(action, user))
  end
end
