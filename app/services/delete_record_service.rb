# frozen_string_literal: true

class DeleteRecordService < ApplicationService
  class DeleteRecordServiceResult < Result
    attr_accessor :user
  end

  def initialize(record:)
    super()
    @record = record
  end

  def call
    user = @record.user
    library_entry = user.library_entries.find_by(work_id: @record.work_id)

    ActiveRecord::Base.transaction do
      @record.destroy!
      user.touch(:record_cache_expired_at)

      if library_entry && @record.episode_record?
        watched_episode_ids = library_entry.watched_episode_ids - [@record.episode_record.episode_id]
        library_entry.update_column(:watched_episode_ids, watched_episode_ids)
      end
    end

    @result.user = user
    @result
  end

  private

  def result_class
    DeleteRecordServiceResult
  end
end
