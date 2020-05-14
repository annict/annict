# frozen_string_literal: true

class NewEpisodeRecordService
  attr_writer :app, :via, :ga_client, :page_category
  attr_reader :episode_record

  def initialize(user, episode_record)
    @user = user
    @episode_record = episode_record
    @work = @episode_record.work
  end

  def save!
    @episode_record.user = @user
    @episode_record.work = @work
    @episode_record.detect_locale!(:body)

    ActiveRecord::Base.transaction do
      @episode_record.record = @user.records.create!(work: @work)
      @episode_record.save!
      @episode_record.update_share_record_status
      update_library_entry
    end

    @episode_record.share_to_sns
    save_activity
    finish_tips
    create_ga_event

    true
  end

  private

  def save_activity
    return if @episode_record.multiple_episode_record.present?
    CreateEpisodeRecordActivityJob.perform_later(@user.id, @episode_record.id)
  end

  def finish_tips
    FinishUserTipsJob.perform_later(@user, "record") if @user.episode_records.initial?(@episode_record)
  end

  def update_library_entry
    library_entry = @user.library_entries.find_by(work: @episode_record.work)
    library_entry.append_episode!(@episode_record.episode) if library_entry.present?
  end

  def create_ga_event
    return if @ga_client.blank?
    @ga_client.events.create(:records, :create, el: "Episode", ds: @via)
  end
end
