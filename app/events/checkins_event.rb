class CheckinsEvent
  def self.properties(action, checkin)
    {
      action: action,
      user_id: checkin.user_id,
      work_id: checkin.episode.work_id,
      episode_id: checkin.episode_id,
      has_comment: checkin.comment.present?,
      shared_sns: checkin.shared_sns?,
      keen: { timestamp: checkin.created_at }
    }
  end

  def self.publish(action, checkin)
    Keen.publish(:checkins, properties(action, checkin))
  end
end
