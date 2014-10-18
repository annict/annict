class FollowsEvent
  def self.properties(action, follow)
    {
      action: action,
      user_id: follow.user_id,
      following_id: follow.following_id,
      keen: { timestamp: follow.created_at }
    }
  end

  def self.publish(action, follow)
    Keen.publish(:follows, properties(action, follow))
  end
end
