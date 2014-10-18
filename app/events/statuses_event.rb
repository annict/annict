class StatusesEvent
  def self.properties(action, status)
    {
      action: action,
      user_id: status.user_id,
      work_id: status.work_id,
      kind: status.kind,
      keen: { timestamp: status.created_at }
    }
  end

  def self.publish(action, status)
    Keen.publish(:statuses, properties(action, status))
  end
end
