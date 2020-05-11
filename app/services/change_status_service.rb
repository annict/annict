# frozen_string_literal: true

class ChangeStatusService
  def initialize(user:, work:)
    @user = user
    @work = work
  end

  def call(status_kind:)
    return unless status_kind

    status_kind = status_kind.to_sym
    library_entry = user.library_entries.find_by(work: work)

    return reset_status!(library_entry) if status_kind == :no_select

    status = user.statuses.new(work: work, kind: status_kind)

    ActiveRecord::Base.transaction do
      status.activity = user.build_or_last_activity(status, work, :create_status)
      persisted_activity = status.activity.persisted?

      status.save!
      status.save_library_entry
      status.update_channel_work

      if persisted_activity
        user.create_repetitive_activity!(status, work, :create_status)
      end

      prev_state_kind = library_entry&.status&.kind
      user.update_works_count!(prev_state_kind, status_kind)
      work.update_watchers_count!(prev_state_kind, status_kind)

      UserWatchedWorksCountJob.perform_later(user.id)
      status.share_to_sns
    end
  end

  private

  attr_reader :user, :work

  def reset_status!(library_entry)
    return unless library_entry

    ActiveRecord::Base.transaction do
      library_entry.update!(status_id: nil)
      UserWatchedWorksCountJob.perform_later(user.id)
    end
  end
end
