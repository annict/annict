class FirstStatusesEvent < StatusesEvent
  def self.publish(action, status)
    Keen.publish(:first_statuses, properties(action, status))
  end
end
