class FirstCheckinsEvent < CheckinsEvent
  def self.publish(action, checkin)
    Keen.publish(:first_checkins, properties(action, checkin))
  end
end
